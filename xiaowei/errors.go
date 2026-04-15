package xiaowei

import (
	"fmt"

	"github.com/godrealms/go-wechat-sdk/utils"
)

// APIError represents a Xiaowei API business error (errcode != 0). Callers
// can use errors.As to distinguish API errcode failures from network or
// transport errors.
type APIError struct {
	ErrCode int
	ErrMsg  string
	Path    string // the API path that triggered the error
}

func (e *APIError) Error() string {
	return fmt.Sprintf("xiaowei: %s errcode=%d errmsg=%s", e.Path, e.ErrCode, e.ErrMsg)
}

// Code returns the numeric errcode. Implements utils.WechatAPIError.
func (e *APIError) Code() int { return e.ErrCode }

// Message returns the human-readable errmsg. Implements utils.WechatAPIError.
func (e *APIError) Message() string { return e.ErrMsg }

// Compile-time assertion that *APIError satisfies utils.WechatAPIError.
var _ utils.WechatAPIError = (*APIError)(nil)
