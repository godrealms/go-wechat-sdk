package pay

import (
	"encoding/json"
	"fmt"

	"github.com/godrealms/go-wechat-sdk/merchant/developed/errorx"
	"github.com/godrealms/go-wechat-sdk/utils"
)

// Compile-time interface assertions.
var (
	_ utils.WechatAPIError   = (*APIError)(nil)
	_ utils.WechatAPIV3Error = (*V3Error)(nil)
)

// APIError represents a WeChat Pay merchant API business error.
type APIError struct {
	ErrCode int
	ErrMsg  string
	Path    string // the API path that triggered the error
}

func (e *APIError) Error() string {
	return fmt.Sprintf("merchant/developed: %s errcode=%d errmsg=%s", e.Path, e.ErrCode, e.ErrMsg)
}

// Code returns the numeric errcode. Implements utils.WechatAPIError.
func (e *APIError) Code() int { return e.ErrCode }

// Message returns the human-readable errmsg. Implements utils.WechatAPIError.
func (e *APIError) Message() string { return e.ErrMsg }

// V3Error represents a WeChat Pay v3 API error envelope. v3 uses STRING codes
// like "PARAM_ERROR" or "OUT_TRADE_NO_USED" — distinct from the legacy int-code
// APIError above (which is kept for backwards compat with existing callers).
//
// Implements utils.WechatAPIV3Error: callers should normally use the
// Code() / Message() / HTTPStatus() accessors via that interface rather than
// reaching for the underlying fields.
type V3Error struct {
	// Status is the HTTP status code returned by WeChat Pay V3.
	Status int
	// ErrCode is the V3 string error code, e.g. "PARAM_ERROR".
	ErrCode string
	// ErrMsg is the human-readable error message.
	ErrMsg string
	// Detail holds the optional structured detail object as raw JSON for
	// forward-compatibility; nil if the envelope had no "detail" key.
	Detail json.RawMessage
	// Path is the API path that triggered the error.
	Path string
}

func (e *V3Error) Error() string {
	return fmt.Sprintf("merchant/developed: %s status=%d code=%s message=%s",
		e.Path, e.Status, e.ErrCode, e.ErrMsg)
}

// Code returns the V3 string error code. Implements utils.WechatAPIV3Error.
func (e *V3Error) Code() string { return e.ErrCode }

// Message returns the human-readable error message. Implements utils.WechatAPIV3Error.
func (e *V3Error) Message() string { return e.ErrMsg }

// HTTPStatus returns the HTTP status code of the failed response.
// Implements utils.WechatAPIV3Error.
func (e *V3Error) HTTPStatus() int { return e.Status }

// Solution returns the WeChat-Pay-documented remediation hint for this code.
// Currently only consults the /pay/transactions/app errorx table — additional
// per-endpoint tables will be wired up as they are populated. Empty if unknown.
func (e *V3Error) Solution() string {
	if e == nil {
		return ""
	}
	if hint, ok := errorx.LookupTransactionsApp(e.ErrCode); ok {
		return hint
	}
	return ""
}
