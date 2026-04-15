package utils

// WechatAPIError is implemented by every package-specific WeChat **legacy** API
// error type in this SDK (公众号 / 小程序 / 企业微信 / Channels / 等)。It allows
// callers to inspect the numeric error code and human-readable message without
// importing a concrete package.
//
// Usage:
//
//	var apiErr utils.WechatAPIError
//	if errors.As(err, &apiErr) {
//	    log.Printf("WeChat API error %d: %s", apiErr.Code(), apiErr.Message())
//	}
//
// **微信支付 V3 API** 使用 string 形式的错误码（例如 "PARAM_ERROR"），
// 与本接口的 `Code() int` 不兼容；请改用 WechatAPIV3Error。
type WechatAPIError interface {
	error
	// Code returns the numeric errcode returned by the WeChat API (e.g. 40001).
	Code() int
	// Message returns the human-readable errmsg returned by the WeChat API.
	Message() string
}

// WechatAPIV3Error is implemented by error types that wrap a WeChat **Pay V3**
// API error envelope. V3 uses string codes like "PARAM_ERROR" /
// "OUT_TRADE_NO_USED" instead of integer codes, so it cannot share the
// WechatAPIError interface. The HTTPStatus method is included because V3
// frequently uses HTTP status as an additional discriminator
// (e.g. 429 = rate limited, 500 = server error).
//
// Usage:
//
//	var v3 utils.WechatAPIV3Error
//	if errors.As(err, &v3) {
//	    log.Printf("WeChat Pay V3 error %s (HTTP %d): %s",
//	        v3.Code(), v3.HTTPStatus(), v3.Message())
//	}
type WechatAPIV3Error interface {
	error
	// Code returns the V3 string error code (e.g. "PARAM_ERROR").
	Code() string
	// Message returns the human-readable error message.
	Message() string
	// HTTPStatus returns the HTTP status code of the failed response.
	HTTPStatus() int
}
