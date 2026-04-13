package channels

import "context"

// ---------- 订单管理 ----------

type OrderInfo struct {
	OrderID    string `json:"order_id"`
	ProductID  string `json:"product_id"`
	Status     int    `json:"status"`
	CreateTime int64  `json:"create_time"`
	UpdateTime int64  `json:"update_time"`
}

type GetOrderReq struct {
	OrderID string `json:"order_id"`
}
type GetOrderResp struct {
	Order OrderInfo `json:"order"`
}

type ListOrderReq struct {
	Status    *int  `json:"status,omitempty"`
	StartTime int64 `json:"start_time,omitempty"`
	EndTime   int64 `json:"end_time,omitempty"`
	Offset    *int  `json:"offset,omitempty"`
	Limit     *int  `json:"limit,omitempty"`
}
type ListOrderResp struct {
	Orders []OrderInfo `json:"orders"`
	Total  int         `json:"total"`
}

// GetOrder 获取订单详情
func (c *Client) GetOrder(ctx context.Context, req *GetOrderReq) (*GetOrderResp, error) {
	var resp GetOrderResp
	if err := c.doPost(ctx, "/channels/ec/order/get", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListOrder 获取订单列表
func (c *Client) ListOrder(ctx context.Context, req *ListOrderReq) (*ListOrderResp, error) {
	var resp ListOrderResp
	if err := c.doPost(ctx, "/channels/ec/order/list", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
