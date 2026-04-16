package mini_store

import (
	"fmt"

	"github.com/godrealms/go-wechat-sdk/utils"
)

// Compile-time check: *APIError must implement utils.WechatAPIError.
var _ utils.WechatAPIError = (*APIError)(nil)

// APIError represents a WeChat Mini Store API business error. Callers can use
// errors.As to distinguish API errcode failures from network/transport errors.
type APIError struct {
	ErrCode int
	ErrMsg  string
	Path    string // the API path that triggered the error
}

func (e *APIError) Error() string {
	return fmt.Sprintf("mini_store: %s errcode=%d errmsg=%s", e.Path, e.ErrCode, e.ErrMsg)
}

// Code returns the numeric errcode. Implements utils.WechatAPIError.
func (e *APIError) Code() int { return e.ErrCode }

// Message returns the human-readable errmsg. Implements utils.WechatAPIError.
func (e *APIError) Message() string { return e.ErrMsg }
