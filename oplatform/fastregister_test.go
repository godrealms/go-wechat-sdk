package oplatform

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newTestFastRegister seeds the store with a non-expired component token
// so ComponentAccessToken() returns from cache and test mocks don't need
// to handle /cgi-bin/component/api_component_token.
func newTestFastRegister(t *testing.T, baseURL string) *FastRegisterClient {
	t.Helper()
	store := NewMemoryStore()
	_ = store.SetVerifyTicket(context.Background(), "TICKET")
	_ = store.SetComponentToken(context.Background(), "CTOK", time.Now().Add(time.Hour))
	c := newTestClient(t, baseURL, WithStore(store))
	return c.FastRegister()
}

func TestFastRegister_DoPost_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/fake_endpoint") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("component_access_token") != "CTOK" {
			t.Errorf("missing component_access_token, got %q", r.URL.Query().Get("component_access_token"))
		}
		_, _ = w.Write([]byte(`{"errcode":0,"value":42}`))
	}))
	defer srv.Close()

	f := newTestFastRegister(t, srv.URL)
	var out struct {
		Value int `json:"value"`
	}
	if err := f.doPost(context.Background(), "/fake_endpoint", map[string]string{"k": "v"}, &out); err != nil {
		t.Fatal(err)
	}
	if out.Value != 42 {
		t.Errorf("got %d", out.Value)
	}
}

func TestFastRegister_DoPost_PathWithExistingQuery(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("action") != "create" {
			t.Errorf("missing action=create: %s", r.URL.RawQuery)
		}
		if q.Get("component_access_token") != "CTOK" {
			t.Errorf("missing component_access_token: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()

	f := newTestFastRegister(t, srv.URL)
	if err := f.doPost(context.Background(), "/fake_endpoint?action=create", nil, nil); err != nil {
		t.Fatal(err)
	}
}

func TestFastRegister_DoPost_Errcode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":89249,"errmsg":"still creating"}`))
	}))
	defer srv.Close()
	f := newTestFastRegister(t, srv.URL)

	err := f.doPost(context.Background(), "/fake_endpoint", nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var werr *WeixinError
	if !errors.As(err, &werr) || werr.ErrCode != 89249 {
		t.Errorf("expected WeixinError 89249, got %v", err)
	}
}

func TestFastRegister_CreateEnterpriseAccount(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/component/fastregisterweapp") {
			t.Errorf("path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("action") != "create" {
			t.Errorf("action: %s", r.URL.Query().Get("action"))
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	f := newTestFastRegister(t, srv.URL)

	_, err := f.CreateEnterpriseAccount(context.Background(), &FastRegEnterpriseReq{
		Name:               "Acme Corp",
		Code:               "91310000123456789X",
		CodeType:           1,
		LegalPersonaWechat: "legal_wx",
		LegalPersonaName:   "张三",
		ComponentPhone:     "13800000000",
	})
	if err != nil {
		t.Fatal(err)
	}
	if body["name"] != "Acme Corp" {
		t.Errorf("name: %+v", body)
	}
}

func TestFastRegister_QueryEnterpriseAccount(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("action") != "search" {
			t.Errorf("action: %s", r.URL.Query().Get("action"))
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0,"status":1,"auth_code":"AC","authorizer_appid":"wxNEW"}`))
	}))
	defer srv.Close()
	f := newTestFastRegister(t, srv.URL)

	status, err := f.QueryEnterpriseAccount(context.Background(), "legal_wx", "张三")
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != 1 || status.AuthorizerAppid != "wxNEW" {
		t.Errorf("unexpected: %+v", status)
	}
	if body["legal_persona_wechat"] != "legal_wx" || body["legal_persona_name"] != "张三" {
		t.Errorf("body: %+v", body)
	}
}

func TestFastRegister_CreatePersonalAccount(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/component/fastregisterpersonalweapp") {
			t.Errorf("path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("action") != "create" {
			t.Errorf("action: %s", r.URL.Query().Get("action"))
		}
		_, _ = w.Write([]byte(`{"errcode":0,"taskid":"T123"}`))
	}))
	defer srv.Close()
	f := newTestFastRegister(t, srv.URL)

	resp, err := f.CreatePersonalAccount(context.Background(), &FastRegPersonalReq{
		IDName: "李四",
		WxUser: "user_wx",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.TaskID != "T123" {
		t.Errorf("taskid: %q", resp.TaskID)
	}
}

func TestFastRegister_QueryPersonalAccount(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("action") != "query" {
			t.Errorf("action: %s", r.URL.Query().Get("action"))
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0,"status":1,"appid":"wxPER","authorization_code":"AC"}`))
	}))
	defer srv.Close()
	f := newTestFastRegister(t, srv.URL)

	status, err := f.QueryPersonalAccount(context.Background(), "T123")
	if err != nil {
		t.Fatal(err)
	}
	if status.AppID != "wxPER" {
		t.Errorf("appid: %q", status.AppID)
	}
	if body["taskid"] != "T123" {
		t.Errorf("body: %+v", body)
	}
}

