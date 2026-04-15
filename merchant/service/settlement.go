package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

// 子商户结算账户查询 / 修改。
//
// 文档:
//   - 查询: https://pay.weixin.qq.com/docs/partner/apis/partner-applyment/sub-merchant-settlement/query-settlement.html
//   - 修改: https://pay.weixin.qq.com/docs/partner/apis/partner-applyment/sub-merchant-settlement/modify-settlement.html
//
// 要点：
//  1. account_number 是敏感字段；查询时平台返回密文，修改时也必须上送密文。
//     SDK 不会帮你解密查询结果（调用方自己用商户 API 证书私钥 RSA-OAEP 解密），
//     但修改时提供 SettlementModifyEncrypted 辅助方法，一次性完成
//     "取平台证书 -> 加密 account_number -> 写 Wechatpay-Serial 头 -> 发 PUT"
//     的流程。
//  2. 修改接口成功时返回 200 + 空响应体，SDK 会当作成功处理。
//  3. 查询 / 修改均作用于 **特约商户（子商户）**，请求路径里的 sub_mchid
//     就是要操作的子商户号；如果调用时传空字符串，SDK 会使用
//     NewClient 时配置的默认 SubMchid。

// SettlementInfo 子商户结算账户信息（查询返回值 / 修改入参的字段集并集）。
//
// 注意：AccountNumber 在查询响应中是 **平台证书加密后的密文**；调用方拿到
// 之后需要自己用商户 API 证书私钥做 RSA-OAEP(SHA256) 解密。SDK 不做自动
// 解密是为了避免把解密后的明文长期保留在进程内存中。
type SettlementInfo struct {
	// AccountType 账户类型，如 ACCOUNT_TYPE_BUSINESS / ACCOUNT_TYPE_PRIVATE。
	AccountType string `json:"account_type"`
	// AccountBank 开户银行（示例: "工商银行"）。
	AccountBank string `json:"account_bank"`
	// BankAddressCode 开户银行省市编码（国标行政区划代码）。
	BankAddressCode string `json:"bank_address_code,omitempty"`
	// BankBranchID 开户银行联行号（部分银行需要）。
	BankBranchID string `json:"bank_branch_id,omitempty"`
	// BankName 开户银行全称（含支行）。
	BankName string `json:"bank_name,omitempty"`
	// AccountNumber 银行账号。
	//
	// 查询返回：平台证书加密后的密文（需要商户侧私钥解密）。
	// 修改入参：商户侧用平台证书加密后的密文。
	AccountNumber string `json:"account_number"`
	// VerifyResult 最近一次打款验证结果，仅查询返回。
	VerifyResult string `json:"verify_result,omitempty"`
	// VerifyFailReason 打款验证失败原因，仅在 verify_result 为失败态时出现。
	VerifyFailReason string `json:"verify_fail_reason,omitempty"`
}

// SettlementModifyRequest 修改结算账户的请求体。
type SettlementModifyRequest struct {
	// ModifyBalance 是否同时修改出款账户。false 为仅修改结算账户。
	ModifyBalance bool `json:"modify_balance"`
	// AccountType 账户类型（必填）。
	AccountType string `json:"account_type"`
	// AccountBank 开户银行（必填）。
	AccountBank string `json:"account_bank"`
	// BankAddressCode 开户银行省市编码（小微商户必填）。
	BankAddressCode string `json:"bank_address_code,omitempty"`
	// BankName 开户银行全称（非必填）。
	BankName string `json:"bank_name,omitempty"`
	// BankBranchID 开户银行联行号（部分场景必填）。
	BankBranchID string `json:"bank_branch_id,omitempty"`
	// AccountNumber 银行账号密文（必填）。调用方可以直接传入已经用
	// EncryptSensitive 加密过的密文，或改用
	// SettlementModifyEncrypted 让 SDK 自动完成加密。
	AccountNumber string `json:"account_number"`
}

// SettlementQuery 查询子商户结算账户。subMchid 为空时使用默认子商户号。
//
// 返回的 SettlementInfo.AccountNumber 是平台证书加密后的密文，请用商户
// API 证书私钥做 RSA-OAEP(SHA256) 解密后再使用（参见 utils/rsa_helper.go
// 或自行解密，SDK 故意不缓存解密后的明文）。
func (c *Client) SettlementQuery(ctx context.Context, subMchid string) (*SettlementInfo, error) {
	if subMchid == "" {
		subMchid = c.subMchid
	}
	if subMchid == "" {
		return nil, errors.New("service: subMchid is required")
	}
	path := fmt.Sprintf("/v3/applyment4sub/sub_merchants/%s/settlement", subMchid)
	var resp SettlementInfo
	if err := c.inner.DoV3(ctx, http.MethodGet, path, nil, nil, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SettlementModify 修改子商户结算账户。调用方需要自行对 req.AccountNumber
// 做平台证书加密，并把加密时所用证书的序列号通过 platformSerial 传入。
// 如果你希望 SDK 自动完成加密，请改用 SettlementModifyEncrypted。
//
// subMchid 为空时使用默认子商户号。成功时返回 nil（接口 200 无响应体）。
func (c *Client) SettlementModify(
	ctx context.Context,
	subMchid string,
	req *SettlementModifyRequest,
	platformSerial string,
) error {
	if req == nil {
		return errors.New("service: SettlementModifyRequest is required")
	}
	if req.AccountNumber == "" {
		return errors.New("service: account_number is required (must be encrypted ciphertext)")
	}
	if subMchid == "" {
		subMchid = c.subMchid
	}
	if subMchid == "" {
		return errors.New("service: subMchid is required")
	}
	var headers http.Header
	if platformSerial != "" {
		headers = http.Header{"Wechatpay-Serial": []string{platformSerial}}
	}
	path := fmt.Sprintf("/v3/applyment4sub/sub_merchants/%s/modify-settlement", subMchid)
	return c.inner.DoV3(ctx, http.MethodPut, path, nil, req, headers, nil)
}

// SettlementModifyEncrypted 是 SettlementModify 的高层封装：接受 **明文**
// 账号，内部调用 EncryptSensitive 加密并自动填充 Wechatpay-Serial 头。
//
// 如果 req.AccountNumber 已经是密文（调用方自己加密过），请直接用
// SettlementModify，以避免对密文再次加密。
func (c *Client) SettlementModifyEncrypted(
	ctx context.Context,
	subMchid string,
	req *SettlementModifyRequest,
	plaintextAccountNumber string,
) error {
	if req == nil {
		return errors.New("service: SettlementModifyRequest is required")
	}
	if plaintextAccountNumber == "" {
		return errors.New("service: plaintextAccountNumber is required")
	}
	cipher, serial, err := c.EncryptSensitive(ctx, plaintextAccountNumber)
	if err != nil {
		return fmt.Errorf("service: encrypt account_number: %w", err)
	}
	// 不修改调用方传入的 req，避免在其进程内保留密文状态。
	clone := *req
	clone.AccountNumber = cipher
	return c.SettlementModify(ctx, subMchid, &clone, serial)
}
