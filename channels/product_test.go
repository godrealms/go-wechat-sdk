package channels

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAddProduct(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/token":
			_, _ = w.Write([]byte(`{"access_token":"TEST_TOKEN","expires_in":7200}`))
		case "/channels/ec/product/add":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Query().Get("access_token") != "TEST_TOKEN" {
				t.Errorf("missing access_token")
			}
			var req AddProductReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatal(err)
			}
			if req.Product.Title != "Test Product" {
				t.Errorf("expected title Test Product, got %s", req.Product.Title)
			}
			_, _ = w.Write([]byte(`{"product_id":"prod_001"}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	resp, err := c.AddProduct(context.Background(), &AddProductReq{
		Product: ProductInfo{Title: "Test Product"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.ProductID != "prod_001" {
		t.Errorf("got product_id=%s, want prod_001", resp.ProductID)
	}
}

func TestUpdateProduct(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/token":
			_, _ = w.Write([]byte(`{"access_token":"TEST_TOKEN","expires_in":7200}`))
		case "/channels/ec/product/update":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			var req UpdateProductReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatal(err)
			}
			if req.Product.ProductID != "prod_001" {
				t.Errorf("expected product_id prod_001, got %s", req.Product.ProductID)
			}
			_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	err := c.UpdateProduct(context.Background(), &UpdateProductReq{
		Product: ProductInfo{ProductID: "prod_001", Title: "Updated"},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetProduct(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/token":
			_, _ = w.Write([]byte(`{"access_token":"TEST_TOKEN","expires_in":7200}`))
		case "/channels/ec/product/get":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			var req GetProductReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatal(err)
			}
			if req.ProductID != "prod_001" {
				t.Errorf("expected product_id prod_001, got %s", req.ProductID)
			}
			_, _ = w.Write([]byte(`{"product":{"product_id":"prod_001","title":"Test Product","status":1,"create_time":1000}}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	resp, err := c.GetProduct(context.Background(), &GetProductReq{ProductID: "prod_001"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Product.ProductID != "prod_001" {
		t.Errorf("got product_id=%s, want prod_001", resp.Product.ProductID)
	}
	if resp.Product.Title != "Test Product" {
		t.Errorf("got title=%s, want Test Product", resp.Product.Title)
	}
}

func TestListProduct(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/token":
			_, _ = w.Write([]byte(`{"access_token":"TEST_TOKEN","expires_in":7200}`))
		case "/channels/ec/product/list":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			_, _ = w.Write([]byte(`{"products":[{"product_id":"prod_001","title":"P1"},{"product_id":"prod_002","title":"P2"}],"total":2}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	limit := 10
	resp, err := c.ListProduct(context.Background(), &ListProductReq{Limit: &limit})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Total != 2 {
		t.Errorf("got total=%d, want 2", resp.Total)
	}
	if len(resp.Products) != 2 {
		t.Fatalf("got %d products, want 2", len(resp.Products))
	}
}

func TestDeleteProduct(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/token":
			_, _ = w.Write([]byte(`{"access_token":"TEST_TOKEN","expires_in":7200}`))
		case "/channels/ec/product/delete":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			var req DeleteProductReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatal(err)
			}
			if req.ProductID != "prod_001" {
				t.Errorf("expected product_id prod_001, got %s", req.ProductID)
			}
			_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	err := c.DeleteProduct(context.Background(), &DeleteProductReq{ProductID: "prod_001"})
	if err != nil {
		t.Fatal(err)
	}
}
