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
func (c *Client) doGet(ctx context.Context, path string, params url.Values, result any) error {
	if err := c.Https.Get(ctx, path, params, result); err != nil {
		return err
	}
	return checkEmbeddedResp(result)
}

// doPost is the POST counterpart to doGet. Same errcode-extraction logic.
func (c *Client) doPost(ctx context.Context, path string, body any, result any) error {
	if err := c.Https.Post(ctx, path, body, result); err != nil {
		return err
	}
	return checkEmbeddedResp(result)
}

// doPostRaw POSTs raw bytes with an explicit Content-Type (no JSON marshaling).
// Use this for endpoints that expect a raw binary body (e.g. voice upload) or
// a pre-formatted text body (e.g. translate). Automatically performs the
// embedded-Resp errcode check on result.
func (c *Client) doPostRaw(ctx context.Context, path string, body []byte, contentType string, result any) error {
	headers := http.Header{}
	if contentType != "" {
		headers.Set("Content-Type", contentType)
	}
	if err := c.Https.DoRequest(ctx, http.MethodPost, path, nil, body, headers, result); err != nil {
		return err
	}
	return checkEmbeddedResp(result)
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
	if err = c.Https.DoRequest(ctx, http.MethodPost, path, nil, buf.Bytes(), headers, result); err != nil {
		return err
	}
	return checkEmbeddedResp(result)
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
