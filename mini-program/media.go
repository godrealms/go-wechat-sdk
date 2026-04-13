package mini_program

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
)

// UploadTempMediaResp is the response from the upload-temporary-media API.
type UploadTempMediaResp struct {
	Type      string `json:"type"`
	MediaID   string `json:"media_id"`
	CreatedAt int64  `json:"created_at"`
}

// UploadTempMedia uploads a temporary media asset (image, voice, video, or thumbnail).
// mediaType must be one of: "image", "voice", "video", "thumb".
func (c *Client) UploadTempMedia(ctx context.Context, mediaType, fileName string, fileData io.Reader) (*UploadTempMediaResp, error) {
	tok, err := c.AccessToken(ctx)
	if err != nil {
		return nil, err
	}
	q := url.Values{"access_token": {tok}, "type": {mediaType}}

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("media", fileName)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(fw, fileData); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("mini_program: close multipart writer: %w", err)
	}

	_, _, respBody, err := c.http.DoRequestWithRawResponse(
		ctx, http.MethodPost, "/cgi-bin/media/upload", q,
		buf.Bytes(),
		http.Header{"Content-Type": {w.FormDataContentType()}},
	)
	if err != nil {
		return nil, err
	}
	var resp struct {
		baseResp
		UploadTempMediaResp
	}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}
	if resp.ErrCode != 0 {
		return nil, &APIError{ErrCode: resp.ErrCode, ErrMsg: resp.ErrMsg, Path: "/cgi-bin/media/upload"}
	}
	return &resp.UploadTempMediaResp, nil
}

// GetTempMedia retrieves a temporary media asset by its media ID and returns the raw file bytes.
func (c *Client) GetTempMedia(ctx context.Context, mediaID string) ([]byte, error) {
	tok, err := c.AccessToken(ctx)
	if err != nil {
		return nil, err
	}
	q := url.Values{"access_token": {tok}, "media_id": {mediaID}}
	_, _, respBody, err := c.http.DoRequestWithRawResponse(
		ctx, http.MethodGet, "/cgi-bin/media/get", q, nil, nil,
	)
	if err != nil {
		return nil, err
	}
	if len(respBody) > 0 && respBody[0] == '{' {
		var resp baseResp
		if json.Unmarshal(respBody, &resp) == nil && resp.ErrCode != 0 {
			return nil, &APIError{ErrCode: resp.ErrCode, ErrMsg: resp.ErrMsg, Path: "/cgi-bin/media/get"}
		}
	}
	return respBody, nil
}
