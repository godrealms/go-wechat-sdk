package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	pay "github.com/godrealms/go-wechat-sdk/merchant/developed"
)

// 子商户进件（特约商户进件 / applyment4sub）相关接口。
//
// 文档: https://pay.weixin.qq.com/docs/partner/apis/partner-applyment/applyments.html
//
// 所有"敏感字段"（法人/经办人姓名、身份证号、手机号、银行卡号等）都必须先用
// 平台证书做 RSA-OAEP(SHA256) 加密后再作为请求体字段上送；同时请求头
// "Wechatpay-Serial" 必须携带加密所用平台证书的序列号。SDK 通过
// EncryptSensitive 一次性返回 (ciphertext, platformSerial)，再把 platformSerial
// 传给 ApplymentSubmit 即可。

// ApplymentSubmitResponse 提交进件单成功的响应体。
type ApplymentSubmitResponse struct {
	// ApplymentID 由微信生成的申请单号，后续用于查询进度。
	ApplymentID int64 `json:"applyment_id"`
}

// ApplymentQueryResponse 进件单状态查询响应体。
// 字段为可选的原因：同一份结构体既要覆盖 business_code 查询路径，也要覆盖
// applyment_id 查询路径；不同状态下返回的字段集合不同。
type ApplymentQueryResponse struct {
	// BusinessCode 业务申请编号（提交时由商户自行生成）。
	BusinessCode string `json:"business_code,omitempty"`
	// ApplymentID 微信支付侧的申请单号。
	ApplymentID int64 `json:"applyment_id,omitempty"`
	// SubMchid 审核完成后返回的特约商户号。
	SubMchid string `json:"sub_mchid,omitempty"`
	// SignURL 审核通过后的超级管理员签约链接。
	SignURL string `json:"sign_url,omitempty"`
	// ApplymentState 申请单当前状态，例如 APPLYMENT_STATE_FINISHED。
	ApplymentState string `json:"applyment_state"`
	// ApplymentStateMsg 申请单状态描述。
	ApplymentStateMsg string `json:"applyment_state_msg"`
	// AuditDetail 若状态为待补充资料，返回需要补充/驳回的字段明细。
	AuditDetail []AuditDetail `json:"audit_detail,omitempty"`
}

// AuditDetail 描述一个被驳回的字段。
type AuditDetail struct {
	ParamName    string `json:"param_name"`
	RejectReason string `json:"reject_reason"`
}

// EncryptSensitive 使用当前缓存（或主动拉取）的平台证书把 plaintext 加密为
// base64 密文，并返回所用证书的序列号。调用方应当把序列号作为
// Wechatpay-Serial 头传给 ApplymentSubmit 或任何需要敏感字段加密的接口。
func (c *Client) EncryptSensitive(ctx context.Context, plaintext string) (ciphertext, platformSerial string, err error) {
	cert, serial, err := c.inner.PlatformCertForEncrypt(ctx)
	if err != nil {
		return "", "", err
	}
	ct, err := pay.EncryptSensitiveField(cert, plaintext)
	if err != nil {
		return "", "", err
	}
	return ct, serial, nil
}

// ApplymentSubmit 提交特约商户进件申请单。
//
// 参数：
//   - body：完整的申请单内容，字段结构参见微信支付官方文档。由于字段繁多、
//     不同主体类型差异较大，本 SDK 不做强类型封装；调用方可以使用
//     map[string]any 自由构造，或者在自己的代码里定义结构体再传入。
//   - platformSerial：敏感字段加密所用的平台证书序列号（即 EncryptSensitive
//     第二个返回值）。非空才会写入 Wechatpay-Serial 头；调用方如果确定当前
//     body 没有任何加密字段，可以传空字符串。
func (c *Client) ApplymentSubmit(ctx context.Context, body any, platformSerial string) (*ApplymentSubmitResponse, error) {
	if body == nil {
		return nil, errors.New("service: applyment body is required")
	}
	var headers http.Header
	if platformSerial != "" {
		headers = http.Header{"Wechatpay-Serial": []string{platformSerial}}
	}
	var resp ApplymentSubmitResponse
	if err := c.inner.DoV3(ctx, http.MethodPost, "/v3/applyment4sub/applyment/", nil, body, headers, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ApplymentQueryByBusinessCode 使用业务申请编号查询进件状态。
// 典型用法：提交后立即用自己生成的 business_code 轮询。
func (c *Client) ApplymentQueryByBusinessCode(ctx context.Context, businessCode string) (*ApplymentQueryResponse, error) {
	if businessCode == "" {
		return nil, errors.New("service: businessCode is required")
	}
	path := fmt.Sprintf("/v3/applyment4sub/applyment/business_code/%s", businessCode)
	var resp ApplymentQueryResponse
	if err := c.inner.DoV3(ctx, http.MethodGet, path, nil, nil, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ApplymentQueryByID 使用微信侧的 applyment_id 查询进件状态。
// 一般在首次通过 business_code 查询拿到 applyment_id 后，后续轮询改走这个。
func (c *Client) ApplymentQueryByID(ctx context.Context, applymentID int64) (*ApplymentQueryResponse, error) {
	if applymentID <= 0 {
		return nil, errors.New("service: applymentID must be > 0")
	}
	path := fmt.Sprintf("/v3/applyment4sub/applyment/applyment_id/%d", applymentID)
	var resp ApplymentQueryResponse
	if err := c.inner.DoV3(ctx, http.MethodGet, path, nil, nil, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
