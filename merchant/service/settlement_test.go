package service

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestSettlementQuery_UsesDefaultSubMchid(t *testing.T) {
	c, fs, srv := newFakeClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{"account_type":"ACCOUNT_TYPE_BUSINESS","account_bank":"工商银行","account_number":"CIPHER","verify_result":"VERIFY_SUCCESS"}`)
	}

	resp, err := c.SettlementQuery(context.Background(), "")
	if err != nil {
		t.Fatalf("SettlementQuery: %v", err)
	}
	if resp.AccountType != "ACCOUNT_TYPE_BUSINESS" {
		t.Errorf("account_type: %s", resp.AccountType)
	}
	if resp.AccountNumber != "CIPHER" {
		t.Errorf("account_number: %s", resp.AccountNumber)
	}
	if resp.VerifyResult != "VERIFY_SUCCESS" {
		t.Errorf("verify_result: %s", resp.VerifyResult)
	}

	req := fs.lastRequest(t)
	if req.Method != http.MethodGet {
		t.Errorf("expected GET, got %s", req.Method)
	}
	if !strings.HasSuffix(req.Path, "/v3/applyment4sub/sub_merchants/1900000002/settlement") {
		t.Errorf("path: %s", req.Path)
	}
}

func TestSettlementQuery_OverrideSubMchid(t *testing.T) {
	c, fs, srv := newFakeClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{"account_type":"ACCOUNT_TYPE_PRIVATE","account_bank":"招商","account_number":"X"}`)
	}
	if _, err := c.SettlementQuery(context.Background(), "1900999999"); err != nil {
		t.Fatalf("SettlementQuery: %v", err)
	}
	req := fs.lastRequest(t)
	if !strings.HasSuffix(req.Path, "/v3/applyment4sub/sub_merchants/1900999999/settlement") {
		t.Errorf("path: %s", req.Path)
	}
}

func TestSettlementQuery_RequiresSubMchid(t *testing.T) {
	c, _, srv := newFakeClient(t)
	defer srv.Close()
	c.subMchid = ""
	if _, err := c.SettlementQuery(context.Background(), ""); err == nil {
		t.Fatal("expected error when no sub_mchid available")
	}
}

func TestSettlementModify_SendsPutWithSerial(t *testing.T) {
	c, fs, srv := newFakeClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		// 修改接口成功时返回 200 + 空响应体
		return 200, []byte(``)
	}

	err := c.SettlementModify(context.Background(), "", &SettlementModifyRequest{
		ModifyBalance: true,
		AccountType:   "ACCOUNT_TYPE_BUSINESS",
		AccountBank:   "工商银行",
		BankName:      "工商银行股份有限公司上海市分行营业部",
		BankBranchID:  "402713354941",
		AccountNumber: "PRE_ENCRYPTED_CIPHER",
	}, "PLAT_SERIAL_MOD")
	if err != nil {
		t.Fatalf("SettlementModify: %v", err)
	}

	req := fs.lastRequest(t)
	if req.Method != http.MethodPut {
		t.Errorf("expected PUT, got %s", req.Method)
	}
	if !strings.HasSuffix(req.Path, "/v3/applyment4sub/sub_merchants/1900000002/modify-settlement") {
		t.Errorf("path: %s", req.Path)
	}
	if got := req.Header.Get("Wechatpay-Serial"); got != "PLAT_SERIAL_MOD" {
		t.Errorf("Wechatpay-Serial: got %q, want PLAT_SERIAL_MOD", got)
	}
	var sent map[string]any
	if err := json.Unmarshal(req.Body, &sent); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if sent["modify_balance"] != true {
		t.Errorf("modify_balance not forwarded: %+v", sent)
	}
	if sent["account_number"] != "PRE_ENCRYPTED_CIPHER" {
		t.Errorf("account_number not forwarded: %+v", sent)
	}
}

func TestSettlementModify_RequiresRequest(t *testing.T) {
	c, _, srv := newFakeClient(t)
	defer srv.Close()
	if err := c.SettlementModify(context.Background(), "", nil, ""); err == nil {
		t.Fatal("expected error for nil request")
	}
}

func TestSettlementModify_RequiresEncryptedAccountNumber(t *testing.T) {
	c, _, srv := newFakeClient(t)
	defer srv.Close()
	err := c.SettlementModify(context.Background(), "", &SettlementModifyRequest{
		AccountType: "ACCOUNT_TYPE_BUSINESS",
		AccountBank: "工商银行",
	}, "")
	if err == nil {
		t.Fatal("expected error when account_number is missing")
	}
}

func TestSettlementModifyEncrypted_AutoEncryptsAndSetsSerial(t *testing.T) {
	c, fs, srv := newFakeClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(``)
	}

	req := &SettlementModifyRequest{
		AccountType: "ACCOUNT_TYPE_BUSINESS",
		AccountBank: "工商银行",
	}
	err := c.SettlementModifyEncrypted(context.Background(), "", req, "6222021234567890123")
	if err != nil {
		t.Fatalf("SettlementModifyEncrypted: %v", err)
	}
	// 入参 req 应保持未被修改，避免调用方残留密文状态。
	if req.AccountNumber != "" {
		t.Errorf("caller req.AccountNumber was mutated: %q", req.AccountNumber)
	}

	sent := fs.lastRequest(t)
	if got := sent.Header.Get("Wechatpay-Serial"); got == "" {
		t.Error("Wechatpay-Serial header missing after auto-encrypt")
	}
	var body map[string]any
	if err := json.Unmarshal(sent.Body, &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	cipher, _ := body["account_number"].(string)
	if cipher == "" {
		t.Fatal("account_number missing in request body")
	}
	if cipher == "6222021234567890123" {
		t.Error("account_number was not encrypted (still plaintext)")
	}
}

func TestSettlementModifyEncrypted_RequiresPlaintext(t *testing.T) {
	c, _, srv := newFakeClient(t)
	defer srv.Close()
	err := c.SettlementModifyEncrypted(context.Background(), "", &SettlementModifyRequest{
		AccountType: "ACCOUNT_TYPE_BUSINESS",
		AccountBank: "工商银行",
	}, "")
	if err == nil {
		t.Fatal("expected error when plaintext is empty")
	}
}
