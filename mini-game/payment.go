package mini_game

import "context"

type CreateOrderReq struct {
	OpenID    string `json:"openid"`
	Env       int    `json:"env"`
	ZoneID    string `json:"zone_id"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}
type CreateOrderResp struct {
	OrderID string `json:"order_id"`
	Balance int64  `json:"balance"`
}

type QueryOrderReq struct {
	OrderID string `json:"order_id"`
	OpenID  string `json:"openid"`
}
type QueryOrderResp struct {
	OrderID    string `json:"order_id"`
	Status     int    `json:"status"`
	PayAmount  int64  `json:"pay_amount"`
	CreateTime int64  `json:"create_time"`
}

func (c *Client) CreateOrder(ctx context.Context, req *CreateOrderReq) (*CreateOrderResp, error) {
	var resp CreateOrderResp
	if err := c.doPost(ctx, "/wxa/game/createorder", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) QueryOrder(ctx context.Context, req *QueryOrderReq) (*QueryOrderResp, error) {
	var resp QueryOrderResp
	if err := c.doPost(ctx, "/wxa/game/queryorder", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
