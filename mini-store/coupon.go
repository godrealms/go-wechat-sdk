package mini_store

import "context"

// Coupon represents a Mini Store coupon/promotion.
type Coupon struct {
	CouponID  string `json:"coupon_id,omitempty"`
	Name      string `json:"coupon_name,omitempty"`
	Type      int    `json:"coupon_type,omitempty"` // 1=fixed-amount, 2=percentage
	Discount  int64  `json:"discount,omitempty"`    // in fen or basis points
	MinAmount int64  `json:"min_amount,omitempty"`  // minimum order value in fen
}

// AddCouponResp is the response from AddCoupon.
type AddCouponResp struct {
	CouponID string `json:"coupon_id"`
}

// AddCoupon creates a new coupon and returns its coupon_id.
func (c *Client) AddCoupon(ctx context.Context, coupon *Coupon) (*AddCouponResp, error) {
	var resp AddCouponResp
	if err := c.doPost(ctx, "/shop/coupon/add", coupon, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCouponReq is the request to get coupon details.
type GetCouponReq struct {
	CouponID string `json:"coupon_id"`
}

// GetCouponResp is the response from GetCoupon.
type GetCouponResp struct {
	Coupon *Coupon `json:"coupon"`
}

// GetCoupon returns the details of a coupon.
func (c *Client) GetCoupon(ctx context.Context, req *GetCouponReq) (*GetCouponResp, error) {
	var resp GetCouponResp
	if err := c.doPost(ctx, "/shop/coupon/get", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateCouponStatusReq changes a coupon's active/inactive status.
type UpdateCouponStatusReq struct {
	CouponID string `json:"coupon_id"`
	Status   int    `json:"status"` // 0=inactive, 1=active
}

// UpdateCouponStatus activates or deactivates a coupon.
func (c *Client) UpdateCouponStatus(ctx context.Context, req *UpdateCouponStatusReq) error {
	return c.doPost(ctx, "/shop/coupon/update_status", req, nil)
}

// ListCouponsReq is the request to list coupons.
type ListCouponsReq struct {
	Status   int `json:"status,omitempty"`
	Page     int `json:"page,omitempty"`
	PageSize int `json:"page_size,omitempty"`
}

// ListCouponsResp is the response from ListCoupons.
type ListCouponsResp struct {
	Coupons  []*Coupon `json:"coupon_list"`
	TotalNum int       `json:"total_num"`
}

// ListCoupons returns a paginated list of coupons.
func (c *Client) ListCoupons(ctx context.Context, req *ListCouponsReq) (*ListCouponsResp, error) {
	var resp ListCouponsResp
	if err := c.doPost(ctx, "/shop/coupon/get_list", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
