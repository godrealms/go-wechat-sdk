// Package developed provides a client for the WeChat Pay (微信支付) v3 API.
// Use NewWechatClient to create a client, then call the builder methods
// (WithAppid, WithMchid, WithPrivateKey, etc.) before making API calls.
// All requests are signed with RSA-SHA256 per the WECHATPAY2-SHA256-RSA2048 scheme.
package developed

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"net/url"
	"github.com/godrealms/go-wechat-sdk/utils"
)

// Client is the WeChat Pay v3 API client. Build one with NewWechatClient and the With* options. Safe for concurrent use after configuration.
type Client struct {
	Appid             string            // 应用ID
	Mchid             string            // 商户号
	CertificateNumber string            // 商户证书序列号
	APIv3Key          string            // 商户APIv3密钥
	certificate       *x509.Certificate // 证书
	privateKey        *rsa.PrivateKey   // 私钥
	publicKey         *rsa.PublicKey    // 公钥
	Http              *utils.HTTP
}

// NewWechatClient returns a Client pre-configured with the production WeChat Pay base URL.
func NewWechatClient() *Client {
	return &Client{
		Http: utils.NewHTTP("https://api.mch.weixin.qq.com"),
	}
}

// WithAppid sets the WeChat AppID (公众号/小程序 appid) associated with this merchant account.
func (c *Client) WithAppid(appid string) *Client {
	c.Appid = appid
	return c
}

// WithMchid sets the merchant ID (商户号) for this client.
func (c *Client) WithMchid(mchid string) *Client {
	c.Mchid = mchid
	return c
}

// WithCertificateNumber sets the API certificate serial number sent in the Authorization header.
func (c *Client) WithCertificateNumber(certificateNumber string) *Client {
	c.CertificateNumber = certificateNumber
	return c
}

// WithAPIv3Key sets the APIv3 key used for AES-GCM decryption of WeChat Pay notification payloads.
func (c *Client) WithAPIv3Key(APIv3Key string) *Client {
	c.APIv3Key = APIv3Key
	return c
}

func (c *Client) WithCertificate(certificate *x509.Certificate) *Client {
	c.certificate = certificate
	return c
}

// WithPrivateKey sets the RSA private key used to sign every API request.
func (c *Client) WithPrivateKey(privateKey *rsa.PrivateKey) *Client {
	c.privateKey = privateKey
	return c
}

func (c *Client) WithPublicKey(publicKey *rsa.PublicKey) *Client {
	c.publicKey = publicKey
	return c
}

// WithHttp replaces the underlying HTTP transport. Useful for injecting a test mock or a custom base URL.
func (c *Client) WithHttp(http *utils.HTTP) *Client {
	c.Http = http
	return c
}

// doV3 is a helper method for making WeChat Pay API v3 requests.
// It's a stub implementation to allow the package to compile.
func (c *Client) doV3(ctx context.Context, method, path string, query url.Values, body, result any) error {
	return fmt.Errorf("doV3: not implemented")
}

// postV3 is a helper method for making POST requests to WeChat Pay API v3.
// It's a stub implementation to allow the package to compile.
func (c *Client) postV3(ctx context.Context, path string, body, result any) error {
	return fmt.Errorf("postV3: not implemented")
}

// getV3 is a helper method for making GET requests to WeChat Pay API v3.
// It's a stub implementation to allow the package to compile.
func (c *Client) getV3(ctx context.Context, path string, query url.Values, result any) error {
	return fmt.Errorf("getV3: not implemented")
}

// verifyResponseSignature is a helper method for verifying WeChat Pay response signatures.
// It's a stub implementation to allow the package to compile.
func (c *Client) verifyResponseSignature(ctx context.Context, header interface{}, body []byte) error {
	return fmt.Errorf("verifyResponseSignature: not implemented")
}

// apiV3Key is a getter for the API v3 key.
// It's a stub implementation to allow the package to compile.
func (c *Client) apiV3Key() string {
	return c.APIv3Key
}

// decryptAES256GCM is a package-level function for decrypting AES-256-GCM encrypted data.
// It's a stub implementation to allow the package to compile.
func decryptAES256GCM(apiV3Key, nonce, associatedData, ciphertext string) ([]byte, error) {
	return nil, fmt.Errorf("decryptAES256GCM: not implemented")
}
