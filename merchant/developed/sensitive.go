package pay

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
)

// PlatformCertForEncrypt 返回一个可用于敏感字段加密的平台证书及其序列号。
//
// 如果本地缓存为空，会先主动拉取一次 /v3/certificates 并解密缓存下来。
// 调用方拿到 (*x509.Certificate, serial) 后：
//  1. 用 EncryptSensitiveField 把身份证号、手机号、银行卡号、姓名等字段逐一
//     加密成 base64 密文填入请求体；
//  2. 把 serial 作为 "Wechatpay-Serial" 请求头传给 DoV3/DoV3WithHeaders，
//     告知服务端本次敏感数据是用哪一张平台证书加密的。
//
// 注意：缓存中可能存在多张证书（证书轮换期间），本方法并不保证总是返回
// 最"新"的一张——任意一张仍然生效的平台证书都可以用于加密，服务端通过
// Wechatpay-Serial 头挑选对应私钥解密即可。
func (c *Client) PlatformCertForEncrypt(ctx context.Context) (*x509.Certificate, string, error) {
	if cert, serial := c.anyPlatformCert(); cert != nil {
		return cert, serial, nil
	}
	if _, err := c.FetchPlatformCertificates(ctx); err != nil {
		return nil, "", fmt.Errorf("fetch platform cert: %w", err)
	}
	cert, serial := c.anyPlatformCert()
	if cert == nil {
		return nil, "", errors.New("pay: no platform certificate available after fetch")
	}
	return cert, serial, nil
}

// anyPlatformCert 从缓存里挑出一张证书。选取规则：当前实现按 map 迭代顺序
// 返回任意一张（Go 的 map 迭代是无序的，但在"有一张就够"的场景里已经足够）。
// 如果将来证书轮换策略变复杂，可以改为按 NotAfter 挑最晚过期的那张。
func (c *Client) anyPlatformCert() (*x509.Certificate, string) {
	c.platformCertsMu.RLock()
	defer c.platformCertsMu.RUnlock()
	for serial, cert := range c.platformCerts {
		return cert, serial
	}
	return nil, ""
}

// EncryptSensitiveField 用 RSA-OAEP (SHA-256) 把 plaintext 加密成 base64 密文，
// 适用于微信支付各类敏感字段（进件、分账收款人姓名、退款用户姓名等）。
//
// 传入的 cert 通常来自 Client.PlatformCertForEncrypt 返回值，
// 对应的序列号必须通过 Wechatpay-Serial 头一并上送。
func EncryptSensitiveField(cert *x509.Certificate, plaintext string) (string, error) {
	if cert == nil {
		return "", errors.New("pay: cert is nil")
	}
	pub, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return "", errors.New("pay: platform cert is not RSA")
	}
	cipher, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pub, []byte(plaintext), nil)
	if err != nil {
		return "", fmt.Errorf("pay: rsa oaep encrypt: %w", err)
	}
	return base64.StdEncoding.EncodeToString(cipher), nil
}
