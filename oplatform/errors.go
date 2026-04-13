package oplatform

import (
	"errors"
	"fmt"
)

// WeixinError 微信业务错误 (errcode != 0).
type WeixinError struct {
	ErrCode int
	ErrMsg  string
}

func (e *WeixinError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("oplatform: errcode=%d errmsg=%s", e.ErrCode, e.ErrMsg)
}

// Code returns the numeric errcode. Implements utils.WechatAPIError.
func (e *WeixinError) Code() int { return e.ErrCode }

// Message returns the human-readable errmsg. Implements utils.WechatAPIError.
func (e *WeixinError) Message() string { return e.ErrMsg }

// 常见哨兵错误。
var (
	// ErrNotFound 由 Store 实现返回，表示 key 不存在（非 I/O 错误）。
	ErrNotFound = errors.New("oplatform: not found")

	// ErrAuthorizerRevoked 当 refresh_token 失效 (errcode=61023) 时返回；
	// 调用方应删除 Store 中该 authorizer 的记录并重新引导授权。
	ErrAuthorizerRevoked = errors.New("oplatform: authorizer refresh_token revoked")

	// ErrVerifyTicketMissing 当 component_access_token 需要刷新但 Store
	// 尚未收到微信推送的 component_verify_ticket 时返回。
	ErrVerifyTicketMissing = errors.New("oplatform: component_verify_ticket not yet received")
)
