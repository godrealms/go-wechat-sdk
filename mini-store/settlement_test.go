package mini_store

import (
	"testing"
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
