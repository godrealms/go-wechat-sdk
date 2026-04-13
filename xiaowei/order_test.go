package xiaowei

import (
	"context"
	"net/http/httptest"
	"testing"
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
