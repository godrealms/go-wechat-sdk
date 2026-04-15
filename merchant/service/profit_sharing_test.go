package service

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestProfitSharingAddReceiver_FillsDefaults(t *testing.T) {
	c, fs, srv := newFakeClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{"type":"PERSONAL_OPENID","account":"oxxxxxxxxxxxxxxxxxxxx"}`)
	}

	_, err := c.ProfitSharingAddReceiver(context.Background(), map[string]any{
		"type":          "PERSONAL_OPENID",
		"account":       "oxxxxxxxxxxxxxxxxxxxx",
		"relation_type": "USER",
	})
	if err != nil {
		t.Fatalf("AddReceiver: %v", err)
	}
	req := fs.lastRequest(t)
	if req.Method != http.MethodPost {
		t.Errorf("method: %s", req.Method)
	}
	if !strings.HasSuffix(req.Path, "/v3/profitsharing/receivers/add") {
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
	if got := req.Header.Get("Wechatpay-Serial"); got != "" {
		t.Errorf("unexpected Wechatpay-Serial: %q", got)
	}
}

func TestProfitSharingAddReceiverWithSerial_SetsHeader(t *testing.T) {
	c, fs, srv := newFakeClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{"type":"PERSONAL_OPENID"}`)
	}
	_, err := c.ProfitSharingAddReceiverWithSerial(context.Background(), map[string]any{
		"type":    "PERSONAL_OPENID",
		"account": "oxxxxxxxxxxxxxxxxxxxx",
		"name":    "encrypted_ciphertext",
	}, "PLAT_SERIAL_XYZ")
	if err != nil {
		t.Fatalf("AddReceiverWithSerial: %v", err)
	}
	req := fs.lastRequest(t)
	if got := req.Header.Get("Wechatpay-Serial"); got != "PLAT_SERIAL_XYZ" {
		t.Errorf("Wechatpay-Serial: got %q, want PLAT_SERIAL_XYZ", got)
	}
}

func TestProfitSharingAddReceiver_RespectsCallerOverrides(t *testing.T) {
	c, fs, srv := newFakeClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) { return 200, []byte(`{}`) }

	_, err := c.ProfitSharingAddReceiver(context.Background(), map[string]any{
		"appid":     "override_app",
		"sub_mchid": "override_sub",
		"type":      "MERCHANT_ID",
	})
	if err != nil {
		t.Fatalf("AddReceiver: %v", err)
	}
	req := fs.lastRequest(t)
	var sent map[string]any
	if err := json.Unmarshal(req.Body, &sent); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if sent["appid"] != "override_app" {
		t.Errorf("appid override lost: %+v", sent)
	}
	if sent["sub_mchid"] != "override_sub" {
		t.Errorf("sub_mchid override lost: %+v", sent)
	}
}

func TestProfitSharingDeleteReceiver_FillsDefaults(t *testing.T) {
	c, fs, srv := newFakeClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) { return 200, []byte(`{}`) }

	_, err := c.ProfitSharingDeleteReceiver(context.Background(), map[string]any{
		"type":    "PERSONAL_OPENID",
		"account": "oxxxxxxxxxxxxxxxxxxxx",
	})
	if err != nil {
		t.Fatalf("DeleteReceiver: %v", err)
	}
	req := fs.lastRequest(t)
	if !strings.HasSuffix(req.Path, "/v3/profitsharing/receivers/delete") {
		t.Errorf("path: %s", req.Path)
	}
	var sent map[string]any
	if err := json.Unmarshal(req.Body, &sent); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if sent["sub_mchid"] != "1900000002" {
		t.Errorf("sub_mchid default missing: %+v", sent)
	}
}

func TestProfitSharingDeleteReceiver_RequiresBody(t *testing.T) {
	c, _, srv := newFakeClient(t)
	defer srv.Close()
	if _, err := c.ProfitSharingDeleteReceiver(context.Background(), nil); err == nil {
		t.Fatal("expected error for nil body")
	}
}

func TestProfitSharingMerchantConfig_UsesDefaultSubMchid(t *testing.T) {
	c, fs, srv := newFakeClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{"sub_mchid":"1900000002","max_ratio":2000}`)
	}

	resp, err := c.ProfitSharingMerchantConfig(context.Background(), "")
	if err != nil {
		t.Fatalf("MerchantConfig: %v", err)
	}
	// max_ratio 会被 json 解成 float64
	if v, _ := resp["max_ratio"].(float64); v != 2000 {
		t.Errorf("max_ratio: %v", resp["max_ratio"])
	}
	req := fs.lastRequest(t)
	if req.Method != http.MethodGet {
		t.Errorf("method: %s", req.Method)
	}
	if !strings.HasSuffix(req.Path, "/v3/profitsharing/merchant-configs/1900000002") {
		t.Errorf("path: %s", req.Path)
	}
}

func TestProfitSharingMerchantConfig_OverrideSubMchid(t *testing.T) {
	c, fs, srv := newFakeClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) { return 200, []byte(`{}`) }

	if _, err := c.ProfitSharingMerchantConfig(context.Background(), "1900999999"); err != nil {
		t.Fatalf("MerchantConfig: %v", err)
	}
	req := fs.lastRequest(t)
	if !strings.HasSuffix(req.Path, "/v3/profitsharing/merchant-configs/1900999999") {
		t.Errorf("path: %s", req.Path)
	}
}

func TestProfitSharingMerchantConfig_EmptyFailsWhenNoDefault(t *testing.T) {
	c, _, srv := newFakeClient(t)
	defer srv.Close()
	c.subMchid = ""
	if _, err := c.ProfitSharingMerchantConfig(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty subMchid and no default")
	}
}
