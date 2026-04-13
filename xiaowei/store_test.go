package xiaowei

import (
	"context"
	"net/http/httptest"
	"testing"
)

func TestGetStoreInfo(t *testing.T) {
	srv := httptest.NewServer(apiHandler("/wxaapi/wxamicrostore/get_store_info",
		`{"errcode":0,"errmsg":"ok","store_info":{"store_name":"My Shop","store_status":1}}`))
	defer srv.Close()
	resp, err := newTestClient(t, srv.URL).GetStoreInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if resp.StoreInfo == nil {
		t.Error("expected non-nil StoreInfo")
	}
}

func TestUpdateStoreInfo(t *testing.T) {
	srv := httptest.NewServer(apiHandler("/wxaapi/wxamicrostore/update_store_info",
		`{"errcode":0,"errmsg":"ok"}`))
	defer srv.Close()
	if err := newTestClient(t, srv.URL).UpdateStoreInfo(context.Background(), &UpdateStoreInfoReq{StoreName: "New Name"}); err != nil {
		t.Fatal(err)
	}
}

func TestGetKYCStatus(t *testing.T) {
	srv := httptest.NewServer(apiHandler("/wxaapi/wxamicrostore/get_kyc_status",
		`{"errcode":0,"errmsg":"ok","kyc_status":2}`))
	defer srv.Close()
	resp, err := newTestClient(t, srv.URL).GetKYCStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if resp.Status != 2 {
		t.Errorf("expected status=2, got %d", resp.Status)
	}
}

func TestSubmitKYC(t *testing.T) {
	srv := httptest.NewServer(apiHandler("/wxaapi/wxamicrostore/submit_kyc",
		`{"errcode":0,"errmsg":"ok"}`))
	defer srv.Close()
	req := &SubmitKYCReq{RealName: "Zhang San", IDCardNo: "110101199001011234"}
	if err := newTestClient(t, srv.URL).SubmitKYC(context.Background(), req); err != nil {
		t.Fatal(err)
	}
}
