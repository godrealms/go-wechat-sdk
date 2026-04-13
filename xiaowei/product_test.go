package xiaowei

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

func TestAddMicroProduct(t *testing.T) {
	srv := httptest.NewServer(apiHandler("/wxaapi/wxamicrostore/add_product",
		`{"errcode":0,"errmsg":"ok","product_id":"P001"}`))
	defer srv.Close()
	resp, err := newTestClient(t, srv.URL).AddMicroProduct(context.Background(), &MicroProduct{Title: "Widget", Price: 999})
	if err != nil {
		t.Fatal(err)
	}
	if resp.ProductID != "P001" {
		t.Errorf("got product_id=%q", resp.ProductID)
	}
}

func TestDelMicroProduct(t *testing.T) {
	srv := httptest.NewServer(apiHandler("/wxaapi/wxamicrostore/del_product",
		`{"errcode":0,"errmsg":"ok"}`))
	defer srv.Close()
	if err := newTestClient(t, srv.URL).DelMicroProduct(context.Background(), &DelMicroProductReq{ProductID: "P001"}); err != nil {
		t.Fatal(err)
	}
}

func TestGetMicroProduct(t *testing.T) {
	srv := httptest.NewServer(apiHandler("/wxaapi/wxamicrostore/get_product",
		`{"errcode":0,"errmsg":"ok","product":{"product_id":"P001","title":"Widget","price":999}}`))
	defer srv.Close()
	resp, err := newTestClient(t, srv.URL).GetMicroProduct(context.Background(), &GetMicroProductReq{ProductID: "P001"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Product == nil {
		t.Error("expected non-nil Product")
	}
}

func TestListMicroProducts(t *testing.T) {
	srv := httptest.NewServer(apiHandler("/wxaapi/wxamicrostore/get_product_list",
		`{"errcode":0,"errmsg":"ok","product_list":[{"product_id":"P001","title":"Widget"}],"total":1}`))
	defer srv.Close()
	resp, err := newTestClient(t, srv.URL).ListMicroProducts(context.Background(), &ListMicroProductsReq{PageSize: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Products) != 1 {
		t.Errorf("expected 1 product, got %d", len(resp.Products))
	}
}

func TestAddMicroProduct_NetworkError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close() // close immediately to force a network error
	fake := &fakeTokenSource{token: "TOK"}
	c, err := NewClient(Config{AppId: "wx", AppSecret: "sec"},
		WithHTTP(utils.NewHTTP(srv.URL, utils.WithTimeout(3*time.Second))),
		WithTokenSource(fake))
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.AddMicroProduct(context.Background(), &MicroProduct{Title: "Widget", Price: 999})
	if err == nil {
		t.Error("expected network error, got nil")
	}
}
