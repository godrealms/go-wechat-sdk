package pay

import (
	"encoding/json"
	"fmt"

	"github.com/godrealms/go-wechat-sdk/merchant/developed/errorx"
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
type V3Error struct {
	HTTPStatus int             // HTTP status from the response
	Code       string          // e.g. "PARAM_ERROR"
	Message    string          // human-readable
	Detail     json.RawMessage // optional structured detail (kept raw for forward-compat); nil if envelope had no "detail" key
	Path       string          // API path that triggered the error
}

func (e *V3Error) Error() string {
	return fmt.Sprintf("merchant/developed: %s status=%d code=%s message=%s",
		e.Path, e.HTTPStatus, e.Code, e.Message)
}

// Solution returns the WeChat-Pay-documented remediation hint for this code.
// Currently only consults the /pay/transactions/app errorx table — additional
// per-endpoint tables will be wired up as they are populated. Empty if unknown.
func (e *V3Error) Solution() string {
	if e == nil {
		return ""
	}
	if hint, ok := errorx.LookupTransactionsApp(e.Code); ok {
		return hint
	}
	return ""
}
