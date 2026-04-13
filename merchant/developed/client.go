// Package pay provides a client for the WeChat Pay (微信支付) v3 API.
// Use NewClient to create a client from a Config, then call API methods.
// All requests are signed with RSA-SHA256 per the WECHATPAY2-SHA256-RSA2048 scheme.
//
// 使用示例：
//
//	client, err := pay.NewClient(pay.Config{
//	    Appid:             "wx1234567890",
//	    Mchid:             "1900000001",
//	    CertificateNumber: "5157F09EFDC096DE15EBE81A47057A7232F1B8E1",
//	    APIv3Key:          "your_apiv3_key_32_bytes_long_xxx",
//	    PrivateKey:        privateKey, // *rsa.PrivateKey
//	})
//	if err != nil { ... }
//	resp, err := client.TransactionsJsapi(ctx, order)
//
// 客户端方法是并发安全的：每次请求构造独立的 Authorization 头，不会污染共享状态。
package pay

import (
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"fmt"
	"sync"

	"github.com/godrealms/go-wechat-sdk/utils"
)

// Config holds the configuration for initializing a Client. All fields are required except HTTP and Certificate.
type Config struct {
	Appid             string            // 应用ID（公众号/小程序/APP）
	Mchid             string            // 商户号
	CertificateNumber string            // 商户证书序列号
	APIv3Key          string            // 商户 APIv3 密钥
	PrivateKey        *rsa.PrivateKey   // 商户 API 私钥
	Certificate       *x509.Certificate // 商户证书（可选，用于本地校验）
	HTTP              *utils.HTTP       // 可选：注入自定义的 HTTP 客户端（含 logger 等）
}

// Client is the WeChat Pay v3 API client. Build one with NewClient or NewWechatClient+With* options. Safe for concurrent use after configuration.
type Client struct {
	appid             string
	mchid             string
	certificateNumber string
	apiV3Key          string
	privateKey        *rsa.PrivateKey
	certificate       *x509.Certificate
	http              *utils.HTTP

	platformCertsMu sync.RWMutex
	platformCerts   map[string]*x509.Certificate // serialNumber -> cert
}

// NewClient creates a new Client from cfg. Returns an error if any required field is missing.
// If cfg.HTTP is nil a default client rooted at the production WeChat Pay base URL is created.
func NewClient(cfg Config) (*Client, error) {
	if cfg.Appid == "" {
		return nil, errors.New("pay: Appid is required")
	}
	if cfg.Mchid == "" {
		return nil, errors.New("pay: Mchid is required")
	}
	if cfg.CertificateNumber == "" {
		return nil, errors.New("pay: CertificateNumber is required")
	}
	if cfg.APIv3Key == "" {
		return nil, errors.New("pay: APIv3Key is required")
	}
	if cfg.PrivateKey == nil {
		return nil, errors.New("pay: PrivateKey is required")
	}

	httpClient := cfg.HTTP
	if httpClient == nil {
		httpClient = utils.NewHTTP("https://api.mch.weixin.qq.com")
	}

	return &Client{
		appid:             cfg.Appid,
		mchid:             cfg.Mchid,
		certificateNumber: cfg.CertificateNumber,
		apiV3Key:          cfg.APIv3Key,
		privateKey:        cfg.PrivateKey,
		certificate:       cfg.Certificate,
		http:              httpClient,
		platformCerts:     make(map[string]*x509.Certificate),
	}, nil
}

// Appid returns the configured WeChat AppID.
func (c *Client) Appid() string { return c.appid }

// Mchid returns the configured merchant ID (商户号).
func (c *Client) Mchid() string { return c.mchid }

// HTTP returns the underlying HTTP client for advanced use.
func (c *Client) HTTP() *utils.HTTP { return c.http }

// PrivateKeyVal returns the RSA private key used to sign requests.
func (c *Client) PrivateKeyVal() *rsa.PrivateKey { return c.privateKey }

