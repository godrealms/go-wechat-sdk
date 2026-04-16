package isv

import (
	"errors"
	"fmt"

	"github.com/godrealms/go-wechat-sdk/utils"
)

// Compile-time check: *WeixinError must implement utils.WechatAPIError.
var _ utils.WechatAPIError = (*WeixinError)(nil)

// 哨兵错误,可用 errors.Is 判断。
var (
	ErrNotFound              = errors.New("isv: not found")
	ErrSuiteTicketMissing    = errors.New("isv: suite_ticket missing in store")
	ErrProviderCorpIDMissing = errors.New("isv: provider corpid not configured")
	ErrProviderSecretMissing = errors.New("isv: provider secret not configured")
	ErrAuthorizerRevoked     = errors.New("isv: authorizer revoked")
)

// WeixinError wraps a non-zero WeCom errcode returned by any API endpoint. It implements the utils.WechatAPIError interface.
type WeixinError struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (e *WeixinError) Error() string {
	return fmt.Sprintf("isv: weixin error %d: %s", e.ErrCode, e.ErrMsg)
}

// Code returns the numeric errcode. Implements utils.WechatAPIError.
func (e *WeixinError) Code() int { return e.ErrCode }

// Message returns the human-readable errmsg. Implements utils.WechatAPIError.
func (e *WeixinError) Message() string { return e.ErrMsg }
