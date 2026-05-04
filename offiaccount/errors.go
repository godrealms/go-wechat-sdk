package offiaccount

import (
	"errors"
	"fmt"

	"github.com/godrealms/go-wechat-sdk/utils"
)

// Compile-time check: *WeixinError must implement utils.WechatAPIError.
var _ utils.WechatAPIError = (*WeixinError)(nil)

// WeixinError wraps a non-zero WeChat errcode returned by any API endpoint.
// It implements the error interface and can be inspected with errors.As.
type WeixinError struct {
	ErrCode int
	ErrMsg  string
}

// Error implements the error interface.
func (e *WeixinError) Error() string {
	return fmt.Sprintf("offiaccount: errcode=%d errmsg=%s", e.ErrCode, e.ErrMsg)
}

// Code returns the numeric errcode. Implements utils.WechatAPIError.
func (e *WeixinError) Code() int { return e.ErrCode }

// Message returns the human-readable errmsg. Implements utils.WechatAPIError.
func (e *WeixinError) Message() string { return e.ErrMsg }

// CheckResp returns a *WeixinError if r.ErrCode is non-zero, otherwise nil.
// Callers use it to translate a decoded Resp into a Go error.
func CheckResp(r *Resp) error {
	if r.ErrCode == 0 {
		return nil
	}
	return &WeixinError{ErrCode: r.ErrCode, ErrMsg: r.ErrMsg}
}

// IsTokenExpired reports whether err carries a WeChat errcode that indicates
// the access_token used in the request is no longer valid. WeChat returns:
//
//   - 40001: access_token invalid (most common — admin reset, parallel refresh)
//   - 40014: access_token format error (rare — corrupted token)
//   - 42001: access_token has expired (early-expiry edge case)
//   - 42007: ticket / token has expired
//
// When this returns true the caller should call Client.Invalidate() and retry.
// doGet and doPost do this automatically; callers using c.Https directly may
// invoke this helper themselves.
func IsTokenExpired(err error) bool {
	var werr *WeixinError
	if !errors.As(err, &werr) {
		return false
	}
	switch werr.ErrCode {
	case 40001, 40014, 42001, 42007:
		return true
	}
	return false
}
