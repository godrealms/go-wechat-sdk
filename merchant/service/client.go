// Package service 提供微信支付「服务商模式」的薄封装。
//
// 服务商模式与直连商户模式在 V3 协议层面完全一致：都用服务商自己的
// 商户号（sp_mchid）+ 证书签名，区别仅在于下单等请求体中需要额外携带
// sub_mchid / sub_appid 字段。因此本包直接复用
// github.com/godrealms/go-wechat-sdk/merchant/developed 的 Client，
// 只是强制要求 sub_mchid 非空，并帮你在部分请求体里自动注入。
//
// 如果你自己的业务请求体已经包含 sub_mchid，可以直接用 developed.Client。
package service

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"fmt"

	pay "github.com/godrealms/go-wechat-sdk/merchant/developed"
)

// Config 服务商模式配置。SpMchid 是服务商自己的商户号；
// 其余证书、APIv3 密钥等均属于服务商本身。
type Config struct {
	SpMchid           string // 服务商商户号
	SpAppid           string // 服务商 AppId
	SubMchid          string // 默认的子商户号（可在单次调用时覆盖）
	SubAppid          string // 默认的子商户 AppId（可选）
	CertificateNumber string
	APIv3Key          string
	PrivateKey        *rsa.PrivateKey
	Certificate       *x509.Certificate
}

// Client 服务商模式客户端。内部持有一个 pay.Client 以复用所有签名/验签逻辑。
type Client struct {
	inner    *pay.Client
	subMchid string
	subAppid string
}

// NewClient 构造服务商客户端。
func NewClient(cfg Config) (*Client, error) {
	if cfg.SpMchid == "" || cfg.SpAppid == "" {
		return nil, fmt.Errorf("service: SpMchid and SpAppid are required")
	}
	if cfg.SubMchid == "" {
		return nil, fmt.Errorf("service: SubMchid is required")
	}
	inner, err := pay.NewClient(pay.Config{
		Appid:             cfg.SpAppid,
		Mchid:             cfg.SpMchid,
		CertificateNumber: cfg.CertificateNumber,
		APIv3Key:          cfg.APIv3Key,
		PrivateKey:        cfg.PrivateKey,
		Certificate:       cfg.Certificate,
	})
	if err != nil {
		return nil, err
	}
	return &Client{
		inner:    inner,
		subMchid: cfg.SubMchid,
		subAppid: cfg.SubAppid,
	}, nil
}

// Inner 返回内部 developed.Client，方便直接调用所有已有的 API。
func (c *Client) Inner() *pay.Client { return c.inner }

// SubMchid 返回默认子商户号。
func (c *Client) SubMchid() string { return c.subMchid }

// SubAppid 返回默认子商户 AppId。
func (c *Client) SubAppid() string { return c.subAppid }

// injectSubFields 在 body 为 map 时自动补上 sp_/sub_ 字段；
// 如果调用方自己已经传了，就尊重调用方。
//
// 该方法返回一个新的 map，不会修改调用方传入的 body，避免出现请求被
// "偷偷增字段" 的副作用，方便调用方在多个请求间复用同一个模板。
func (c *Client) injectSubFields(body map[string]any) map[string]any {
	out := make(map[string]any, len(body)+4)
	for k, v := range body {
		out[k] = v
	}
	if _, ok := out["sp_mchid"]; !ok {
		out["sp_mchid"] = c.inner.Mchid()
	}
	if _, ok := out["sp_appid"]; !ok {
		out["sp_appid"] = c.inner.Appid()
	}
	if _, ok := out["sub_mchid"]; !ok && c.subMchid != "" {
		out["sub_mchid"] = c.subMchid
	}
	if _, ok := out["sub_appid"]; !ok && c.subAppid != "" {
		out["sub_appid"] = c.subAppid
	}
	return out
}

// PartnerTransactionsJsapi 服务商模式 JSAPI 下单（薄封装）。
func (c *Client) PartnerTransactionsJsapi(ctx context.Context, body map[string]any) (map[string]any, error) {
	return c.postPartner(ctx, "/v3/pay/partner/transactions/jsapi", body)
}

// PartnerTransactionsApp 服务商模式 APP 下单。
func (c *Client) PartnerTransactionsApp(ctx context.Context, body map[string]any) (map[string]any, error) {
	return c.postPartner(ctx, "/v3/pay/partner/transactions/app", body)
}

// PartnerTransactionsH5 服务商模式 H5 下单。
func (c *Client) PartnerTransactionsH5(ctx context.Context, body map[string]any) (map[string]any, error) {
	return c.postPartner(ctx, "/v3/pay/partner/transactions/h5", body)
}

// PartnerTransactionsNative 服务商模式 Native 下单。
func (c *Client) PartnerTransactionsNative(ctx context.Context, body map[string]any) (map[string]any, error) {
	return c.postPartner(ctx, "/v3/pay/partner/transactions/native", body)
}

func (c *Client) postPartner(ctx context.Context, path string, body map[string]any) (map[string]any, error) {
	body = c.injectSubFields(body)
	result := map[string]any{}
	if err := c.inner.PostV3Raw(ctx, path, body, &result); err != nil {
		return nil, err
	}
	return result, nil
}
