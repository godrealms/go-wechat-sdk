package isv

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
)

// doUpload sends a multipart/form-data POST with access_token injected.
func (cc *CorpClient) doUpload(ctx context.Context, path string, extra url.Values, fieldName, fileName string, fileData io.Reader, out interface{}) error {
	tok, err := cc.AccessToken(ctx)
	if err != nil {
		return err
	}
	q := url.Values{"access_token": {tok}}
	for k, vs := range extra {
		q[k] = vs
	}

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile(fieldName, fileName)
	if err != nil {
		return fmt.Errorf("isv: create form file: %w", err)
	}
	if _, err := io.Copy(fw, fileData); err != nil {
		return fmt.Errorf("isv: copy file data: %w", err)
	}
	w.Close()

	fullURL := cc.parent.baseURL + path + "?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, &buf)
	if err != nil {
		return fmt.Errorf("isv: new request: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := cc.parent.http.Do(req)
	if err != nil {
		return fmt.Errorf("isv: upload: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("isv: read body: %w", err)
	}
	return decodeRaw(body, out)
}

// UploadMedia uploads a temporary media file (image, voice, video, or file).
func (cc *CorpClient) UploadMedia(ctx context.Context, mediaType, fileName string, fileData io.Reader) (*UploadMediaResp, error) {
	extra := url.Values{"type": {mediaType}}
	var resp UploadMediaResp
	if err := cc.doUpload(ctx, "/cgi-bin/media/upload", extra, "media", fileName, fileData, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
