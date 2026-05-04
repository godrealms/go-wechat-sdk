package offiaccount

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
)

// doGet wraps c.Https.Get with an automatic errcode check against an embedded
// Resp field on the result struct. This exists because historically ~85% of
// api.*.go methods parsed responses but never checked result.ErrCode,
// silently dropping business errors. See audit 2026-04-14.
//
// If result's underlying struct has no embedded Resp field, doGet behaves
// exactly like c.Https.Get — this keeps it safe to retrofit across call sites
// whose result types may not uniformly embed Resp.
//
// On a token-expired response (errcode 40001/40014/42001/42007), doGet
// invalidates the cached token, fetches a fresh one, replaces access_token in
// params or path, and retries the request once. See IsTokenExpired.
func (c *Client) doGet(ctx context.Context, path string, params url.Values, result any) error {
	return c.doWithRetry(ctx, &path, &params, func() error {
		if err := c.Https.Get(ctx, path, params, result); err != nil {
			return err
		}
		return checkEmbeddedResp(result)
	})
}

// doPost is the POST counterpart to doGet. Same errcode-extraction logic and
// 40001 self-heal retry.
func (c *Client) doPost(ctx context.Context, path string, body any, result any) error {
	return c.doWithRetry(ctx, &path, nil, func() error {
		if err := c.Https.Post(ctx, path, body, result); err != nil {
			return err
		}
		return checkEmbeddedResp(result)
	})
}

// doPostRaw POSTs raw bytes with an explicit Content-Type (no JSON marshaling).
// Use this for endpoints that expect a raw binary body (e.g. voice upload) or
// a pre-formatted text body (e.g. translate). Automatically performs the
// embedded-Resp errcode check on result and the 40001 self-heal retry.
func (c *Client) doPostRaw(ctx context.Context, path string, body []byte, contentType string, result any) error {
	headers := http.Header{}
	if contentType != "" {
		headers.Set("Content-Type", contentType)
	}
	return c.doWithRetry(ctx, &path, nil, func() error {
		if err := c.Https.DoRequest(ctx, http.MethodPost, path, nil, body, headers, result); err != nil {
			return err
		}
		return checkEmbeddedResp(result)
	})
}

// doPostMultipartFile POSTs a single file field as multipart/form-data.
// Field name and filename are WeChat-specific (typically "img" / "media").
// Performs the embedded-Resp errcode check on result.
//
// For the common single-file case, callers should use this. When the endpoint
// also requires auxiliary text fields (e.g. /cgi-bin/material/add_material's
// description, /cgi-bin/media/uploadimg's type), use doPostMultipart.
func (c *Client) doPostMultipartFile(ctx context.Context, path, fieldName, filename string, fileBytes []byte, result any) error {
	return c.doPostMultipart(ctx, path, fieldName, filename, fileBytes, nil, result)
}

// doPostMultipart is doPostMultipartFile plus optional extra text fields
// written alongside the file part. Extra fields are written in map-iteration
// order; WeChat does not care about ordering for any known endpoint.
func (c *Client) doPostMultipart(
	ctx context.Context,
	path, fieldName, filename string,
	fileBytes []byte,
	extraFields map[string]string,
	result any,
) error {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile(fieldName, filename)
	if err != nil {
		return err
	}
	if _, err = part.Write(fileBytes); err != nil {
		return err
	}
	for k, v := range extraFields {
		if err = writer.WriteField(k, v); err != nil {
			return err
		}
	}
	if err = writer.Close(); err != nil {
		return err
	}
	headers := http.Header{}
	headers.Set("Content-Type", writer.FormDataContentType())
	body := buf.Bytes()
	return c.doWithRetry(ctx, &path, nil, func() error {
		if err := c.Https.DoRequest(ctx, http.MethodPost, path, nil, body, headers, result); err != nil {
			return err
		}
		return checkEmbeddedResp(result)
	})
}

// doWithRetry runs fn once. If the error indicates an expired access_token
// (see IsTokenExpired), it invalidates the cache, fetches a fresh token,
// patches access_token in either *paramsRef (for GET-style query params) or
// *pathRef (when the token is embedded in the path's query string), and runs
// fn one more time.
//
// The retry is skipped when:
//   - the error is not a token-expired error
//   - AccessTokenE fails on refresh (the original error is returned)
//   - neither paramsRef nor pathRef carries an access_token to replace
//
// At most one retry is attempted per call. If the second attempt also returns
// a token-expired error, that error is propagated to the caller.
func (c *Client) doWithRetry(
	ctx context.Context,
	pathRef *string,
	paramsRef *url.Values,
	fn func() error,
) error {
	err := fn()
	if !IsTokenExpired(err) {
		return err
	}
	c.Invalidate()
	newToken, terr := c.AccessTokenE(ctx)
	if terr != nil {
		return err
	}
	if !patchAccessToken(pathRef, paramsRef, newToken) {
		return err
	}
	return fn()
}

// patchAccessToken replaces the access_token value in either paramsRef (if
// non-nil and contains an access_token key) or pathRef (if its query string
// contains access_token=...). Returns true if a replacement was made.
//
// paramsRef is replaced with a clone to avoid mutating a caller-owned map.
func patchAccessToken(pathRef *string, paramsRef *url.Values, newToken string) bool {
	if paramsRef != nil && (*paramsRef).Get("access_token") != "" {
		clone := url.Values{}
		for k, vs := range *paramsRef {
			clone[k] = append([]string(nil), vs...)
		}
		clone.Set("access_token", newToken)
		*paramsRef = clone
		return true
	}
	if pathRef == nil {
		return false
	}
	u, err := url.Parse(*pathRef)
	if err != nil {
		return false
	}
	q := u.Query()
	if q.Get("access_token") == "" {
		return false
	}
	q.Set("access_token", newToken)
	u.RawQuery = q.Encode()
	*pathRef = u.String()
	return true
}

// checkEmbeddedResp uses reflection to find an embedded Resp field on the
// given value. Handles three shapes:
//  1. result is *Resp directly (e.g. result := &Resp{})
//  2. result is a pointer to a struct that embeds Resp (common case)
//  3. anything else — returns nil, to stay backwards compatible with response
//     types that don't follow the embed-Resp pattern
//
// If a Resp is found and its ErrCode is non-zero, returns a *WeixinError.
func checkEmbeddedResp(result any) error {
	if result == nil {
		return nil
	}
	// Fast path: the caller passed *Resp directly.
	if resp, ok := result.(*Resp); ok {
		return CheckResp(resp)
	}
	v := reflect.ValueOf(result)
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}
	// Root-struct path: the underlying type IS Resp (covers e.g. var result Resp
	// passed as &result when the caller didn't alias it through the interface).
	if v.Type() == reflect.TypeOf(Resp{}) && v.CanAddr() {
		return CheckResp(v.Addr().Interface().(*Resp))
	}
	f := v.FieldByName("Resp")
	if !f.IsValid() || !f.CanAddr() {
		return nil
	}
	resp, ok := f.Addr().Interface().(*Resp)
	if !ok {
		return nil
	}
	return CheckResp(resp)
}
