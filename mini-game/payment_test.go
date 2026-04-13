package mini_game

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateOrder_ErrcodeError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/token":
			_, _ = w.Write([]byte(`{"access_token":"TEST_TOKEN","expires_in":7200}`))
		default:
			_, _ = w.Write([]byte(`{"errcode":40001,"errmsg":"invalid token"}`))
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.CreateOrder(context.Background(), &CreateOrderReq{
		OpenID: "oUSER1", ProductID: "prod_001", Quantity: 1,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.Code() != 40001 {
		t.Errorf("expected Code() == 40001, got %d", apiErr.Code())
	}
}

func TestCreateOrder(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/token":
			_, _ = w.Write([]byte(`{"access_token":"TEST_TOKEN","expires_in":7200}`))
		case "/wxa/game/createorder":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Query().Get("access_token") != "TEST_TOKEN" {
				t.Errorf("missing access_token")
			}
			var req CreateOrderReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatal(err)
			}
			if req.OpenID != "oUSER1" {
				t.Errorf("expected openid oUSER1, got %s", req.OpenID)
			}
			if req.ProductID != "prod_001" {
				t.Errorf("expected product_id prod_001, got %s", req.ProductID)
			}
			if req.Quantity != 1 {
				t.Errorf("expected quantity 1, got %d", req.Quantity)
			}
			_, _ = w.Write([]byte(`{"order_id":"order_001","balance":9900}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	resp, err := c.CreateOrder(context.Background(), &CreateOrderReq{
		OpenID:    "oUSER1",
		Env:       0,
		ZoneID:    "zone_1",
		ProductID: "prod_001",
		Quantity:  1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.OrderID != "order_001" {
		t.Errorf("got order_id=%s, want order_001", resp.OrderID)
	}
	if resp.Balance != 9900 {
		t.Errorf("got balance=%d, want 9900", resp.Balance)
	}
}

func TestQueryOrder(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/token":
			_, _ = w.Write([]byte(`{"access_token":"TEST_TOKEN","expires_in":7200}`))
		case "/wxa/game/queryorder":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Query().Get("access_token") != "TEST_TOKEN" {
				t.Errorf("missing access_token")
			}
			var req QueryOrderReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatal(err)
			}
			if req.OrderID != "order_001" {
				t.Errorf("expected order_id order_001, got %s", req.OrderID)
			}
			_, _ = w.Write([]byte(`{"order_id":"order_001","status":1,"pay_amount":100,"create_time":1700000000}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	resp, err := c.QueryOrder(context.Background(), &QueryOrderReq{
		OrderID: "order_001",
		OpenID:  "oUSER1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.OrderID != "order_001" {
		t.Errorf("got order_id=%s, want order_001", resp.OrderID)
	}
	if resp.Status != 1 {
		t.Errorf("got status=%d, want 1", resp.Status)
	}
	if resp.PayAmount != 100 {
		t.Errorf("got pay_amount=%d, want 100", resp.PayAmount)
	}
	if resp.CreateTime != 1700000000 {
		t.Errorf("got create_time=%d, want 1700000000", resp.CreateTime)
	}
}
