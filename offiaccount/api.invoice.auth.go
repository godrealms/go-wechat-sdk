package offiaccount

import (
	"fmt"
)

// SetInvoiceBizAttr 设置授权页与商户信息
// action: 接口功能类型，可选值：set_auth_field、get_auth_field、set_pay_mch、get_pay_mch、set_contact、get_contact
// req: 请求参数
func (c *Client) SetInvoiceBizAttr(action string, req *SetBizAttrRequest) (*Resp, error) {
	// 构造请求URL
	path := fmt.Sprintf("/card/invoice/setbizattr?action=%s", action)

	// 发送请求
	var result Resp
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetAuthData 获取授权页数据
// req: 获取授权页数据请求参数
func (c *Client) GetAuthData(req *GetAuthDataRequest) (*GetAuthDataResult, error) {
	// 构造请求URL
	path := "/card/invoice/getauthdata"

	// 发送请求
	var result GetAuthDataResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetInvoiceTicket 获取授权页ticket
func (c *Client) GetInvoiceTicket() (*Ticket, error) {
	// 构造请求URL
	path := "/cgi-bin/ticket/getticket?type=wx_card"

	// 发送请求
	var result Ticket
	err := c.Https.Get(c.ctx, path, nil, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetAuthUrl 获取授权页链接
// req: 获取授权页链接请求参数
func (c *Client) GetAuthUrl(req *GetAuthUrlRequest) (*GetAuthUrlResult, error) {
	// 构造请求URL
	path := "/card/invoice/getauthurl"

	// 发送请求
	var result GetAuthUrlResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// RejectInsert 拒绝领取发票
// req: 拒绝领取发票请求参数
func (c *Client) RejectInsert(req *RejectInsertRequest) (*Resp, error) {
	// 构造请求URL
	path := "/card/invoice/rejectinsert"

	// 发送请求
	var result Resp
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
