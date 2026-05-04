package mini_store

import (
	"context"
	"strings"
	"testing"
)

// TestValidation_NetworkNotReached verifies the H8 input validations short-
// circuit the request before any HTTP call. We pass nil/empty inputs to each
// money-adjacent method and assert it returns an error whose text mentions
// "is required" or "must be" — both indicators that the inline validation
// fired (versus an opaque server error from a wasted round-trip).
func TestValidation_NetworkNotReached(t *testing.T) {
	c := &Client{} // no http; if validation passes, the call would NPE
	ctx := context.Background()

	tests := []struct {
		name    string
		fn      func() error
		mustSay string
	}{
		{"UpdateOrderPrice nil", func() error { return c.UpdateOrderPrice(ctx, nil) }, "req is required"},
		{"UpdateOrderPrice empty OrderID", func() error {
			return c.UpdateOrderPrice(ctx, &UpdateOrderPriceReq{NewPrice: 100})
		}, "OrderID is required"},
		{"UpdateOrderPrice zero NewPrice", func() error {
			return c.UpdateOrderPrice(ctx, &UpdateOrderPriceReq{OrderID: "X", NewPrice: 0})
		}, "must be > 0"},
		{"UpdateOrderPrice negative NewPrice", func() error {
			return c.UpdateOrderPrice(ctx, &UpdateOrderPriceReq{OrderID: "X", NewPrice: -1})
		}, "must be > 0"},
		{"CloseOrder nil", func() error { return c.CloseOrder(ctx, nil) }, "OrderID is required"},
		{"CloseOrder empty OrderID", func() error { return c.CloseOrder(ctx, &CloseOrderReq{}) }, "OrderID is required"},
		{"UploadShipping nil", func() error { return c.UploadShipping(ctx, nil) }, "req is required"},
		{"UploadShipping empty OrderID", func() error {
			return c.UploadShipping(ctx, &UploadShippingReq{DeliveryCompany: "X", DeliverySN: "Y"})
		}, "OrderID is required"},
		{"UploadShipping empty DeliveryCompany", func() error {
			return c.UploadShipping(ctx, &UploadShippingReq{OrderID: "X", DeliverySN: "Y"})
		}, "DeliveryCompany is required"},
		{"UploadShipping empty DeliverySN", func() error {
			return c.UploadShipping(ctx, &UploadShippingReq{OrderID: "X", DeliveryCompany: "Y"})
		}, "DeliverySN is required"},
		{"GetAfterSaleOrder nil", func() error { _, err := c.GetAfterSaleOrder(ctx, nil); return err }, "AfterSaleOrderID"},
		{"AcceptRefund nil", func() error { return c.AcceptRefund(ctx, nil) }, "AfterSaleOrderID"},
		{"RejectRefund nil", func() error { return c.RejectRefund(ctx, nil) }, "AfterSaleOrderID"},
		{"AddCoupon nil", func() error { _, err := c.AddCoupon(ctx, nil); return err }, "coupon is required"},
		{"AddCoupon missing Type", func() error {
			_, err := c.AddCoupon(ctx, &Coupon{Name: "X"})
			return err
		}, "Type must be"},
		{"AddCoupon zero Discount", func() error {
			_, err := c.AddCoupon(ctx, &Coupon{Name: "X", Type: 1})
			return err
		}, "Discount must be > 0"},
		{"GetCoupon nil", func() error { _, err := c.GetCoupon(ctx, nil); return err }, "CouponID"},
		{"UpdateCouponStatus nil", func() error { return c.UpdateCouponStatus(ctx, nil) }, "CouponID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			if err == nil {
				t.Fatalf("expected validation error, got nil")
			}
			if !strings.Contains(err.Error(), tt.mustSay) {
				t.Errorf("error %q should contain %q", err.Error(), tt.mustSay)
			}
		})
	}
}
