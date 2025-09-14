package offiaccount

import (
	"net/url"
)

// GetFiscalAuthData 查询财政电子票据授权信息
// https://developers.weixin.qq.com/doc/service/api/invoice/FiscalReceipt/api_invoicebizgetauthdata.html
func (c *Client) GetFiscalAuthData(request *GetFiscalAuthDataRequest) (*GetFiscalAuthDataResult, error) {
	result := &GetFiscalAuthDataResult{}
	err := c.Https.Post(c.ctx, "/card/invoice/getauthdata", request, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetTicket 获取sdk临时票据
// https://developers.weixin.qq.com/doc/service/api/invoice/FiscalReceipt/api_getticket.html
func (c *Client) GetTicket() (*GetTicketResult, error) {
	result := &GetTicketResult{}
	query := url.Values{"type": {"wx_card"}}
	err := c.Https.Get(c.ctx, "/cgi-bin/ticket/getticket", query, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// RejectInsertFiscal 拒绝开票
// https://developers.weixin.qq.com/doc/service/api/invoice/FiscalReceipt/api_invoicebizrejectinsert.html
func (c *Client) RejectInsertFiscal(request *RejectInsertFiscalRequest) (*Resp, error) {
	result := &Resp{}
	err := c.Https.Post(c.ctx, "/card/invoice/rejectinsert", request, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// SetFiscalInvoiceUrl 设置财政电子票据授权页链接
// https://developers.weixin.qq.com/doc/service/api/invoice/FiscalReceipt/api_setinvoiceurl.html
func (c *Client) SetFiscalInvoiceUrl() (*SetInvoiceUrlResult, error) {
	result := &SetInvoiceUrlResult{}
	body := map[string]string{}
	err := c.Https.Post(c.ctx, "/card/invoice/seturl", body, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetPlatformPdf 查询已上传的PDF文件
// https://developers.weixin.qq.com/doc/service/api/invoice/FiscalReceipt/api_invoiceplatformgetpdf.html
func (c *Client) GetPlatformPdf(request *GetPlatformPdfRequest) (*GetPlatformPdfResult, error) {
	result := &GetPlatformPdfResult{}
	err := c.Https.Post(c.ctx, "/card/invoice/platform/getpdf", request, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateInvoicePlatformStatus 更新发票状态
// https://developers.weixin.qq.com/doc/service/api/invoice/FiscalReceipt/api_invoicekpupdatainvoicestatus.html
func (c *Client) UpdateInvoicePlatformStatus(request *UpdateInvoicePlatformStatusRequest) (*Resp, error) {
	result := &Resp{}
	err := c.Https.Post(c.ctx, "/card/invoice/platform/updatestatus", request, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetFiscalAuthUrl 获取授权页链接
// https://developers.weixin.qq.com/doc/service/api/invoice/FiscalReceipt/api_notaxinvoicegetauthurl.html
func (c *Client) GetFiscalAuthUrl(request *GetFiscalAuthUrlRequest) (*GetFiscalAuthUrlResult, error) {
	result := &GetFiscalAuthUrlResult{}
	err := c.Https.Post(c.ctx, "/nontax/getbillauthurl", request, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// CreateFiscalCard 创建财政电子票据模板
// https://developers.weixin.qq.com/doc/service/api/invoice/FiscalReceipt/api_notaxinvoicecreatecard.html
func (c *Client) CreateFiscalCard(request *CreateFiscalCardRequest) (*CreateFiscalCardResult, error) {
	result := &CreateFiscalCardResult{}
	err := c.Https.Post(c.ctx, "/nontax/createbillcard", request, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// InsertFiscalInvoice 票据插入用户卡包
// https://developers.weixin.qq.com/doc/service/api/invoice/FiscalReceipt/api_notaxinvoiceinsert.html
func (c *Client) InsertFiscalInvoice(request *InsertFiscalInvoiceRequest) (*InsertInvoiceResult, error) {
	result := &InsertInvoiceResult{}
	err := c.Https.Post(c.ctx, "/nontax/insertbill", request, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// SetPlatformPdf 上传发票PDF
// https://developers.weixin.qq.com/doc/service/api/invoice/FiscalReceipt/api_invoiceplatformsetpdf.html
func (c *Client) SetPlatformPdf(pdfFilePath string) (*SetPdfResult, error) {
	result := &SetPdfResult{}
	// 使用通用的Post方法发送文件
	err := c.Https.Post(c.ctx, "/card/invoice/platform/setpdf", pdfFilePath, result)
	return result, err
}
