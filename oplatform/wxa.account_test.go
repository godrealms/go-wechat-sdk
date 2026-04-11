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

func TestWxaAdmin_GetAccountBasicInfo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/account/getaccountbasicinfo") {
			t.Errorf("path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("access_token") != "ATOK" {
			t.Errorf("access_token: %s", r.URL.Query().Get("access_token"))
		}
		_, _ = w.Write([]byte(`{
  "errcode": 0,
  "appid": "wxBIZ",
  "account_type": 0,
  "principal_type": 1,
  "principal_name": "Acme Corp",
  "realname_status": 1,
  "nickname": "Cool App",
  "head_img": "https://example.com/h.png",
  "signature": "hello",
  "registered_country": 1,
  "wx_verify_info": {"qualification_verify": true, "naming_verify": true},
  "signature_info": {"signature": "hello", "modify_used_count": 1, "modify_quota": 5},
  "head_image_info": {"head_image_url": "https://example.com/h.png", "modify_used_count": 0, "modify_quota": 5}
}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	info, err := w.GetAccountBasicInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if info.AppID != "wxBIZ" || info.PrincipalName != "Acme Corp" {
		t.Errorf("top-level: %+v", info)
	}
	if !info.WxVerifyInfo.QualificationVerify {
		t.Errorf("wx_verify_info not parsed")
	}
	if info.SignatureInfo.ModifyQuota != 5 {
		t.Errorf("signature_info: %+v", info.SignatureInfo)
	}
	if info.HeadImageInfo.HeadImageURL == "" {
		t.Errorf("head_image_info: %+v", info.HeadImageInfo)
	}
}
