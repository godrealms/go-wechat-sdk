package offiaccount

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// GetMaterial 获取永久素材
// mediaID: 要获取的素材的media_id
//
// 返回值根据素材类型不同：
//   - 图文素材：第一个返回值非 nil
//   - 视频素材：第二个返回值非 nil
//   - 图片 / 语音 / 其他二进制素材：第三个返回值是原始字节
//
// 如果服务端返回 errcode 非 0，直接返回 *WeixinError。
func (c *Client) GetMaterial(ctx context.Context, mediaID string) (*GetMaterialNewsResult, *GetMaterialVideoResult, []byte, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/material/get_material?access_token=%s", token)

	// 构造请求体
	body := map[string]interface{}{
		"media_id": mediaID,
	}

	// 直接发送POST请求获取响应数据
	fullURL := c.Https.BaseURL + path
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("marshal request body failed: %w", err)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "POST", fullURL, bytes.NewReader(jsonBody))
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

	// 先探测 errcode：对于非图文/视频素材，微信可能返回图片/语音原始二进制，
	// 但当出错时一定是 JSON {errcode, errmsg}。只有在看起来是 JSON 时才尝试解析。
	if len(respBody) > 0 && respBody[0] == '{' {
		var probe Resp
		if err = json.Unmarshal(respBody, &probe); err == nil && probe.ErrCode != 0 {
			return nil, nil, nil, &WeixinError{ErrCode: probe.ErrCode, ErrMsg: probe.ErrMsg}
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
	}

	// 图片 / 语音 / 缩略图等二进制素材：返回原始数据
	return nil, nil, respBody, nil
}

// GetMaterialCount 获取素材总数
func (c *Client) GetMaterialCount(ctx context.Context) (*MaterialCount, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/material/get_materialcount?access_token=%s", token)

	// 发送请求
	var result MaterialCount
	err = c.doGet(ctx, path, nil, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// BatchGetMaterial 批量获取素材
// req: 批量获取素材请求参数
func (c *Client) BatchGetMaterial(ctx context.Context, req *BatchGetMaterialRequest) (*BatchGetMaterialResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/material/batchget_material?access_token=%s", token)

	// 发送请求
	var result BatchGetMaterialResult
	err = c.doPost(ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// UploadImg 上传图文消息内的图片
// filename: 图片文件名
// reader: 图片文件内容读取器
func (c *Client) UploadImg(ctx context.Context, filename string, reader io.Reader) (*UploadImgResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read image failed: %w", err)
	}

	path := fmt.Sprintf("/cgi-bin/media/uploadimg?access_token=%s", token)

	var result UploadImgResult
	if err = c.doPostMultipartFile(ctx, path, "media", filename, data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UploadImageByPath 通过文件路径上传图文消息内的图片
// filepath: 图片文件路径
func (c *Client) UploadImageByPath(ctx context.Context, filepath string) (*UploadImgResult, error) {
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
	return c.UploadImg(ctx, filename, file)
}

// AddMaterial 新增永久素材
// materialType: 媒体类型
// filename: 媒体文件名
// reader: 媒体文件内容读取器
// description: 视频素材描述信息（仅视频素材需要）
func (c *Client) AddMaterial(ctx context.Context, materialType MaterialType, filename string, reader io.Reader, description *AddMaterialVideoDescription) (*AddMaterialResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read media failed: %w", err)
	}

	path := fmt.Sprintf("/cgi-bin/material/add_material?access_token=%s&type=%s",
		token, url.QueryEscape(string(materialType)))

	// 视频素材需要附带 description 字段
	var extra map[string]string
	if materialType == MaterialTypeVideo && description != nil {
		descriptionData, err := json.Marshal(description)
		if err != nil {
			return nil, fmt.Errorf("marshal description failed: %w", err)
		}
		extra = map[string]string{"description": string(descriptionData)}
	}

	var result AddMaterialResult
	if err = c.doPostMultipart(ctx, path, "media", filename, data, extra, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// AddMaterialByPath 通过文件路径新增永久素材
// materialType: 媒体类型
// filepath: 媒体文件路径
// description: 视频素材描述信息（仅视频素材需要）
func (c *Client) AddMaterialByPath(ctx context.Context, materialType MaterialType, filepath string, description *AddMaterialVideoDescription) (*AddMaterialResult, error) {
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
	return c.AddMaterial(ctx, materialType, filename, file, description)
}

// DelMaterial 删除永久素材
// mediaID: 要删除的素材media_id
func (c *Client) DelMaterial(ctx context.Context, mediaID string) (*Resp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/material/del_material?access_token=%s", token)

	// 构造请求体
	body := map[string]interface{}{
		"media_id": mediaID,
	}

	// 发送请求
	var result Resp
	err = c.doPost(ctx, path, body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
