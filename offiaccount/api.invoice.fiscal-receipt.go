package offiaccount

import (
	"context"
	"fmt"
	"io"
	"net/url"
)

// GetFiscalAuthData 查询财政电子票据授权信息
// https://developers.weixin.qq.com/doc/service/api/invoice/FiscalReceipt/api_invoicebizgetauthdata.html
func (c *Client) GetFiscalAuthData(ctx context.Context, request *GetFiscalAuthDataRequest) (*GetFiscalAuthDataResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	result := &GetFiscalAuthDataResult{}
	path := fmt.Sprintf("/card/invoice/getauthdata?access_token=%s", token)
	if err := c.doPost(ctx, path, request, result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetTicket 获取sdk临时票据
// https://developers.weixin.qq.com/doc/service/api/invoice/FiscalReceipt/api_getticket.html
func (c *Client) GetTicket(ctx context.Context) (*GetTicketResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	result := &GetTicketResult{}
	query := url.Values{
		"access_token": {token},
		"type":         {"wx_card"},
	}
	if err := c.doGet(ctx, "/cgi-bin/ticket/getticket", query, result); err != nil {
		return nil, err
	}
	return result, nil
}

// RejectInsertFiscal 拒绝开票
// https://developers.weixin.qq.com/doc/service/api/invoice/FiscalReceipt/api_invoicebizrejectinsert.html
func (c *Client) RejectInsertFiscal(ctx context.Context, request *RejectInsertFiscalRequest) (*Resp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	result := &Resp{}
	path := fmt.Sprintf("/card/invoice/rejectinsert?access_token=%s", token)
	if err := c.doPost(ctx, path, request, result); err != nil {
		return nil, err
	}
	return result, nil
}

// SetFiscalInvoiceUrl 设置财政电子票据授权页链接
// https://developers.weixin.qq.com/doc/service/api/invoice/FiscalReceipt/api_setinvoiceurl.html
func (c *Client) SetFiscalInvoiceUrl(ctx context.Context) (*SetInvoiceUrlResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	result := &SetInvoiceUrlResult{}
	body := map[string]string{}
	path := fmt.Sprintf("/card/invoice/seturl?access_token=%s", token)
	if err := c.doPost(ctx, path, body, result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetPlatformPdf 查询已上传的PDF文件
// https://developers.weixin.qq.com/doc/service/api/invoice/FiscalReceipt/api_invoiceplatformgetpdf.html
func (c *Client) GetPlatformPdf(ctx context.Context, request *GetPlatformPdfRequest) (*GetPlatformPdfResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	result := &GetPlatformPdfResult{}
	path := fmt.Sprintf("/card/invoice/platform/getpdf?access_token=%s", token)
	if err := c.doPost(ctx, path, request, result); err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateInvoicePlatformStatus 更新发票状态
// https://developers.weixin.qq.com/doc/service/api/invoice/FiscalReceipt/api_invoicekpupdatainvoicestatus.html
func (c *Client) UpdateInvoicePlatformStatus(ctx context.Context, request *UpdateInvoicePlatformStatusRequest) (*Resp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	result := &Resp{}
	path := fmt.Sprintf("/card/invoice/platform/updatestatus?access_token=%s", token)
	if err := c.doPost(ctx, path, request, result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetFiscalAuthUrl 获取授权页链接
// https://developers.weixin.qq.com/doc/service/api/invoice/FiscalReceipt/api_notaxinvoicegetauthurl.html
func (c *Client) GetFiscalAuthUrl(ctx context.Context, request *GetFiscalAuthUrlRequest) (*GetFiscalAuthUrlResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	result := &GetFiscalAuthUrlResult{}
	path := fmt.Sprintf("/nontax/getbillauthurl?access_token=%s", token)
	if err := c.doPost(ctx, path, request, result); err != nil {
		return nil, err
	}
	return result, nil
}

// CreateFiscalCard 创建财政电子票据模板
// https://developers.weixin.qq.com/doc/service/api/invoice/FiscalReceipt/api_notaxinvoicecreatecard.html
func (c *Client) CreateFiscalCard(ctx context.Context, request *CreateFiscalCardRequest) (*CreateFiscalCardResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	result := &CreateFiscalCardResult{}
	path := fmt.Sprintf("/nontax/createbillcard?access_token=%s", token)
	if err := c.doPost(ctx, path, request, result); err != nil {
		return nil, err
	}
	return result, nil
}

// InsertFiscalInvoice 票据插入用户卡包
// https://developers.weixin.qq.com/doc/service/api/invoice/FiscalReceipt/api_notaxinvoiceinsert.html
func (c *Client) InsertFiscalInvoice(ctx context.Context, request *InsertFiscalInvoiceRequest) (*InsertInvoiceResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	result := &InsertInvoiceResult{}
	path := fmt.Sprintf("/nontax/insertbill?access_token=%s", token)
	if err := c.doPost(ctx, path, request, result); err != nil {
		return nil, err
	}
	return result, nil
}

// SetPlatformPdf 上传发票PDF
//
// 此接口需要 multipart/form-data 上传 PDF 文件，直接转发给 SetPdf。
// https://developers.weixin.qq.com/doc/service/api/invoice/FiscalReceipt/api_invoiceplatformsetpdf.html
func (c *Client) SetPlatformPdf(ctx context.Context, filename string, pdf io.Reader) (*SetPdfResult, error) {
	return c.SetPdf(ctx, filename, pdf)
}
