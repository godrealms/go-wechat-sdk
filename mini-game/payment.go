package mini_game

import (
	"context"
	"fmt"
)

// CreateOrderReq holds the parameters for creating a Mini Game virtual currency order.
type CreateOrderReq struct {
	OpenID    string `json:"openid"`
	Env       int    `json:"env"`
	ZoneID    string `json:"zone_id"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

// CreateOrderResp is the response returned by CreateOrder.
type CreateOrderResp struct {
	OrderID string `json:"order_id"`
	Balance int64  `json:"balance"`
}

// QueryOrderReq holds the parameters for querying an existing Mini Game order.
type QueryOrderReq struct {
	OrderID string `json:"order_id"`
	OpenID  string `json:"openid"`
}

// QueryOrderResp is the response returned by QueryOrder.
type QueryOrderResp struct {
	OrderID    string `json:"order_id"`
	Status     int    `json:"status"`
	PayAmount  int64  `json:"pay_amount"`
	CreateTime int64  `json:"create_time"`
}

// CreateOrder creates a new Mini Game virtual currency purchase order.
func (c *Client) CreateOrder(ctx context.Context, req *CreateOrderReq) (*CreateOrderResp, error) {
	if req == nil {
		return nil, fmt.Errorf("mini_game: CreateOrder: req is required")
	}
	if req.OpenID == "" {
		return nil, fmt.Errorf("mini_game: CreateOrder: req.OpenID is required")
	}
	if req.ProductID == "" {
		return nil, fmt.Errorf("mini_game: CreateOrder: req.ProductID is required")
	}
	if req.Quantity <= 0 {
		return nil, fmt.Errorf("mini_game: CreateOrder: req.Quantity must be > 0 (got %d)", req.Quantity)
	}
	var resp CreateOrderResp
	if err := c.doPost(ctx, "/wxa/game/createorder", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// QueryOrder retrieves the status and details of an existing Mini Game order.
func (c *Client) QueryOrder(ctx context.Context, req *QueryOrderReq) (*QueryOrderResp, error) {
	if req == nil {
		return nil, fmt.Errorf("mini_game: QueryOrder: req is required")
	}
	if req.OrderID == "" {
		return nil, fmt.Errorf("mini_game: QueryOrder: req.OrderID is required")
	}
	if req.OpenID == "" {
		return nil, fmt.Errorf("mini_game: QueryOrder: req.OpenID is required")
	}
	var resp QueryOrderResp
	if err := c.doPost(ctx, "/wxa/game/queryorder", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
