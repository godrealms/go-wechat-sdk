package offiaccount

import "fmt"

// WeixinError represents a WeChat API business error (errcode != 0).
// It implements utils.WechatAPIError for generic error inspection.
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

// CheckResp returns a *WeixinError if r.ErrCode != 0, otherwise nil.
// Use this after every WeChat API call to normalise error handling.
func CheckResp(r *Resp) error {
	if r.ErrCode == 0 {
		return nil
	}
	return &WeixinError{ErrCode: r.ErrCode, ErrMsg: r.ErrMsg}
}
