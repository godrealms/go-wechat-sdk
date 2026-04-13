package utils

// WechatAPIError is implemented by every package-specific WeChat API error type
// in this SDK. It allows callers to inspect the numeric error code and human-
// readable message without importing a concrete package.
//
// Usage:
//
//	var apiErr utils.WechatAPIError
//	if errors.As(err, &apiErr) {
//	    log.Printf("WeChat API error %d: %s", apiErr.Code(), apiErr.Message())
//	}
type WechatAPIError interface {
	error
	// Code returns the numeric errcode returned by the WeChat API (e.g. 40001).
	Code() int
	// Message returns the human-readable errmsg returned by the WeChat API.
	Message() string
}
