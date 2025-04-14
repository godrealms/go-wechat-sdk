package wechat

import (
	"crypto/rsa"
	"crypto/x509"
	"github.com/godrealms/go-wechat-sdk/utils"
)

// Client 微信商户支付
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

func NewWechatClient() *Client {
	return &Client{
		Http: utils.NewHTTP("https://api.mch.weixin.qq.com"),
	}
}

func (c *Client) WithAppid(appid string) *Client {
	c.Appid = appid
	return c
}

func (c *Client) WithMchid(mchid string) *Client {
	c.Mchid = mchid
	return c
}

func (c *Client) WithCertificateNumber(certificateNumber string) *Client {
	c.CertificateNumber = certificateNumber
	return c
}

func (c *Client) WithAPIv3Key(APIv3Key string) *Client {
	c.APIv3Key = APIv3Key
	return c
}

func (c *Client) WithCertificate(certificate *x509.Certificate) *Client {
	c.certificate = certificate
	return c
}

func (c *Client) WithPrivateKey(privateKey *rsa.PrivateKey) *Client {
	c.privateKey = privateKey
	return c
}

func (c *Client) WithPublicKey(publicKey *rsa.PublicKey) *Client {
	c.publicKey = publicKey
	return c
}

func (c *Client) WithHttp(http *utils.HTTP) *Client {
	c.Http = http
	return c
}
