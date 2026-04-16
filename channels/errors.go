package channels

import (
	"fmt"

	"github.com/godrealms/go-wechat-sdk/utils"
)

// Compile-time check: *APIError must implement utils.WechatAPIError.
var _ utils.WechatAPIError = (*APIError)(nil)

// APIError wraps a WeChat Channels API error with the request path, errcode, and errmsg.
type APIError struct {
	ErrCode int
	ErrMsg  string
	Path    string // the API path that triggered the error
}

func (e *APIError) Error() string {
	return fmt.Sprintf("channels: %s errcode=%d errmsg=%s", e.Path, e.ErrCode, e.ErrMsg)
}

// Code returns the numeric errcode. Implements utils.WechatAPIError.
func (e *APIError) Code() int { return e.ErrCode }

// Message returns the human-readable errmsg. Implements utils.WechatAPIError.
func (e *APIError) Message() string { return e.ErrMsg }