func TestFastRegister_CreateBetaAccount(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/component/fastregisterbetaweapp") {
			t.Errorf("path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("action") != "create" {
			t.Errorf("action: %s", r.URL.Query().Get("action"))
		}
		_, _ = w.Write([]byte(`{"errcode":0,"unique_id":"U1"}`))
	}))
	defer srv.Close()
	f := newTestFastRegister(t, srv.URL)

	resp, err := f.CreateBetaAccount(context.Background(), &FastRegBetaReq{
		Name:               "Acme",
		Code:               "91310000123456789X",
		CodeType:           1,
		LegalPersonaWechat: "legal_wx",
		LegalPersonaName:   "张三",
		ComponentPhone:     "13800000000",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.UniqueID != "U1" {
		t.Errorf("unique_id: %q", resp.UniqueID)
	}
}

func TestFastRegister_QueryBetaAccount(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("action") != "search" {
			t.Errorf("action: %s", r.URL.Query().Get("action"))
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0,"status":1,"appid":"wxBETA"}`))
	}))
	defer srv.Close()
	f := newTestFastRegister(t, srv.URL)

	status, err := f.QueryBetaAccount(context.Background(), "U1")
	if err != nil {
		t.Fatal(err)
	}
	if status.AppID != "wxBETA" {
		t.Errorf("appid: %q", status.AppID)
	}
	if body["unique_id"] != "U1" {
		t.Errorf("body: %+v", body)
	}
}

func TestFastRegister_GenerateAdminRebindQrcode(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/account/componentrebindadmin") {
			t.Errorf("path: %s", r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0,"taskid":"T1","qrcode_url":"https://mp.weixin.qq.com/qr/ABC"}`))
	}))
	defer srv.Close()
	f := newTestFastRegister(t, srv.URL)

	qr, err := f.GenerateAdminRebindQrcode(context.Background(), "https://example.com/rebind/cb")
	if err != nil {
		t.Fatal(err)
	}
	if qr.TaskID != "T1" || qr.QrcodeURL == "" {
		t.Errorf("unexpected: %+v", qr)
	}
	if body["redirect_uri"] != "https://example.com/rebind/cb" {
		t.Errorf("body: %+v", body)
	}
}

func TestFastRegister_CreateEnterpriseAccount_Errcode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":89247,"errmsg":"duplicate account"}`))
	}))
	defer srv.Close()
	f := newTestFastRegister(t, srv.URL)

	_, err := f.CreateEnterpriseAccount(context.Background(), &FastRegEnterpriseReq{
		Name: "Acme",
	})
	var werr *WeixinError
	if !errors.As(err, &werr) || werr.ErrCode != 89247 {
		t.Errorf("expected 89247, got %v", err)
	}
}
