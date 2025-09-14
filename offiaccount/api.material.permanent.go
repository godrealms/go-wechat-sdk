package offiaccount

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// GetMaterial 获取永久素材
// mediaID: 要获取的素材的media_id
func (c *Client) GetMaterial(mediaID string) (*GetMaterialNewsResult, *GetMaterialVideoResult, []byte, error) {
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/material/get_material?access_token=%s", c.GetAccessToken())

	// 构造请求体
	body := map[string]interface{}{
		"media_id": mediaID,
	}

	// 直接发送POST请求获取响应数据
	fullURL := c.Https.BaseURL + path
	jsonBody, _ := json.Marshal(body)

	// 创建请求
	req, err := http.NewRequest("POST", fullURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("create request failed: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := c.Https.Client.Do(req)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("read response body failed: %w", err)
	}

	// 检查响应状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, nil, nil, fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, string(respBody))
	}

	// 尝试解析为图文素材
	var newsResult GetMaterialNewsResult
	if err = json.Unmarshal(respBody, &newsResult); err == nil && len(newsResult.NewsItem) > 0 {
		return &newsResult, nil, nil, nil
	}

	// 尝试解析为视频素材
	var videoResult GetMaterialVideoResult
	if err = json.Unmarshal(respBody, &videoResult); err == nil && videoResult.Title != "" {
		return nil, &videoResult, nil, nil
	}

	// 如果都不是，则返回原始数据，由调用方处理
	return nil, nil, respBody, nil
}

// GetMaterialCount 获取素材总数
func (c *Client) GetMaterialCount() (*MaterialCount, error) {
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/material/get_materialcount?access_token=%s", c.GetAccessToken())

	// 发送请求
	var result MaterialCount
	err := c.Https.Get(c.ctx, path, nil, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// BatchGetMaterial 批量获取素材
// req: 批量获取素材请求参数
func (c *Client) BatchGetMaterial(req *BatchGetMaterialRequest) (*BatchGetMaterialResult, error) {
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/material/batchget_material?access_token=%s", c.GetAccessToken())

	// 发送请求
	var result BatchGetMaterialResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// UploadImg 上传图文消息内的图片
// filename: 图片文件名
// reader: 图片文件内容读取器
func (c *Client) UploadImg(filename string, reader io.Reader) (*UploadImgResult, error) {
	// 获取access_token
	token := c.GetAccessToken()
	if token == "" {
		return nil, fmt.Errorf("get access token failed")
	}

	// 构造请求URL
	params := url.Values{}
	params.Add("access_token", token)
	path := fmt.Sprintf("/cgi-bin/media/uploadimg?%s", params.Encode())

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
	httpReq, err := http.NewRequest("POST", fullURL, &requestBody)
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
	var result UploadImgResult
	if len(respBody) > 0 {
		if err = json.Unmarshal(respBody, &result); err != nil {
			return nil, fmt.Errorf("unmarshal response body failed: %v:%s", err, string(respBody))
		}
	}

	return &result, nil
}

// UploadImageByPath 通过文件路径上传图文消息内的图片
// filepath: 图片文件路径
func (c *Client) UploadImageByPath(filepath string) (*UploadImgResult, error) {
	// 打开文件
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("open file failed: %v", err)
	}
	defer file.Close()

	// 获取文件名
	parts := strings.Split(filepath, "/")
	filename := parts[len(parts)-1]

	// 上传图片
	return c.UploadImg(filename, file)
}

// AddMaterial 新增永久素材
// materialType: 媒体类型
// filename: 媒体文件名
// reader: 媒体文件内容读取器
// description: 视频素材描述信息（仅视频素材需要）
func (c *Client) AddMaterial(materialType MaterialType, filename string, reader io.Reader, description *AddMaterialVideoDescription) (*AddMaterialResult, error) {
	// 获取access_token
	token := c.GetAccessToken()
	if token == "" {
		return nil, fmt.Errorf("get access token failed")
	}

	// 构造请求URL
	params := url.Values{}
	params.Add("access_token", token)
	params.Add("type", string(materialType))
	path := fmt.Sprintf("/cgi-bin/material/add_material?%s", params.Encode())

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

	// 如果是视频素材，添加描述信息
	if materialType == MaterialTypeVideo && description != nil {
		descriptionData, err := json.Marshal(description)
		if err != nil {
			return nil, fmt.Errorf("marshal description failed: %v", err)
		}
		_ = writer.WriteField("description", string(descriptionData))
	}

	// 关闭writer
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("closing writer failed: %v", err)
	}

	// 构建完整URL
	fullURL := c.Https.BaseURL + path

	// 创建HTTP请求
	httpReq, err := http.NewRequest("POST", fullURL, &requestBody)
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
	var result AddMaterialResult
	if len(respBody) > 0 {
		if err = json.Unmarshal(respBody, &result); err != nil {
			return nil, fmt.Errorf("unmarshal response body failed: %v:%s", err, string(respBody))
		}
	}

	return &result, nil
}

// AddMaterialByPath 通过文件路径新增永久素材
// materialType: 媒体类型
// filepath: 媒体文件路径
// description: 视频素材描述信息（仅视频素材需要）
func (c *Client) AddMaterialByPath(materialType MaterialType, filepath string, description *AddMaterialVideoDescription) (*AddMaterialResult, error) {
	// 打开文件
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("open file failed: %v", err)
	}
	defer file.Close()

	// 获取文件名
	parts := strings.Split(filepath, "/")
	filename := parts[len(parts)-1]

	// 新增素材
	return c.AddMaterial(materialType, filename, file, description)
}

// DelMaterial 删除永久素材
// mediaID: 要删除的素材media_id
func (c *Client) DelMaterial(mediaID string) (*Resp, error) {
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/material/del_material?access_token=%s", c.GetAccessToken())

	// 构造请求体
	body := map[string]interface{}{
		"media_id": mediaID,
	}

	// 发送请求
	var result Resp
	err := c.Https.Post(c.ctx, path, body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
