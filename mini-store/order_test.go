package mini_store

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
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
	srv := apiServer(t, "/shop/aftersale/get_after_sale_order", `{"errcode":0,"errmsg":"ok","after_sale_order":{"after_sale_order_id":"ASO001"}}`)
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

func TestGetOrder_NetworkError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close() // close immediately to force a network error
	fake := &fakeTokenSource{token: "TOK"}
	c, err := NewClient(Config{AppId: "wx", AppSecret: "sec"},
		WithHTTP(utils.NewHTTP(srv.URL, utils.WithTimeout(3*time.Second))),
		WithTokenSource(fake))
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.GetOrder(context.Background(), &GetOrderReq{OrderID: "ORD001"})
	if err == nil {
		t.Error("expected network error, got nil")
	}
}
