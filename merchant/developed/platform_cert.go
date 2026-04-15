package pay

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

// PlatformCertificate 描述微信支付平台证书的元信息。
type PlatformCertificate struct {
	SerialNo    string             `json:"serial_no"`
	EffectiveAt time.Time          `json:"effective_time"`
	ExpireAt    time.Time          `json:"expire_time"`
	Cert        *x509.Certificate  `json:"-"`
	Encrypt     EncryptCertificate `json:"encrypt_certificate"`
}

// EncryptCertificate 是 /v3/certificates 接口返回的加密结构。
type EncryptCertificate struct {
	Algorithm      string `json:"algorithm"`
	Nonce          string `json:"nonce"`
	AssociatedData string `json:"associated_data"`
	Ciphertext     string `json:"ciphertext"`
}

type certListResp struct {
	Data []PlatformCertificate `json:"data"`
}

// FetchPlatformCertificates 主动从微信支付拉取平台证书并缓存到 Client 内。
// 该方法自行处理签名（避免与 verifyResponseSignature 形成递归）。
func (c *Client) FetchPlatformCertificates(ctx context.Context) ([]*x509.Certificate, error) {
	if err := c.validateForRequest(); err != nil {
		return nil, err
	}

	const urlPath = "/v3/certificates"
	nonce := utils.GenerateNonceString(32)
	ts := time.Now().Unix()
	auth, err := c.authorizationHeader(http.MethodGet, urlPath, "", nonce, ts)
	if err != nil {
		return nil, err
	}

	headers := http.Header{
		"Accept":        []string{"application/json"},
		"Authorization": []string{auth},
		"User-Agent":    []string{"go-wechat-sdk/1.x"},
	}

	_, _, body, err := c.http.DoRequestWithRawResponse(ctx, http.MethodGet, urlPath, nil, nil, headers)
	if err != nil {
		return nil, err
	}

	var resp certListResp
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode certificates failed: %w: %s", err, string(body))
	}

	// 解密 / 解析放在锁外，仅在最终写缓存时短暂持锁，避免长时间阻塞读取者。
	parsed := make(map[string]*x509.Certificate, len(resp.Data))
	out := make([]*x509.Certificate, 0, len(resp.Data))
	for i := range resp.Data {
		certPem, err := decryptAES256GCM(c.apiV3Key,
			resp.Data[i].Encrypt.Nonce,
			resp.Data[i].Encrypt.AssociatedData,
			resp.Data[i].Encrypt.Ciphertext)
		if err != nil {
			return nil, fmt.Errorf("decrypt platform cert failed: %w", err)
		}
		cert, err := utils.LoadCertificate(string(certPem))
		if err != nil {
			return nil, fmt.Errorf("parse platform cert failed: %w", err)
		}
		serial := utils.GetCertificateSerialNumber(*cert)
		parsed[serial] = cert
		out = append(out, cert)
	}

	c.platformCertsMu.Lock()
	for serial, cert := range parsed {
		c.platformCerts[serial] = cert
	}
	c.platformCertsMu.Unlock()
	return out, nil
}

// AddPlatformCertificate 把已知的平台证书写入缓存（用于测试或离线引导）。
func (c *Client) AddPlatformCertificate(cert *x509.Certificate) {
	if cert == nil {
		return
	}
	c.platformCertsMu.Lock()
	defer c.platformCertsMu.Unlock()
	c.platformCerts[utils.GetCertificateSerialNumber(*cert)] = cert
}

// platformCertBySerial 返回已缓存的某序列号对应的平台证书。
func (c *Client) platformCertBySerial(serial string) *x509.Certificate {
	c.platformCertsMu.RLock()
	defer c.platformCertsMu.RUnlock()
	return c.platformCerts[serial]
}

// verifyResponseSignature 校验微信支付返回的响应签名。
//
// 流程参考微信支付 V3 文档 https://pay.weixin.qq.com/doc/v3/merchant/4012365334
//  1. 取响应头中的 Wechatpay-Timestamp / Wechatpay-Nonce / Wechatpay-Signature / Wechatpay-Serial；
//  2. 拼接 timestamp\nnonce\nbody\n 形成验签串；
//  3. 用对应序列号的平台证书公钥做 SHA256-RSA 验签。
//
// If any signature header is missing this returns an error — the previous
// "silent skip" behavior was a security hole and has been removed.
// Additionally enforces a ±5 minute window on Wechatpay-Timestamp to mitigate replay.
// 若本地没有对应序列号的证书，会自动调用 FetchPlatformCertificates 拉取一次再重试。
func (c *Client) verifyResponseSignature(ctx context.Context, header http.Header, body []byte) error {
	timestamp := header.Get("Wechatpay-Timestamp")
	nonce := header.Get("Wechatpay-Nonce")
	signature := header.Get("Wechatpay-Signature")
	serial := header.Get("Wechatpay-Serial")
	if timestamp == "" || nonce == "" || signature == "" || serial == "" {
		return fmt.Errorf("missing wechatpay signature header (ts=%q nonce=%q sig=%q serial=%q)",
			timestamp, nonce, signature, serial)
	}
	if err := checkWechatpayTimestamp(timestamp); err != nil {
		return err
	}

	cert := c.platformCertBySerial(serial)
	if cert == nil {
		if _, err := c.FetchPlatformCertificates(ctx); err != nil {
			return fmt.Errorf("fetch platform cert: %w", err)
		}
		cert = c.platformCertBySerial(serial)
		if cert == nil {
			return fmt.Errorf("no platform certificate for serial %s", serial)
		}
	}

	pub, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return errors.New("platform cert is not an RSA certificate")
	}

	source := fmt.Sprintf("%s\n%s\n%s\n", timestamp, nonce, string(body))
	return utils.VerifySHA256WithRSA(source, signature, pub)
}

// decryptAES256GCM 用 APIv3Key 解密 base64 密文。
func decryptAES256GCM(key, nonce, associatedData, ciphertext string) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, err
	}
	c, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}
	plain, err := gcm.Open(nil, []byte(nonce), decoded, []byte(associatedData))
	if err != nil {
		return nil, err
	}
	if len(plain) == 0 {
		return nil, errors.New("empty plaintext")
	}
	return plain, nil
}

// checkWechatpayTimestamp validates the Wechatpay-Timestamp header against the local clock.
// WeChat Pay v3 uses a ±5 minute replay window.
func checkWechatpayTimestamp(ts string) error {
	n, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid wechatpay timestamp %q: %w", ts, err)
	}
	delta := time.Now().Unix() - n
	if delta < 0 {
		delta = -delta
	}
	if delta > 300 {
		return fmt.Errorf("wechatpay timestamp out of window: delta=%ds", delta)
	}
	return nil
}
