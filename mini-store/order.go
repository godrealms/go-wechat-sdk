package mini_store

import (
	"context"
	"fmt"
)

// Order represents a Mini Store order.
type Order struct {
	OrderID    string `json:"order_id,omitempty"`
	Status     int    `json:"status,omitempty"`
	UserOpenID string `json:"openid,omitempty"`
}

// GetOrderReq is the request to get a single order.
type GetOrderReq struct {
	OrderID string `json:"order_id"`
}

// GetOrderResp is the response from GetOrder.
type GetOrderResp struct {
	Order *Order `json:"order"`
}

// GetOrder returns the details of a single order by order_id.
func (c *Client) GetOrder(ctx context.Context, req *GetOrderReq) (*GetOrderResp, error) {
	var resp GetOrderResp
	if err := c.doPost(ctx, "/shop/order/get", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListOrdersReq is the request to list orders.
type ListOrdersReq struct {
	Status    int   `json:"status,omitempty"`
	Page      int   `json:"page,omitempty"`
	PageSize  int   `json:"page_size,omitempty"`
	StartTime int64 `json:"start_time,omitempty"`
	EndTime   int64 `json:"end_time,omitempty"`
}

// ListOrdersResp is the response from ListOrders.
type ListOrdersResp struct {
	Orders   []*Order `json:"order_list"`
	TotalNum int      `json:"total_num"`
}

// ListOrders returns a paginated list of orders, optionally filtered by status and time range.
func (c *Client) ListOrders(ctx context.Context, req *ListOrdersReq) (*ListOrdersResp, error) {
	var resp ListOrdersResp
	if err := c.doPost(ctx, "/shop/order/get_list", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateOrderPriceReq is the request to modify an order price.
type UpdateOrderPriceReq struct {
	OrderID  string `json:"order_id"`
	NewPrice int64  `json:"new_price"` // in fen (1/100 CNY)
}

// UpdateOrderPrice modifies the total price of a pending order before payment.
func (c *Client) UpdateOrderPrice(ctx context.Context, req *UpdateOrderPriceReq) error {
	if req == nil {
		return fmt.Errorf("mini_store: UpdateOrderPrice: req is required")
	}
	if req.OrderID == "" {
		return fmt.Errorf("mini_store: UpdateOrderPrice: req.OrderID is required")
	}
	if req.NewPrice <= 0 {
		return fmt.Errorf("mini_store: UpdateOrderPrice: req.NewPrice must be > 0 (got %d)", req.NewPrice)
	}
	return c.doPost(ctx, "/shop/order/update_price", req, nil)
}

// CloseOrderReq is the request to close an open order.
type CloseOrderReq struct {
	OrderID string `json:"order_id"`
}

// CloseOrder closes an open order, preventing payment.
func (c *Client) CloseOrder(ctx context.Context, req *CloseOrderReq) error {
	if req == nil || req.OrderID == "" {
		return fmt.Errorf("mini_store: CloseOrder: req.OrderID is required")
	}
	return c.doPost(ctx, "/shop/order/close", req, nil)
}

// UploadShippingReq is the request to upload shipping information.
type UploadShippingReq struct {
	OrderID         string `json:"order_id"`
	DeliveryCompany string `json:"delivery_company"`
	DeliverySN      string `json:"delivery_sn"`
}

// UploadShipping records delivery tracking information for an order.
func (c *Client) UploadShipping(ctx context.Context, req *UploadShippingReq) error {
	if req == nil {
		return fmt.Errorf("mini_store: UploadShipping: req is required")
	}
	if req.OrderID == "" {
		return fmt.Errorf("mini_store: UploadShipping: req.OrderID is required")
	}
	if req.DeliveryCompany == "" {
		return fmt.Errorf("mini_store: UploadShipping: req.DeliveryCompany is required")
	}
	if req.DeliverySN == "" {
		return fmt.Errorf("mini_store: UploadShipping: req.DeliverySN is required")
	}
	return c.doPost(ctx, "/shop/delivery/send", req, nil)
}

// GetAfterSaleOrderReq is the request to get an after-sale order.
type GetAfterSaleOrderReq struct {
	AfterSaleOrderID string `json:"after_sale_order_id"`
}

// AfterSaleOrderDetail holds the details of an after-sale (refund/return) order.
type AfterSaleOrderDetail struct {
	AfterSaleOrderID string `json:"after_sale_order_id,omitempty"`
	OrderID          string `json:"order_id,omitempty"`
	RefundAmount     int64  `json:"refund_amount,omitempty"`
	Status           int    `json:"status,omitempty"`
	Reason           string `json:"reason,omitempty"`
}

// GetAfterSaleOrderResp is the response from GetAfterSaleOrder.
type GetAfterSaleOrderResp struct {
	AfterSaleOrder *AfterSaleOrderDetail `json:"after_sale_order"`
}

// GetAfterSaleOrder returns the details of an after-sale (refund/return) order.
func (c *Client) GetAfterSaleOrder(ctx context.Context, req *GetAfterSaleOrderReq) (*GetAfterSaleOrderResp, error) {
	if req == nil || req.AfterSaleOrderID == "" {
		return nil, fmt.Errorf("mini_store: GetAfterSaleOrder: req.AfterSaleOrderID is required")
	}
	var resp GetAfterSaleOrderResp
	if err := c.doPost(ctx, "/shop/aftersale/get_after_sale_order", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AcceptRefundReq is the request to accept a refund.
type AcceptRefundReq struct {
	AfterSaleOrderID string `json:"after_sale_order_id"`
}

// AcceptRefund approves a customer refund request.
func (c *Client) AcceptRefund(ctx context.Context, req *AcceptRefundReq) error {
	if req == nil || req.AfterSaleOrderID == "" {
		return fmt.Errorf("mini_store: AcceptRefund: req.AfterSaleOrderID is required")
	}
	return c.doPost(ctx, "/shop/aftersale/accept_refund", req, nil)
}

// RejectRefundReq is the request to reject a refund.
type RejectRefundReq struct {
	AfterSaleOrderID string `json:"after_sale_order_id"`
	RejectReason     string `json:"reject_reason,omitempty"`
}

// RejectRefund declines a customer refund request with an optional reason.
func (c *Client) RejectRefund(ctx context.Context, req *RejectRefundReq) error {
	if req == nil || req.AfterSaleOrderID == "" {
		return fmt.Errorf("mini_store: RejectRefund: req.AfterSaleOrderID is required")
	}
	return c.doPost(ctx, "/shop/aftersale/reject_refund", req, nil)
}
