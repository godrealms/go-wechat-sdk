package offiaccount

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

// GetMaterial 获取永久素材
// mediaID: 要获取的素材的media_id
//
// 返回值根据素材类型不同：
//   - 图文素材：第一个返回值非 nil
//   - 视频素材：第二个返回值非 nil
//   - 图片 / 语音 / 其他二进制素材：第三个返回值是原始字节
//
// 如果服务端返回 errcode 非 0，返回 *WeixinError。
//
// 实现走 c.Https.DoRequestWithRawResponse + c.doWithRetry，因此享受 SDK
// 的 logger 凭据脱敏、响应体大小上限和 40001 自愈重试,与其他 API 方法
// 保持一致。
func (c *Client) GetMaterial(ctx context.Context, mediaID string) (*GetMaterialNewsResult, *GetMaterialVideoResult, []byte, error) {
	if mediaID == "" {
		return nil, nil, nil, fmt.Errorf("offiaccount: GetMaterial: mediaID is required")
	}
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	body := map[string]any{"media_id": mediaID}
	rawBody, err := json.Marshal(body)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("offiaccount: GetMaterial: marshal request body: %w", err)
	}
	headers := http.Header{"Content-Type": []string{"application/json"}}
	path := fmt.Sprintf("/cgi-bin/material/get_material?access_token=%s", token)

	var (
		news   *GetMaterialNewsResult
		video  *GetMaterialVideoResult
		binary []byte
	)

	err = c.doWithRetry(ctx, &path, nil, func() error {
		// Reset captures so a self-heal retry starts clean.
		news, video, binary = nil, nil, nil
		_, _, respBody, derr := c.Https.DoRequestWithRawResponse(ctx, http.MethodPost, path, nil, rawBody, headers)
		if derr != nil {
			return derr
		}

		// WeChat returns binary for image/voice/thumb but always JSON for
		// errors. Probe by first byte: if it isn't '{' the body is
		// definitely binary (no JSON object can start with another byte).
		if len(respBody) == 0 || respBody[0] != '{' {
			binary = respBody
			return nil
		}

		// Body looks like JSON. Decode the envelope first; a parse failure
		// here is a real problem (a status-200 body that says it's JSON but
		// isn't), not "treat as binary" — surface it.
		var probe Resp
		if uerr := json.Unmarshal(respBody, &probe); uerr != nil {
			return fmt.Errorf("offiaccount: GetMaterial: malformed JSON response: %w", uerr)
		}
		if probe.ErrCode != 0 {
			return &WeixinError{ErrCode: probe.ErrCode, ErrMsg: probe.ErrMsg}
		}

		// errcode is zero — distinguish news vs. video by their distinguishing fields.
		var newsResult GetMaterialNewsResult
		if uerr := json.Unmarshal(respBody, &newsResult); uerr == nil && len(newsResult.NewsItem) > 0 {
			news = &newsResult
			return nil
		}
		var videoResult GetMaterialVideoResult
		if uerr := json.Unmarshal(respBody, &videoResult); uerr == nil && videoResult.Title != "" {
			video = &videoResult
			return nil
		}

		// JSON envelope with no recognised payload — return raw bytes for
		// callers who want to inspect them (extremely rare).
		binary = respBody
		return nil
	})
	if err != nil {
		return nil, nil, nil, err
	}
	return news, video, binary, nil
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
// filePath: 图片文件路径
func (c *Client) UploadImageByPath(ctx context.Context, filePath string) (*UploadImgResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("offiaccount: UploadImageByPath: open file: %w", err)
	}
	defer file.Close()

	// filepath.Base handles both "/" and "\\" — necessary for Windows callers
	// since the previous strings.Split(p, "/") returned the full path verbatim
	// when given a Windows-style "C:\\foo\\bar.jpg".
	return c.UploadImg(ctx, filepath.Base(filePath), file)
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
// filePath: 媒体文件路径
// description: 视频素材描述信息（仅视频素材需要）
func (c *Client) AddMaterialByPath(ctx context.Context, materialType MaterialType, filePath string, description *AddMaterialVideoDescription) (*AddMaterialResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("offiaccount: AddMaterialByPath: open file: %w", err)
	}
	defer file.Close()

	return c.AddMaterial(ctx, materialType, filepath.Base(filePath), file, description)
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
	body := map[string]any{
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
