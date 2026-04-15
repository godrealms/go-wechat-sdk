package service

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestApplymentSubmit_HappyPath(t *testing.T) {
	c, fs, srv := newFakeClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{"applyment_id":2000002124775691}`)
	}

	resp, err := c.ApplymentSubmit(context.Background(), map[string]any{
		"business_code": "1900013511_10000",
	}, "PLATFORM_SERIAL_ABC")
	if err != nil {
		t.Fatalf("ApplymentSubmit failed: %v", err)
	}
	if resp.ApplymentID != 2000002124775691 {
		t.Errorf("unexpected applyment_id: %d", resp.ApplymentID)
	}

	req := fs.lastRequest(t)
	if req.Method != http.MethodPost {
		t.Errorf("expected POST, got %s", req.Method)
	}
	if !strings.HasSuffix(req.Path, "/v3/applyment4sub/applyment/") {
		t.Errorf("unexpected path: %s", req.Path)
	}
	if got := req.Header.Get("Wechatpay-Serial"); got != "PLATFORM_SERIAL_ABC" {
		t.Errorf("Wechatpay-Serial header: got %q, want PLATFORM_SERIAL_ABC", got)
	}
	var sent map[string]any
	if err := json.Unmarshal(req.Body, &sent); err != nil {
		t.Fatalf("unmarshal sent body: %v", err)
	}
	if sent["business_code"] != "1900013511_10000" {
		t.Errorf("business_code not forwarded: %+v", sent)
	}
}

func TestApplymentSubmit_RequiresBody(t *testing.T) {
	c, _, srv := newFakeClient(t)
	defer srv.Close()
	if _, err := c.ApplymentSubmit(context.Background(), nil, ""); err == nil {
		t.Fatal("expected error for nil body")
	}
}

func TestApplymentSubmit_OmitsSerialHeaderWhenEmpty(t *testing.T) {
	c, fs, srv := newFakeClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{"applyment_id":1}`)
	}
	_, err := c.ApplymentSubmit(context.Background(), map[string]any{"business_code": "x"}, "")
	if err != nil {
		t.Fatalf("ApplymentSubmit: %v", err)
	}
	req := fs.lastRequest(t)
	if got := req.Header.Get("Wechatpay-Serial"); got != "" {
		t.Errorf("Wechatpay-Serial unexpectedly set to %q", got)
	}
}

func TestApplymentQueryByBusinessCode_BuildsPath(t *testing.T) {
	c, fs, srv := newFakeClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{"business_code":"BC1","applyment_state":"APPLYMENT_STATE_FINISHED","applyment_state_msg":"ok"}`)
	}

	resp, err := c.ApplymentQueryByBusinessCode(context.Background(), "BC1")
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if resp.ApplymentState != "APPLYMENT_STATE_FINISHED" {
		t.Errorf("applyment_state: %s", resp.ApplymentState)
	}
	req := fs.lastRequest(t)
	if req.Method != http.MethodGet {
		t.Errorf("expected GET, got %s", req.Method)
	}
	if !strings.HasSuffix(req.Path, "/v3/applyment4sub/applyment/business_code/BC1") {
		t.Errorf("path: %s", req.Path)
	}
}

func TestApplymentQueryByID_BuildsPath(t *testing.T) {
	c, fs, srv := newFakeClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{"applyment_id":2000002124775691,"applyment_state":"APPLYMENT_STATE_AUDITING","applyment_state_msg":"审核中"}`)
	}
	if _, err := c.ApplymentQueryByID(context.Background(), 2000002124775691); err != nil {
		t.Fatalf("query: %v", err)
	}
	req := fs.lastRequest(t)
	if !strings.HasSuffix(req.Path, "/v3/applyment4sub/applyment/applyment_id/2000002124775691") {
		t.Errorf("path: %s", req.Path)
	}
}

func TestApplymentQueryByID_RejectsZero(t *testing.T) {
	c, _, srv := newFakeClient(t)
	defer srv.Close()
	if _, err := c.ApplymentQueryByID(context.Background(), 0); err == nil {
		t.Fatal("expected error for zero applymentID")
	}
}

func TestApplymentQueryByBusinessCode_RejectsEmpty(t *testing.T) {
	c, _, srv := newFakeClient(t)
	defer srv.Close()
	if _, err := c.ApplymentQueryByBusinessCode(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty business code")
	}
}

func TestEncryptSensitive_ReturnsSerialAndCiphertext(t *testing.T) {
	c, _, srv := newFakeClient(t)
	defer srv.Close()

	ct, serial, err := c.EncryptSensitive(context.Background(), "张三")
	if err != nil {
		t.Fatalf("EncryptSensitive: %v", err)
	}
	if ct == "" {
		t.Error("ciphertext empty")
	}
	if serial == "" {
		t.Error("serial empty")
	}
}
