package offiaccount

import (
	"context"
	"fmt"
	"net/url"
)

// SetInvoiceBizAttr 设置授权页与商户信息
// action: 接口功能类型，可选值：set_auth_field、get_auth_field、set_pay_mch、get_pay_mch、set_contact、get_contact
// req: 请求参数
func (c *Client) SetInvoiceBizAttr(ctx context.Context, action string, req *SetBizAttrRequest) (*Resp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/card/invoice/setbizattr?access_token=%s&action=%s", token, action)

	// 发送请求
	var result Resp
	if err := c.Https.Post(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetAuthData 获取授权页数据
// req: 获取授权页数据请求参数
func (c *Client) GetAuthData(ctx context.Context, req *GetAuthDataRequest) (*GetAuthDataResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/card/invoice/getauthdata?access_token=%s", token)

	// 发送请求
	var result GetAuthDataResult
	if err := c.Https.Post(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetInvoiceTicket 获取授权页ticket
func (c *Client) GetInvoiceTicket(ctx context.Context) (*Ticket, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := "/cgi-bin/ticket/getticket"
	params := url.Values{
		"access_token": {token},
		"type":         {"wx_card"},
	}

	// 发送请求
	var result Ticket
	if err := c.Https.Get(ctx, path, params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetAuthUrl 获取授权页链接
// req: 获取授权页链接请求参数
func (c *Client) GetAuthUrl(ctx context.Context, req *GetAuthUrlRequest) (*GetAuthUrlResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/card/invoice/getauthurl?access_token=%s", token)

	// 发送请求
	var result GetAuthUrlResult
	if err := c.Https.Post(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// RejectInsert 拒绝领取发票
// req: 拒绝领取发票请求参数
func (c *Client) RejectInsert(ctx context.Context, req *RejectInsertRequest) (*Resp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/card/invoice/rejectinsert?access_token=%s", token)

	// 发送请求
	var result Resp
	if err := c.Https.Post(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
