package offiaccount

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
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
	// 获取access_token
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}

	// 构造请求URL
	params := url.Values{}
	params.Add("access_token", token)
	params.Add("type", string(mediaType))
	path := fmt.Sprintf("/cgi-bin/media/upload?%s", params.Encode())

	// 创建multipart表单
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// 添加文件字段
	part, err := writer.CreateFormFile("media", filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create a file field: %v", err)
	}

	// 复制文件内容
	_, err = io.Copy(part, reader)
	if err != nil {
		return nil, fmt.Errorf("copying file contents failed: %v", err)
	}

	// 关闭writer
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("closing writer failed: %v", err)
	}

	// 构建完整URL
	fullURL := c.Https.BaseURL + path

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", fullURL, &requestBody)
	if err != nil {
		return nil, fmt.Errorf("the http request was created failed: %v", err)
	}

	// 设置Content-Type
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())

	// 发送请求
	resp, err := c.Https.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("sending http request failed: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %v", err)
	}

	// 检查响应状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, string(respBody))
	}

	// 解析响应
	var result UploadTempMediaResult
	if len(respBody) > 0 {
		if err = json.Unmarshal(respBody, &result); err != nil {
			return nil, fmt.Errorf("unmarshal response body failed: %v:%s", err, string(respBody))
		}
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

	// 检查响应是否为JSON格式（错误情况）
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		// 解析为错误响应
		var result GetTempMediaVideoResult
		if len(respBody) > 0 {
			if err = json.Unmarshal(respBody, &result); err != nil {
				return nil, nil, fmt.Errorf("unmarshal response body failed: %v:%s", err, string(respBody))
			}
		}
		return nil, &result, nil
	}

	// 检查是否为视频素材的JSON响应
	if strings.Contains(string(respBody), "video_url") {
		var result GetTempMediaVideoResult
		if len(respBody) > 0 {
			if err = json.Unmarshal(respBody, &result); err != nil {
				return nil, nil, fmt.Errorf("unmarshal response body failed: %v:%s", err, string(respBody))
			}
		}
		return nil, &result, nil
	}

	// 返回原始文件数据
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

	// 检查响应是否为JSON格式（错误情况）
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		// 解析为错误响应
		var result Resp
		if len(respBody) > 0 {
			if err = json.Unmarshal(respBody, &result); err != nil {
				return nil, nil, fmt.Errorf("unmarshal response body failed: %v:%s", err, string(respBody))
			}
		}
		return nil, &result, nil
	}

	// 返回原始文件数据
	return respBody, nil, nil
}
