package mini_store

import (
	"testing"
)

func TestGetOrder(t *testing.T) {
	srv := apiServer(t, "/shop/order/get", `{"errcode":0,"errmsg":"ok","order":{"order_id":"ORD001","status":1}}`)
	defer srv.Close()
	resp, err := newTestClient(t, srv.URL).GetOrder(ctx(t), &GetOrderReq{OrderID: "ORD001"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Order == nil {
		t.Error("expected non-nil Order")
	}
}

func TestListOrders(t *testing.T) {
	srv := apiServer(t, "/shop/order/get_list", `{"errcode":0,"errmsg":"ok","order_list":[{"order_id":"ORD001"}],"total_num":1}`)
	defer srv.Close()
	resp, err := newTestClient(t, srv.URL).ListOrders(ctx(t), &ListOrdersReq{PageSize: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Orders) != 1 {
		t.Errorf("expected 1 order, got %d", len(resp.Orders))
	}
}

func TestUpdateOrderPrice(t *testing.T) {
	srv := apiServer(t, "/shop/order/update_price", `{"errcode":0,"errmsg":"ok"}`)
	defer srv.Close()
	if err := newTestClient(t, srv.URL).UpdateOrderPrice(ctx(t), &UpdateOrderPriceReq{OrderID: "ORD001", NewPrice: 100}); err != nil {
		t.Fatal(err)
	}
}

func TestCloseOrder(t *testing.T) {
	srv := apiServer(t, "/shop/order/close", `{"errcode":0,"errmsg":"ok"}`)
	defer srv.Close()
	if err := newTestClient(t, srv.URL).CloseOrder(ctx(t), &CloseOrderReq{OrderID: "ORD001"}); err != nil {
		t.Fatal(err)
	}
}

func TestUploadShipping(t *testing.T) {
	srv := apiServer(t, "/shop/delivery/send", `{"errcode":0,"errmsg":"ok"}`)
	defer srv.Close()
	req := &UploadShippingReq{OrderID: "ORD001", DeliveryCompany: "SF", DeliverySN: "SF123456"}
	if err := newTestClient(t, srv.URL).UploadShipping(ctx(t), req); err != nil {
		t.Fatal(err)
	}
}

func TestGetAfterSaleOrder(t *testing.T) {
	srv := apiServer(t, "/shop/aftersale/get_after_sale_order", `{"errcode":0,"errmsg":"ok","after_sale_order":{"id":"ASO001"}}`)
	defer srv.Close()
	resp, err := newTestClient(t, srv.URL).GetAfterSaleOrder(ctx(t), &GetAfterSaleOrderReq{AfterSaleOrderID: "ASO001"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.AfterSaleOrder == nil {
		t.Error("expected non-nil AfterSaleOrder")
	}
}

func TestAcceptRefund(t *testing.T) {
	srv := apiServer(t, "/shop/aftersale/accept_refund", `{"errcode":0,"errmsg":"ok"}`)
	defer srv.Close()
	if err := newTestClient(t, srv.URL).AcceptRefund(ctx(t), &AcceptRefundReq{AfterSaleOrderID: "ASO001"}); err != nil {
		t.Fatal(err)
	}
}

func TestRejectRefund(t *testing.T) {
	srv := apiServer(t, "/shop/aftersale/reject_refund", `{"errcode":0,"errmsg":"ok"}`)
	defer srv.Close()
	if err := newTestClient(t, srv.URL).RejectRefund(ctx(t), &RejectRefundReq{AfterSaleOrderID: "ASO001", RejectReason: "Not eligible"}); err != nil {
		t.Fatal(err)
	}
}
