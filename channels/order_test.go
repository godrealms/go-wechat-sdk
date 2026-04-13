package channels

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetOrder(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/token":
			_, _ = w.Write([]byte(`{"access_token":"TEST_TOKEN","expires_in":7200}`))
		case "/channels/ec/order/get":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Query().Get("access_token") != "TEST_TOKEN" {
				t.Errorf("missing access_token")
			}
			var req GetOrderReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatal(err)
			}
			if req.OrderID != "order_001" {
				t.Errorf("expected order_id order_001, got %s", req.OrderID)
			}
			_, _ = w.Write([]byte(`{"order":{"order_id":"order_001","product_id":"prod_001","status":1,"create_time":1000,"update_time":2000}}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	resp, err := c.GetOrder(context.Background(), &GetOrderReq{OrderID: "order_001"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Order.OrderID != "order_001" {
		t.Errorf("got order_id=%s, want order_001", resp.Order.OrderID)
	}
	if resp.Order.Status != 1 {
		t.Errorf("got status=%d, want 1", resp.Order.Status)
	}
}

func TestListOrder(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/token":
			_, _ = w.Write([]byte(`{"access_token":"TEST_TOKEN","expires_in":7200}`))
		case "/channels/ec/order/list":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			_, _ = w.Write([]byte(`{"orders":[{"order_id":"order_001","product_id":"prod_001","status":1,"create_time":1000,"update_time":2000}],"total":1}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	limit := 10
	resp, err := c.ListOrder(context.Background(), &ListOrderReq{Limit: &limit})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Total != 1 {
		t.Errorf("got total=%d, want 1", resp.Total)
	}
	if len(resp.Orders) != 1 {
		t.Fatalf("got %d orders, want 1", len(resp.Orders))
	}
	if resp.Orders[0].OrderID != "order_001" {
		t.Errorf("got order_id=%s, want order_001", resp.Orders[0].OrderID)
	}
}
