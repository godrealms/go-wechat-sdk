package mini_store

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func apiServer(t *testing.T, apiPath, respBody string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/token":
			_, _ = w.Write([]byte(`{"access_token":"TOK","expires_in":7200}`))
		case apiPath:
			_, _ = w.Write([]byte(respBody))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestAddProduct(t *testing.T) {
	for _, tc := range []struct {
		name    string
		resp    string
		wantErr bool
	}{
		{"success", `{"errcode":0,"errmsg":"ok","product_id":"spu_001"}`, false},
		{"error", `{"errcode":40001,"errmsg":"invalid credential"}`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := apiServer(t, "/shop/spu/add", tc.resp)
			defer srv.Close()
			c := newTestClient(t, srv.URL)
			resp, err := c.AddProduct(ctx(t), &Product{Title: "Test"})
			if tc.wantErr != (err != nil) {
				t.Fatalf("wantErr=%v got err=%v", tc.wantErr, err)
			}
			if !tc.wantErr && resp.ProductID != "spu_001" {
				t.Errorf("got product_id=%q", resp.ProductID)
			}
		})
	}
}

func TestDelProduct(t *testing.T) {
	srv := apiServer(t, "/shop/spu/del", `{"errcode":0,"errmsg":"ok"}`)
	defer srv.Close()
	if err := newTestClient(t, srv.URL).DelProduct(ctx(t), &DelProductReq{ProductID: "spu_001"}); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateProduct(t *testing.T) {
	srv := apiServer(t, "/shop/spu/update", `{"errcode":0,"errmsg":"ok"}`)
	defer srv.Close()
	req := &UpdateProductReq{ProductID: "spu_001", Product: &Product{Title: "Updated"}}
	if err := newTestClient(t, srv.URL).UpdateProduct(ctx(t), req); err != nil {
		t.Fatal(err)
	}
}

func TestGetProduct(t *testing.T) {
	srv := apiServer(t, "/shop/spu/get", `{"errcode":0,"errmsg":"ok","spu":{"title":"Test","status":1}}`)
	defer srv.Close()
	resp, err := newTestClient(t, srv.URL).GetProduct(ctx(t), &GetProductReq{ProductID: "spu_001"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.SPU == nil {
		t.Error("expected non-nil SPU")
	}
}

func TestListProducts(t *testing.T) {
	srv := apiServer(t, "/shop/spu/get_list", `{"errcode":0,"errmsg":"ok","spus":[{"title":"T1"}],"total_num":1}`)
	defer srv.Close()
	resp, err := newTestClient(t, srv.URL).ListProducts(ctx(t), &ListProductsReq{PageSize: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.SPUs) != 1 {
		t.Errorf("expected 1 SPU, got %d", len(resp.SPUs))
	}
}

func TestUpdateProductStatus(t *testing.T) {
	srv := apiServer(t, "/shop/spu/update_without_audit", `{"errcode":0,"errmsg":"ok"}`)
	defer srv.Close()
	if err := newTestClient(t, srv.URL).UpdateProductStatus(ctx(t), &UpdateProductStatusReq{ProductID: "spu_001", Status: 1}); err != nil {
		t.Fatal(err)
	}
}

func TestSubmitProductAudit(t *testing.T) {
	srv := apiServer(t, "/shop/audit/audit_spu", `{"errcode":0,"errmsg":"ok"}`)
	defer srv.Close()
	if err := newTestClient(t, srv.URL).SubmitProductAudit(ctx(t), &SubmitProductAuditReq{ProductID: "spu_001"}); err != nil {
		t.Fatal(err)
	}
}

func TestCancelProductAudit(t *testing.T) {
	srv := apiServer(t, "/shop/audit/cancel_audit_spu", `{"errcode":0,"errmsg":"ok"}`)
	defer srv.Close()
	if err := newTestClient(t, srv.URL).CancelProductAudit(ctx(t), &CancelProductAuditReq{ProductID: "spu_001"}); err != nil {
		t.Fatal(err)
	}
}

func ctx(t *testing.T) context.Context { t.Helper(); return context.Background() }
