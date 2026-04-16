package aispeech

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// baseResp holds the common WeChat error fields present in every API response.
type baseResp struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// doPost sends a POST JSON request to path with access_token in the query,
// always checks errcode before unmarshalling into out.
//
// On a non-zero errcode it returns *APIError so callers can errors.As() it.
// If the response body is not a valid JSON envelope at all (e.g. an HTML
// error page from a proxy or a binary blob), it returns a wrapped error
// rather than silently treating the body as success — that's the
// "silent unmarshal" footgun this helper exists to prevent.
//
// Note: this helper deliberately uses DoRequestWithRawResponse instead of
// http.Post so that all JSON decoding (and error wrapping) happens in this
// file under a single uniform error format. http.Post would auto-decode
// into json.RawMessage and surface its own "unmarshal response body
// failed" error, which would not be wrapped with our package prefix.
func (c *Client) doPost(ctx context.Context, path string, body any, out any) error {
	tok, err := c.AccessToken(ctx)
	if err != nil {
		return err
	}
	q := url.Values{"access_token": {tok}}

	raw, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("aispeech: %s: marshal request: %w", path, err)
	}
	_, _, respBody, err := c.http.DoRequestWithRawResponse(
		ctx, http.MethodPost, path, q, raw, nil,
	)
	if err != nil {
		return err
	}
	return decodeEnvelope(path, respBody, out)
}

// decodeEnvelope is the shared error-aware unmarshal step. It surfaces
// *APIError for non-zero errcodes and a wrapped error for any malformed
// JSON envelope.
func decodeEnvelope(path string, respBody []byte, out any) error {
	var base baseResp
	if err := json.Unmarshal(respBody, &base); err != nil {
		return fmt.Errorf("aispeech: %s: decode envelope: %w (body snippet: %s)",
			path, err, snippet(respBody))
	}
	if base.ErrCode != 0 {
		return &APIError{ErrCode: base.ErrCode, ErrMsg: base.ErrMsg, Path: path}
	}
	if out != nil {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("aispeech: %s: decode result: %w", path, err)
		}
	}
	return nil
}

// snippet returns at most the first 200 bytes of body as a string, for use
// in error messages. Bodies are typically JSON; we deliberately do not
// attempt to redact PII because the snippet is short enough that callers
// who need to log it can decide for themselves.
func snippet(b []byte) string {
	const max = 200
	if len(b) <= max {
		return string(b)
	}
	return string(b[:max]) + "...(truncated)"
}
