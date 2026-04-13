package channels

import "context"

// OrderInfo contains the details of a single Channels e-commerce order.
type OrderInfo struct {
	OrderID    string `json:"order_id"`
	ProductID  string `json:"product_id"`
	Status     int    `json:"status"`
	CreateTime int64  `json:"create_time"`
	UpdateTime int64  `json:"update_time"`
}

// GetOrderReq holds the order ID for querying a single Channels order.
type GetOrderReq struct {
	OrderID string `json:"order_id"`
}

// GetOrderResp is the response returned by GetOrder.
type GetOrderResp struct {
	Order OrderInfo `json:"order"`
}

// ListOrderReq holds the filter and pagination parameters for listing Channels orders.
type ListOrderReq struct {
	Status    *int  `json:"status,omitempty"`
	StartTime int64 `json:"start_time,omitempty"`
	EndTime   int64 `json:"end_time,omitempty"`
	Offset    *int  `json:"offset,omitempty"`
	Limit     *int  `json:"limit,omitempty"`
}

// ListOrderResp is the response returned by ListOrder.
type ListOrderResp struct {
	Orders []OrderInfo `json:"orders"`
	Total  int         `json:"total"`
}

// GetOrder retrieves the details of the specified Channels e-commerce order.
func (c *Client) GetOrder(ctx context.Context, req *GetOrderReq) (*GetOrderResp, error) {
	var resp GetOrderResp
	if err := c.doPost(ctx, "/channels/ec/order/get", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListOrder retrieves a paginated list of Channels e-commerce orders matching the given filters.
func (c *Client) ListOrder(ctx context.Context, req *ListOrderReq) (*ListOrderResp, error) {
	var resp ListOrderResp
	if err := c.doPost(ctx, "/channels/ec/order/list", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
