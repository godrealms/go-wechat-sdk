package service

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestProfitSharingOrder_FillsDefaults(t *testing.T) {
	c, fs, srv := newFakeClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{"order_id":"3008450740201411110007820472"}`)
	}

	resp, err := c.ProfitSharingOrder(context.Background(), map[string]any{
		"transaction_id": "4208450740201411110007820472",
		"out_order_no":   "P20150806125346",
		"receivers": []map[string]any{
			{"type": "MERCHANT_ID", "account": "190001001", "amount": 100, "description": "分给商户"},
		},
		"unfreeze_unsplit": true,
	})
	if err != nil {
		t.Fatalf("ProfitSharingOrder: %v", err)
	}
	if resp["order_id"] != "3008450740201411110007820472" {
		t.Errorf("order_id: %v", resp["order_id"])
	}

	req := fs.lastRequest(t)
	if req.Method != http.MethodPost {
		t.Errorf("method: %s", req.Method)
	}
	if !strings.HasSuffix(req.Path, "/v3/profitsharing/orders") {
		t.Errorf("path: %s", req.Path)
	}
	var sent map[string]any
	if err := json.Unmarshal(req.Body, &sent); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if sent["appid"] != "wx_sp_appid" {
		t.Errorf("appid not injected: %+v", sent)
	}
	if sent["sub_mchid"] != "1900000002" {
		t.Errorf("sub_mchid not injected: %+v", sent)
	}
	// Wechatpay-Serial 未开启时不应设置。
	if got := req.Header.Get("Wechatpay-Serial"); got != "" {
		t.Errorf("unexpected Wechatpay-Serial: %q", got)
	}
}

func TestProfitSharingOrderWithSerial_SetsHeader(t *testing.T) {
	c, fs, srv := newFakeClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) { return 200, []byte(`{}`) }

	_, err := c.ProfitSharingOrderWithSerial(context.Background(), map[string]any{
		"transaction_id": "tx",
		"out_order_no":   "out1",
		"receivers": []map[string]any{
			{"type": "PERSONAL_OPENID", "account": "oxxxx", "amount": 1, "description": "x", "name": "ENC_NAME"},
		},
	}, "PLAT_SERIAL_ORDER")
	if err != nil {
		t.Fatalf("ProfitSharingOrderWithSerial: %v", err)
	}
	req := fs.lastRequest(t)
	if got := req.Header.Get("Wechatpay-Serial"); got != "PLAT_SERIAL_ORDER" {
		t.Errorf("Wechatpay-Serial: %q", got)
	}
}

func TestProfitSharingOrder_RequiresBody(t *testing.T) {
	c, _, srv := newFakeClient(t)
	defer srv.Close()
	if _, err := c.ProfitSharingOrder(context.Background(), nil); err == nil {
		t.Fatal("expected error for nil body")
	}
}

func TestProfitSharingQueryOrder_BuildsPathAndQuery(t *testing.T) {
	c, fs, srv := newFakeClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{"order_id":"3008","state":"FINISHED"}`)
	}

	if _, err := c.ProfitSharingQueryOrder(
		context.Background(), "", "P20150806125346", "tx_12345",
	); err != nil {
		t.Fatalf("ProfitSharingQueryOrder: %v", err)
	}
	req := fs.lastRequest(t)
	if req.Method != http.MethodGet {
		t.Errorf("method: %s", req.Method)
	}
	if !strings.Contains(req.Path, "/v3/profitsharing/orders/P20150806125346?") {
		t.Errorf("path: %s", req.Path)
	}
	if !strings.Contains(req.Path, "sub_mchid=1900000002") {
		t.Errorf("sub_mchid query missing: %s", req.Path)
	}
	if !strings.Contains(req.Path, "transaction_id=tx_12345") {
		t.Errorf("transaction_id query missing: %s", req.Path)
	}
}

func TestProfitSharingQueryOrder_RequiresArgs(t *testing.T) {
	c, _, srv := newFakeClient(t)
	defer srv.Close()
	if _, err := c.ProfitSharingQueryOrder(context.Background(), "", "", "tx"); err == nil {
		t.Error("expected error for empty outOrderNo")
	}
	if _, err := c.ProfitSharingQueryOrder(context.Background(), "", "out", ""); err == nil {
		t.Error("expected error for empty transactionId")
	}
}

func TestProfitSharingReturn_FillsDefaults(t *testing.T) {
	c, fs, srv := newFakeClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) { return 200, []byte(`{"return_id":"R1"}`) }

	_, err := c.ProfitSharingReturn(context.Background(), map[string]any{
		"order_id":      "3008",
		"out_return_no": "R20150806",
		"return_mchid":  "190001001",
		"amount":        10,
		"description":   "退",
	})
	if err != nil {
		t.Fatalf("ProfitSharingReturn: %v", err)
	}
	req := fs.lastRequest(t)
	if !strings.HasSuffix(req.Path, "/v3/profitsharing/return-orders") {
		t.Errorf("path: %s", req.Path)
	}
	var sent map[string]any
	_ = json.Unmarshal(req.Body, &sent)
	if sent["sub_mchid"] != "1900000002" {
		t.Errorf("sub_mchid not injected: %+v", sent)
	}
}

