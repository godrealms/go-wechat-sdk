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

// UploadTempMediaResp 上传临时素材响应。
type UploadTempMediaResp struct {
	Type      string `json:"type"`
	MediaID   string `json:"media_id"`
	CreatedAt int64  `json:"created_at"`
}

// UploadTempMedia 上传临时素材（图片/语音/视频/缩略图）。
// mediaType: "image" | "voice" | "video" | "thumb"
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
	w.Close()

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
		return nil, fmt.Errorf("mini_program: upload media errcode=%d errmsg=%s", resp.ErrCode, resp.ErrMsg)
	}
	return &resp.UploadTempMediaResp, nil
}

// GetTempMedia 获取临时素材。返回文件二进制内容。
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
			return nil, fmt.Errorf("mini_program: get media errcode=%d errmsg=%s", resp.ErrCode, resp.ErrMsg)
		}
	}
	return respBody, nil
}
