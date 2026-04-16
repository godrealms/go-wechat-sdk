package aispeech

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/godrealms/go-wechat-sdk/utils"
)

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

// decodeEnvelope delegates to the shared utils.DecodeEnvelope, producing a
// package-local *APIError on non-zero errcodes.
func decodeEnvelope(path string, respBody []byte, out any) error {
	return utils.DecodeEnvelope("aispeech", path, respBody, out, func(code int, msg, p string) error {
		return &APIError{ErrCode: code, ErrMsg: msg, Path: p}
	})
}
