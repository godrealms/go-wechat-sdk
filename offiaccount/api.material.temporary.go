package offiaccount

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// UploadTempMedia 上传临时素材
// mediaType: 媒体文件类型
// filename: 媒体文件名
// reader: 媒体文件内容读取器
func (c *Client) UploadTempMedia(ctx context.Context, mediaType TempMediaType, filename string, reader io.Reader) (*UploadTempMediaResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}

	// 读取文件到内存（微信临时素材上限 10MB，可接受）
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read media failed: %w", err)
	}

	path := fmt.Sprintf("/cgi-bin/media/upload?access_token=%s&type=%s",
		token, url.QueryEscape(string(mediaType)))

	var result UploadTempMediaResult
	if err = c.doPostMultipartFile(ctx, path, "media", filename, data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UploadTempMediaByPath 通过文件路径上传临时素材
// mediaType: 媒体文件类型
// filepath: 媒体文件路径
func (c *Client) UploadTempMediaByPath(ctx context.Context, mediaType TempMediaType, filepath string) (*UploadTempMediaResult, error) {
	// 打开文件
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("open file failed: %v", err)
	}
	defer file.Close()

	// 获取文件名
	parts := strings.Split(filepath, "/")
	filename := parts[len(parts)-1]

	// 上传临时素材
	return c.UploadTempMedia(ctx, mediaType, filename, file)
}

// GetTempMedia 获取临时素材
// mediaID: 媒体文件ID
//
// 返回值：
//   - 图片 / 语音 / 缩略图：第一个返回值为原始字节
//   - 视频素材：第二个返回值为 *GetTempMediaVideoResult（内含 video_url）
//
// 如果服务端返回 errcode 非 0，直接返回 *WeixinError。
func (c *Client) GetTempMedia(ctx context.Context, mediaID string) ([]byte, *GetTempMediaVideoResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, nil, err
	}
	// 构造请求URL
	params := url.Values{}
	params.Add("access_token", token)
	params.Add("media_id", mediaID)
	path := fmt.Sprintf("/cgi-bin/media/get?%s", params.Encode())

	// 构建完整URL
	fullURL := c.Https.BaseURL + path

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("the http request was created failed: %v", err)
	}

	// 发送请求
	resp, err := c.Https.Client.Do(httpReq)
	if err != nil {
		return nil, nil, fmt.Errorf("sending http request failed: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read response failed: %v", err)
	}

	// 如果看起来像 JSON（以 '{' 开头），先探测 errcode 再区分 video / error
	if len(respBody) > 0 && respBody[0] == '{' {
		var probe Resp
		if err = json.Unmarshal(respBody, &probe); err == nil && probe.ErrCode != 0 {
			return nil, nil, &WeixinError{ErrCode: probe.ErrCode, ErrMsg: probe.ErrMsg}
		}

		// 无 errcode → 视频素材 JSON，解析 video_url
		var result GetTempMediaVideoResult
		if err = json.Unmarshal(respBody, &result); err != nil {
			return nil, nil, fmt.Errorf("unmarshal response body failed: %v:%s", err, string(respBody))
		}
		// 如果 Title 或 VideoURL 有值，视作视频结果；否则回落到原始字节
		if result.VideoURL != "" {
			return nil, &result, nil
		}
	}

	// 返回原始文件数据（图片 / 语音等）
	return respBody, nil, nil
}

// GetHDVoice 获取高清语音素材
// mediaID: 语音素材ID
func (c *Client) GetHDVoice(ctx context.Context, mediaID string) ([]byte, *Resp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, nil, err
	}
	// 构造请求URL
	params := url.Values{}
	params.Add("access_token", token)
	params.Add("media_id", mediaID)
	path := fmt.Sprintf("/cgi-bin/media/get/jssdk?%s", params.Encode())

	// 构建完整URL
	fullURL := c.Https.BaseURL + path

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("the http request was created failed: %v", err)
	}

	// 发送请求
	resp, err := c.Https.Client.Do(httpReq)
	if err != nil {
		return nil, nil, fmt.Errorf("sending http request failed: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read response failed: %v", err)
	}

	// 如果看起来像 JSON（以 '{' 开头），一定是错误响应
	if len(respBody) > 0 && respBody[0] == '{' {
		var probe Resp
		if err = json.Unmarshal(respBody, &probe); err != nil {
			return nil, nil, fmt.Errorf("unmarshal response body failed: %v:%s", err, string(respBody))
		}
		if probe.ErrCode != 0 {
			return nil, &probe, &WeixinError{ErrCode: probe.ErrCode, ErrMsg: probe.ErrMsg}
		}
		return nil, &probe, nil
	}

	// 返回原始文件数据
	return respBody, nil, nil
}
