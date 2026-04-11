package oplatform

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWxaAdmin_SetNickname(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/setnickname") {
			t.Errorf("path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok","wording":"","audit_id":"12345"}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	resp, err := w.SetNickname(context.Background(), &WxaSetNicknameReq{Nickname: "cool app"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.AuditID != "12345" {
		t.Errorf("audit_id: %q", resp.AuditID)
	}
	if gotBody["nick_name"] != "cool app" {
		t.Errorf("body: %+v", gotBody)
	}
}

func TestWxaAdmin_QueryNickname(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/api_wxa_querynickname") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"nickname":"cool","audit_stat":3,"create_time":1700,"audit_time":1800}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	resp, err := w.QueryNickname(context.Background(), "12345")
	if err != nil {
		t.Fatal(err)
	}
	if resp.Nickname != "cool" || resp.AuditStat != 3 {
		t.Errorf("unexpected: %+v", resp)
	}
}

func TestWxaAdmin_CheckNickname(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/wxverify/checkwxverifynickname") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"hit_condition":false,"wording":""}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	resp, err := w.CheckNickname(context.Background(), "cool")
	if err != nil {
		t.Fatal(err)
	}
	if resp.HitCondition {
		t.Errorf("expected false")
	}
}

func TestWxaAdmin_ModifyHeadImage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/account/modifyheadimage") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	if err := w.ModifyHeadImage(context.Background(), "MEDIA_ID"); err != nil {
		t.Fatal(err)
	}
}

func TestWxaAdmin_ModifySignature(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/account/modifysignature") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	if err := w.ModifySignature(context.Background(), "awesome sig"); err != nil {
		t.Fatal(err)
	}
}