func TestProfitSharingReturn_RequiresBody(t *testing.T) {
	c, _, srv := newFakeClient(t)
	defer srv.Close()
	if _, err := c.ProfitSharingReturn(context.Background(), nil); err == nil {
		t.Fatal("expected error")
	}
}

func TestProfitSharingQueryReturn_BuildsPathAndQuery(t *testing.T) {
	c, fs, srv := newFakeClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) { return 200, []byte(`{"state":"SUCCESS"}`) }

	if _, err := c.ProfitSharingQueryReturn(
		context.Background(), "1900999999", "R20150806", "P20150806",
	); err != nil {
		t.Fatalf("ProfitSharingQueryReturn: %v", err)
	}
	req := fs.lastRequest(t)
	if !strings.Contains(req.Path, "/v3/profitsharing/return-orders/R20150806?") {
		t.Errorf("path: %s", req.Path)
	}
	if !strings.Contains(req.Path, "sub_mchid=1900999999") {
		t.Errorf("sub_mchid override lost: %s", req.Path)
	}
	if !strings.Contains(req.Path, "out_order_no=P20150806") {
		t.Errorf("out_order_no missing: %s", req.Path)
	}
}

func TestProfitSharingQueryReturn_RequiresArgs(t *testing.T) {
	c, _, srv := newFakeClient(t)
	defer srv.Close()
	if _, err := c.ProfitSharingQueryReturn(context.Background(), "", "", "P"); err == nil {
		t.Error("expected error for empty outReturnNo")
	}
	if _, err := c.ProfitSharingQueryReturn(context.Background(), "", "R", ""); err == nil {
		t.Error("expected error for empty outOrderNo")
	}
}

func TestProfitSharingUnfreeze_FillsDefaults(t *testing.T) {
	c, fs, srv := newFakeClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) { return 200, []byte(`{"state":"FINISHED"}`) }

	_, err := c.ProfitSharingUnfreeze(context.Background(), map[string]any{
		"transaction_id": "tx_12345",
		"out_order_no":   "P20150806",
		"description":    "解冻剩余",
	})
	if err != nil {
		t.Fatalf("ProfitSharingUnfreeze: %v", err)
	}
	req := fs.lastRequest(t)
	if !strings.HasSuffix(req.Path, "/v3/profitsharing/orders/unfreeze") {
		t.Errorf("path: %s", req.Path)
	}
	var sent map[string]any
	_ = json.Unmarshal(req.Body, &sent)
	if sent["sub_mchid"] != "1900000002" {
		t.Errorf("sub_mchid not injected: %+v", sent)
	}
	if sent["appid"] != "wx_sp_appid" {
		t.Errorf("appid not injected: %+v", sent)
	}
}

func TestProfitSharingMerchantAmounts_BuildsQuery(t *testing.T) {
	c, fs, srv := newFakeClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) { return 200, []byte(`{"unsplit_amount":100}`) }

	if _, err := c.ProfitSharingMerchantAmounts(context.Background(), "", "tx_12345"); err != nil {
		t.Fatalf("ProfitSharingMerchantAmounts: %v", err)
	}
	req := fs.lastRequest(t)
	if !strings.Contains(req.Path, "/v3/profitsharing/transactions/tx_12345/amounts?") {
		t.Errorf("path: %s", req.Path)
	}
	if !strings.Contains(req.Path, "sub_mchid=1900000002") {
		t.Errorf("sub_mchid query missing: %s", req.Path)
	}
}

func TestProfitSharingMerchantAmounts_RequiresTransactionId(t *testing.T) {
	c, _, srv := newFakeClient(t)
	defer srv.Close()
	if _, err := c.ProfitSharingMerchantAmounts(context.Background(), "", ""); err == nil {
		t.Fatal("expected error")
	}
}

func TestProfitSharingBills_BuildsQuery(t *testing.T) {
	c, fs, srv := newFakeClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{"download_url":"https://api.mch.weixin.qq.com/v3/billdownload/file?token=xxx"}`)
	}

	if _, err := c.ProfitSharingBills(context.Background(), "2026-04-15", "", "GZIP"); err != nil {
		t.Fatalf("ProfitSharingBills: %v", err)
	}
	req := fs.lastRequest(t)
	if !strings.HasPrefix(req.Path, "/v3/profitsharing/bills?") {
		t.Errorf("path: %s", req.Path)
	}
	if !strings.Contains(req.Path, "bill_date=2026-04-15") {
		t.Errorf("bill_date missing: %s", req.Path)
	}
	if !strings.Contains(req.Path, "sub_mchid=1900000002") {
		t.Errorf("sub_mchid missing: %s", req.Path)
	}
	if !strings.Contains(req.Path, "tar_type=GZIP") {
		t.Errorf("tar_type missing: %s", req.Path)
	}
}

func TestProfitSharingBills_RequiresBillDate(t *testing.T) {
	c, _, srv := newFakeClient(t)
	defer srv.Close()
	if _, err := c.ProfitSharingBills(context.Background(), "", "", ""); err == nil {
		t.Fatal("expected error")
	}
}
