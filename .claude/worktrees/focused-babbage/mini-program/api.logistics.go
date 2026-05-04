package mini_program

import (
	"fmt"
	"net/url"
)

// AddExpressOrder 生成运单
// POST /cgi-bin/express/business/order/add
func (c *Client) AddExpressOrder(req *AddExpressOrderRequest) (*AddExpressOrderResult, error) {
	path := fmt.Sprintf("/cgi-bin/express/business/order/add?access_token=%s", c.GetAccessToken())
	result := &AddExpressOrderResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}

// GetExpressOrder 查询运单
// POST /cgi-bin/express/business/order/get
func (c *Client) GetExpressOrder(orderId, openId, deliveryId, waybillId string) (*GetExpressOrderResult, error) {
	path := fmt.Sprintf("/cgi-bin/express/business/order/get?access_token=%s", c.GetAccessToken())
	body := map[string]string{
		"order_id":    orderId,
		"openid":      openId,
		"delivery_id": deliveryId,
		"waybill_id":  waybillId,
	}
	result := &GetExpressOrderResult{}
	err := c.Https.Post(c.Ctx, path, body, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}

// CancelExpressOrder 取消运单
// POST /cgi-bin/express/business/order/cancel
func (c *Client) CancelExpressOrder(req *CancelExpressOrderRequest) (*CancelExpressOrderResult, error) {
	path := fmt.Sprintf("/cgi-bin/express/business/order/cancel?access_token=%s", c.GetAccessToken())
	result := &CancelExpressOrderResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}

// GetAllDelivery 获取支持的快递公司列表
// GET /cgi-bin/express/business/delivery/getall
func (c *Client) GetAllDelivery() (*GetAllDeliveryResult, error) {
	query := c.TokenQuery(url.Values{})
	result := &GetAllDeliveryResult{}
	err := c.Https.Get(c.Ctx, "/cgi-bin/express/business/delivery/getall", query, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}
