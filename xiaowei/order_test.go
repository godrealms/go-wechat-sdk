package xiaowei

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

func TestGetMicroOrder(t *testing.T) {
	srv := httptest.NewServer(apiHandler("/wxaapi/wxamicrostore/get_order",
		`{"errcode":0,"errmsg":"ok","order":{"order_id":"ORD001","status":2}}`))
	defer srv.Close()
	resp, err := newTestClient(t, srv.URL).GetMicroOrder(context.Background(), &GetMicroOrderReq{OrderID: "ORD001"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Order == nil {
		t.Error("expected non-nil Order")
	}
}

func TestListMicroOrders(t *testing.T) {
	srv := httptest.NewServer(apiHandler("/wxaapi/wxamicrostore/get_order_list",
		`{"errcode":0,"errmsg":"ok","order_list":[{"order_id":"ORD001"}],"total_num":1}`))
	defer srv.Close()
	resp, err := newTestClient(t, srv.URL).ListMicroOrders(context.Background(), &ListMicroOrdersReq{PageSize: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Orders) != 1 {
		t.Errorf("expected 1 order, got %d", len(resp.Orders))
	}
}

func TestShipMicroOrder(t *testing.T) {
	srv := httptest.NewServer(apiHandler("/wxaapi/wxamicrostore/ship_order",
		`{"errcode":0,"errmsg":"ok"}`))
	defer srv.Close()
	req := &ShipMicroOrderReq{OrderID: "ORD001", DeliveryCompany: "SF", TrackingNumber: "SF123"}
	if err := newTestClient(t, srv.URL).ShipMicroOrder(context.Background(), req); err != nil {
		t.Fatal(err)
	}
}

func TestRefundMicroOrder(t *testing.T) {
	srv := httptest.NewServer(apiHandler("/wxaapi/wxamicrostore/refund_order",
		`{"errcode":0,"errmsg":"ok"}`))
	defer srv.Close()
	req := &RefundMicroOrderReq{OrderID: "ORD001", RefundAmount: 0}
	if err := newTestClient(t, srv.URL).RefundMicroOrder(context.Background(), req); err != nil {
		t.Fatal(err)
	}
}

func TestGetMicroOrder_NetworkError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close() // close immediately to force a network error
	fake := &fakeTokenSource{token: "TOK"}
	c, err := NewClient(Config{AppId: "wx", AppSecret: "sec"},
		WithHTTP(utils.NewHTTP(srv.URL, utils.WithTimeout(3*time.Second))),
		WithTokenSource(fake))
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.GetMicroOrder(context.Background(), &GetMicroOrderReq{OrderID: "ORD001"})
	if err == nil {
		t.Error("expected network error, got nil")
	}
}
