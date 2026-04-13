package service

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
)

func newTestClient(t *testing.T) *Client {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	c, err := NewClient(Config{
		SpMchid:           "1900000001",
		SpAppid:           "wx_sp_appid",
		SubMchid:          "1900000002",
		SubAppid:          "wx_sub_appid",
		CertificateNumber: "TEST",
		APIv3Key:          "01234567890123456789012345678901",
		PrivateKey:        priv,
	})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestNewClient_RequiresFields(t *testing.T) {
	if _, err := NewClient(Config{}); err == nil {
		t.Fatal("expected error for empty config")
	}
	if _, err := NewClient(Config{SpMchid: "x", SpAppid: "y"}); err == nil {
		t.Fatal("expected error when SubMchid missing")
	}
}

func TestInjectSubFields_DoesNotMutateCaller(t *testing.T) {
	c := newTestClient(t)
	in := map[string]any{"description": "demo"}
	out := c.injectSubFields(in)

	if _, ok := in["sp_mchid"]; ok {
		t.Errorf("caller map was mutated: %+v", in)
	}
	if out["sp_mchid"] != "1900000001" {
		t.Errorf("sp_mchid not injected: %+v", out)
	}
	if out["sp_appid"] != "wx_sp_appid" {
		t.Errorf("sp_appid not injected: %+v", out)
	}
	if out["sub_mchid"] != "1900000002" {
		t.Errorf("sub_mchid not injected: %+v", out)
	}
	if out["sub_appid"] != "wx_sub_appid" {
		t.Errorf("sub_appid not injected: %+v", out)
	}
	if out["description"] != "demo" {
		t.Errorf("original fields lost: %+v", out)
	}
}

func TestInjectSubFields_RespectsCallerOverrides(t *testing.T) {
	c := newTestClient(t)
	in := map[string]any{
		"sp_mchid":  "override_sp",
		"sub_mchid": "override_sub",
	}
	out := c.injectSubFields(in)
	if out["sp_mchid"] != "override_sp" {
		t.Errorf("override lost: %+v", out)
	}
	if out["sub_mchid"] != "override_sub" {
		t.Errorf("override lost: %+v", out)
	}
}

func TestInjectSubFields_NilBody(t *testing.T) {
	c := newTestClient(t)
	out := c.injectSubFields(nil)
	if out["sp_mchid"] != "1900000001" {
		t.Errorf("expected defaults injected, got %+v", out)
	}
}