// CertificateVal returns the merchant's local certificate.
func (c *Client) CertificateVal() *x509.Certificate { return c.certificate }

// CertificateNumber returns the configured certificate serial number.
func (c *Client) CertificateNumber() string { return c.certificateNumber }

// ===== 兼容旧 API（已弃用，建议使用 NewClient） =====

// NewWechatClient returns a Client pre-configured with the production WeChat Pay base URL.
//
// Deprecated: Use NewClient(Config{...}) instead. The new constructor validates all required
// fields at creation time, surfacing errors earlier.
func NewWechatClient() *Client {
	return &Client{
		http:          utils.NewHTTP("https://api.mch.weixin.qq.com"),
		platformCerts: make(map[string]*x509.Certificate),
	}
}

// WithAppid sets the WeChat AppID (公众号/小程序 appid) associated with this merchant account.
//
// Deprecated: 使用 NewClient(Config{...})。
func (c *Client) WithAppid(v string) *Client { c.appid = v; return c }

// WithMchid sets the merchant ID (商户号) for this client.
//
// Deprecated: 使用 NewClient(Config{...})。
func (c *Client) WithMchid(v string) *Client { c.mchid = v; return c }

// WithCertificateNumber sets the API certificate serial number sent in the Authorization header.
//
// Deprecated: 使用 NewClient(Config{...})。
func (c *Client) WithCertificateNumber(v string) *Client { c.certificateNumber = v; return c }

// WithAPIv3Key sets the APIv3 key used for AES-GCM decryption of WeChat Pay notification payloads.
//
// Deprecated: 使用 NewClient(Config{...})。
func (c *Client) WithAPIv3Key(v string) *Client { c.apiV3Key = v; return c }

// WithCertificate sets the merchant certificate.
//
// Deprecated: 使用 NewClient(Config{...})。
func (c *Client) WithCertificate(v *x509.Certificate) *Client { c.certificate = v; return c }

// WithPrivateKey sets the RSA private key used to sign every API request.
//
// Deprecated: 使用 NewClient(Config{...})。
func (c *Client) WithPrivateKey(v *rsa.PrivateKey) *Client { c.privateKey = v; return c }

// WithPublicKey is a no-op kept for backward-compatibility.
//
// Deprecated: 微信支付 v3 不需要单独的公钥；保留仅为编译兼容。
func (c *Client) WithPublicKey(_ *rsa.PublicKey) *Client { return c }

// WithHttp replaces the underlying HTTP transport. Useful for injecting a test mock or a custom base URL.
//
// Deprecated: 使用 NewClient(Config{HTTP: ...})。
func (c *Client) WithHttp(h *utils.HTTP) *Client { c.http = h; return c }

// validateForRequest confirms that all required fields are set before a request.
func (c *Client) validateForRequest() error {
	if c.privateKey == nil {
		return errors.New("pay: PrivateKey is not set")
	}
	if c.mchid == "" {
		return errors.New("pay: Mchid is not set")
	}
	if c.certificateNumber == "" {
		return errors.New("pay: CertificateNumber is not set")
	}
	if c.http == nil {
		return errors.New("pay: HTTP client is not set")
	}
	return nil
}

// authorizationHeader builds the WECHATPAY2-SHA256-RSA2048 Authorization header value.
func (c *Client) authorizationHeader(method, urlPath, body, nonce string, timestamp int64) (string, error) {
	msg := fmt.Sprintf("%s\n%s\n%d\n%s\n%s\n", method, urlPath, timestamp, nonce, body)
	signature, err := utils.SignSHA256WithRSA(msg, c.privateKey)
	if err != nil {
		return "", fmt.Errorf("sign request failed: %w", err)
	}
	return fmt.Sprintf(
		`WECHATPAY2-SHA256-RSA2048 mchid="%s",nonce_str="%s",signature="%s",timestamp="%d",serial_no="%s"`,
		c.mchid, nonce, signature, timestamp, c.certificateNumber,
	), nil
}
