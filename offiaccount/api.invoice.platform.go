package offiaccount

import (
	"context"
	"fmt"
	"io"
)

// SetInvoiceUrl 设置商户联系方式（获取开票平台识别码）
func (c *Client) SetInvoiceUrl(ctx context.Context) (*SetUrlResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/card/invoice/seturl?access_token=%s", token)

	// 发送请求
	var result SetUrlResult
	if err = c.doPost(ctx, path, nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetPdf 获取pdf文件
// req: 获取pdf文件请求参数
func (c *Client) GetPdf(ctx context.Context, req *GetPdfRequest) (*GetPdfResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/card/invoice/platform/getpdf?access_token=%s", token)

	// 发送请求
	var result GetPdfResult
	if err = c.doPost(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateInvoiceStatus 更新发票状态
// req: 更新发票状态请求参数
func (c *Client) UpdateInvoiceStatus(ctx context.Context, req *UpdateInvoiceStatusRequest) (*Resp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/card/invoice/platform/updatestatus?access_token=%s", token)

	// 发送请求
	var result Resp
	if err = c.doPost(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// SetPdf 设置pdf文件
// pdf: pdf文件内容
//
// 走 doPostMultipartFile，以便自动进行 errcode 检查（历史实现自己拼 HTTP
// 请求并丢掉了 errcode，属于 audit 发现的同一类问题）。
func (c *Client) SetPdf(ctx context.Context, filename string, pdf io.Reader) (*SetPdfResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/card/invoice/platform/setpdf?access_token=%s", token)

	// 读入内存：helper 需要 []byte，PDF 文件通常在几十 KB ~ 数 MB 量级可接受。
	data, err := io.ReadAll(pdf)
	if err != nil {
		return nil, fmt.Errorf("offiaccount: SetPdf: read pdf: %w", err)
	}

	var result SetPdfResult
	if err = c.doPostMultipartFile(ctx, path, "pdf", filename, data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateCard 创建发票卡券模板
// req: 创建发票卡券模板请求参数
func (c *Client) CreateCard(ctx context.Context, req *CreateCardRequest) (*CreateCardResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/card/invoice/platform/createcard?access_token=%s", token)

	// 发送请求
	var result CreateCardResult
	if err = c.doPost(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// InsertInvoice 将电子发票卡券插入用户卡包
// req: 将电子发票卡券插入用户卡包请求参数
func (c *Client) InsertInvoice(ctx context.Context, req *InsertInvoiceRequest) (*InsertInvoiceResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/card/invoice/insert?access_token=%s", token)

	// 发送请求
	var result InsertInvoiceResult
	if err = c.doPost(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
