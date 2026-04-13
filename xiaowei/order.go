package xiaowei

import "context"

// MicroOrder represents a Xiaowei order.
type MicroOrder struct {
	OrderID string `json:"order_id,omitempty"`
	Status  int    `json:"status,omitempty"`
	Amount  int64  `json:"amount,omitempty"` // total in fen
}

// GetMicroOrderReq is the request to get an order.
type GetMicroOrderReq struct {
	OrderID string `json:"order_id"`
}

// GetMicroOrderResp is the response from GetMicroOrder.
type GetMicroOrderResp struct {
	Order *MicroOrder `json:"order"`
}

// GetMicroOrder returns the details of a Xiaowei order.
func (c *Client) GetMicroOrder(ctx context.Context, req *GetMicroOrderReq) (*GetMicroOrderResp, error) {
	var resp GetMicroOrderResp
	if err := c.doPost(ctx, "/wxaapi/wxamicrostore/get_order", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListMicroOrdersReq is the request to list orders.
type ListMicroOrdersReq struct {
	Status    int   `json:"status,omitempty"`
	Page      int   `json:"page,omitempty"`
	PageSize  int   `json:"page_size,omitempty"`
	StartTime int64 `json:"start_time,omitempty"`
	EndTime   int64 `json:"end_time,omitempty"`
}

// ListMicroOrdersResp is the response from ListMicroOrders.
type ListMicroOrdersResp struct {
	Orders   []*MicroOrder `json:"order_list"`
	TotalNum int           `json:"total_num"`
}

// ListMicroOrders returns a paginated list of Xiaowei orders.
func (c *Client) ListMicroOrders(ctx context.Context, req *ListMicroOrdersReq) (*ListMicroOrdersResp, error) {
	var resp ListMicroOrdersResp
	if err := c.doPost(ctx, "/wxaapi/wxamicrostore/get_order_list", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ShipMicroOrderReq is the request to mark an order as shipped.
type ShipMicroOrderReq struct {
	OrderID         string `json:"order_id"`
	DeliveryCompany string `json:"delivery_company"`
	TrackingNumber  string `json:"tracking_number"`
}

// ShipMicroOrder marks a Xiaowei order as shipped with tracking information.
func (c *Client) ShipMicroOrder(ctx context.Context, req *ShipMicroOrderReq) error {
	return c.doPost(ctx, "/wxaapi/wxamicrostore/ship_order", req, nil)
}

// RefundMicroOrderReq is the request to refund an order.
type RefundMicroOrderReq struct {
	OrderID      string `json:"order_id"`
	RefundAmount int64  `json:"refund_amount"` // amount in fen; 0 means full refund
	RefundReason string `json:"refund_reason,omitempty"`
}

// RefundMicroOrder initiates a refund for a Xiaowei order.
func (c *Client) RefundMicroOrder(ctx context.Context, req *RefundMicroOrderReq) error {
	return c.doPost(ctx, "/wxaapi/wxamicrostore/refund_order", req, nil)
}
