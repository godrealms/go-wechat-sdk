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
func (cc *CorpClient) doUpload(ctx context.Context, path string, extra url.Values, fieldName, fileName string, fileData io.Reader, out any) error {
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
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("isv: upload http %d: %s", resp.StatusCode, string(body))
	}
	return decodeRaw(path, body, out)
}

// UploadMedia uploads a temporary media file (image, voice, video, or file).
// mediaType must be one of: "image", "voice", "video", "file" — the WeCom
// (企业微信) accepted set, which differs from mini-program (no "thumb").
func (cc *CorpClient) UploadMedia(ctx context.Context, mediaType, fileName string, fileData io.Reader) (*UploadMediaResp, error) {
	if _, ok := validWWMediaTypes[mediaType]; !ok {
		return nil, fmt.Errorf("isv: UploadMedia: mediaType must be one of image/voice/video/file, got %q", mediaType)
	}
	if fileName == "" {
		return nil, fmt.Errorf("isv: UploadMedia: fileName is required")
	}
	if fileData == nil {
		return nil, fmt.Errorf("isv: UploadMedia: fileData is required")
	}
	extra := url.Values{"type": {mediaType}}
	var resp UploadMediaResp
	if err := cc.doUpload(ctx, "/cgi-bin/media/upload", extra, "media", fileName, fileData, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
