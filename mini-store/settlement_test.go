package mini_store

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

func TestGetMerchantInfo(t *testing.T) {
	srv := apiServer(t, "/shop/merchant/get_merchant_info", `{"errcode":0,"errmsg":"ok","merchant_info":{"name":"Test Shop"}}`)
	defer srv.Close()
	resp, err := newTestClient(t, srv.URL).GetMerchantInfo(ctx(t))
	if err != nil {
		t.Fatal(err)
	}
	if resp.MerchantInfo == nil {
		t.Error("expected non-nil MerchantInfo")
	}
}

func TestGetSettlement(t *testing.T) {
	srv := apiServer(t, "/shop/pay/get_pay_list", `{"errcode":0,"errmsg":"ok","settlement_list":[{"id":"S001"}]}`)
	defer srv.Close()
	resp, err := newTestClient(t, srv.URL).GetSettlement(ctx(t), &GetSettlementReq{StartTime: 1700000000, EndTime: 1700086400})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.SettlementList) == 0 {
		t.Error("expected non-empty SettlementList")
	}
}

func TestGetBrandList(t *testing.T) {
	srv := apiServer(t, "/shop/account/get_brand_list", `{"errcode":0,"errmsg":"ok","brand_list":[{"id":1,"name":"Apple"}]}`)
	defer srv.Close()
	resp, err := newTestClient(t, srv.URL).GetBrandList(ctx(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.BrandList) == 0 {
		t.Error("expected non-empty BrandList")
	}
}

func TestGetCategoryList(t *testing.T) {
	srv := apiServer(t, "/shop/cat/get", `{"errcode":0,"errmsg":"ok","cat_list":[{"f_cat_id":0,"name":"Electronics"}]}`)
	defer srv.Close()
	resp, err := newTestClient(t, srv.URL).GetCategoryList(ctx(t), &GetCategoryListReq{})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.CategoryList) == 0 {
		t.Error("expected non-empty CategoryList")
	}
}

func TestGetMerchantInfo_NetworkError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close() // close immediately to force a network error
	fake := &fakeTokenSource{token: "TOK"}
	c, err := NewClient(Config{AppId: "wx", AppSecret: "sec"},
		WithHTTP(utils.NewHTTP(srv.URL, utils.WithTimeout(3*time.Second))),
		WithTokenSource(fake))
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.GetMerchantInfo(context.Background())
	if err == nil {
		t.Error("expected network error, got nil")
	}
}
