package mini_store

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

func TestAddCoupon(t *testing.T) {
	srv := apiServer(t, "/shop/coupon/add", `{"errcode":0,"errmsg":"ok","coupon_id":"CPN001"}`)
	defer srv.Close()
	resp, err := newTestClient(t, srv.URL).AddCoupon(ctx(t), &Coupon{Name: "10 off", Type: 1, Discount: 1000})
	if err != nil {
		t.Fatal(err)
	}
	if resp.CouponID != "CPN001" {
		t.Errorf("got coupon_id=%q", resp.CouponID)
	}
}

func TestGetCoupon(t *testing.T) {
	srv := apiServer(t, "/shop/coupon/get", `{"errcode":0,"errmsg":"ok","coupon":{"coupon_id":"CPN001","coupon_name":"10 off"}}`)
	defer srv.Close()
	resp, err := newTestClient(t, srv.URL).GetCoupon(ctx(t), &GetCouponReq{CouponID: "CPN001"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Coupon == nil {
		t.Error("expected non-nil Coupon")
	}
}

func TestUpdateCouponStatus(t *testing.T) {
	srv := apiServer(t, "/shop/coupon/update_status", `{"errcode":0,"errmsg":"ok"}`)
	defer srv.Close()
	if err := newTestClient(t, srv.URL).UpdateCouponStatus(ctx(t), &UpdateCouponStatusReq{CouponID: "CPN001", Status: 1}); err != nil {
		t.Fatal(err)
	}
}

func TestListCoupons(t *testing.T) {
	srv := apiServer(t, "/shop/coupon/get_list", `{"errcode":0,"errmsg":"ok","coupon_list":[{"coupon_id":"CPN001"}],"total_num":1}`)
	defer srv.Close()
	resp, err := newTestClient(t, srv.URL).ListCoupons(ctx(t), &ListCouponsReq{PageSize: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Coupons) != 1 {
		t.Errorf("expected 1 coupon, got %d", len(resp.Coupons))
	}
}

func TestAddCoupon_NetworkError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close() // close immediately to force a network error
	fake := &fakeTokenSource{token: "TOK"}
	c, err := NewClient(Config{AppId: "wx", AppSecret: "sec"},
		WithHTTP(utils.NewHTTP(srv.URL, utils.WithTimeout(3*time.Second))),
		WithTokenSource(fake))
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.AddCoupon(context.Background(), &Coupon{Name: "10 off", Type: 1, Discount: 1000})
	if err == nil {
		t.Error("expected network error, got nil")
	}
}
