package offiaccount

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
)

// SetInvoiceUrl 设置商户联系方式（获取开票平台识别码）
func (c *Client) SetInvoiceUrl() (*SetUrlResult, error) {
	// 构造请求URL
	path := "/card/invoice/seturl"

	// 发送请求
	var result SetUrlResult
	err := c.Https.Post(c.ctx, path, nil, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetPdf 获取pdf文件
// req: 获取pdf文件请求参数
func (c *Client) GetPdf(req *GetPdfRequest) (*GetPdfResult, error) {
	// 构造请求URL
	path := "/card/invoice/platform/getpdf"

	// 发送请求
	var result GetPdfResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateInvoiceStatus 更新发票状态
// req: 更新发票状态请求参数
func (c *Client) UpdateInvoiceStatus(req *UpdateInvoiceStatusRequest) (*Resp, error) {
	// 构造请求URL
	path := "/card/invoice/platform/updatestatus"

	// 发送请求
	var result Resp
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// SetPdf 设置pdf文件
// pdf: pdf文件内容
func (c *Client) SetPdf(filename string, pdf io.Reader) (*SetPdfResult, error) {
	// 获取access_token
	token := c.GetAccessToken()
	if token == "" {
		return nil, fmt.Errorf("get access token failed")
	}

	// 构造请求URL
	params := url.Values{}
	params.Add("access_token", token)
	path := fmt.Sprintf("/card/invoice/platform/setpdf?%s", params.Encode())

	// 创建multipart表单
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// 添加文件字段
	part, err := writer.CreateFormFile("pdf", filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create a file field: %v", err)
	}

	// 复制文件内容
	_, err = io.Copy(part, pdf)
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
	var result SetPdfResult
	if len(respBody) > 0 {
		if err = json.Unmarshal(respBody, &result); err != nil {
			return nil, fmt.Errorf("unmarshal response body failed: %v:%s", err, string(respBody))
		}
	}

	return &result, nil
}

// CreateCard 创建发票卡券模板
// req: 创建发票卡券模板请求参数
func (c *Client) CreateCard(req *CreateCardRequest) (*CreateCardResult, error) {
	// 构造请求URL
	path := "/card/invoice/platform/createcard"

	// 发送请求
	var result CreateCardResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// InsertInvoice 将电子发票卡券插入用户卡包
// req: 将电子发票卡券插入用户卡包请求参数
func (c *Client) InsertInvoice(req *InsertInvoiceRequest) (*InsertInvoiceResult, error) {
	// 构造请求URL
	path := "/card/invoice/insert"

	// 发送请求
	var result InsertInvoiceResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
