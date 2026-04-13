package offiaccount

import "fmt"

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
