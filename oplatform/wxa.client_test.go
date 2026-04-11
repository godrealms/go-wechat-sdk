package oplatform

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// Shared test helper used by wxa.*_test.go.
func newTestWxaAdmin(t *testing.T, baseURL string) *WxaAdminClient {
	t.Helper()
	store := NewMemoryStore()
	_ = store.SetVerifyTicket(context.Background(), "TICKET")
	_ = store.SetAuthorizer(context.Background(), "wxBiz", AuthorizerTokens{
		AccessToken:  "ATOK",
		RefreshToken: "R",
		ExpireAt:     time.Now().Add(time.Hour),
	})
	c := newTestClient(t, baseURL, WithStore(store))
	return c.Authorizer("wxBiz").WxaAdmin()
}

func TestWxaAdmin_DoPost_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/fake_endpoint") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("access_token") != "ATOK" {
			t.Errorf("missing access_token, got %q", r.URL.Query().Get("access_token"))
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok","value":42}`))
	}))
	defer srv.Close()

	w := newTestWxaAdmin(t, srv.URL)
	var out struct {
		Value int `json:"value"`
	}
	if err := w.doPost(context.Background(), "/wxa/fake_endpoint", map[string]string{"k": "v"}, &out); err != nil {
		t.Fatal(err)
	}
	if out.Value != 42 {
		t.Errorf("got %d, want 42", out.Value)
	}
}

func TestWxaAdmin_DoPost_Errcode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":85013,"errmsg":"version not exist"}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	err := w.doPost(context.Background(), "/wxa/fake_endpoint", nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var werr *WeixinError
	if !errors.As(err, &werr) || werr.ErrCode != 85013 {
		t.Errorf("expected WeixinError 85013, got %v", err)
	}
}

func TestWxaAdmin_DoPost_NilOut(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)
	if err := w.doPost(context.Background(), "/wxa/fake", nil, nil); err != nil {
		t.Fatal(err)
	}
}

func TestWxaAdmin_DoGet_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("access_token") != "ATOK" {
			t.Errorf("missing access_token")
		}
		if r.URL.Query().Get("foo") != "bar" {
			t.Errorf("missing foo=bar, got %q", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"name":"zzz"}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	var out struct {
		Name string `json:"name"`
	}
	if err := w.doGet(context.Background(), "/wxa/fake_get", url.Values{"foo": {"bar"}}, &out); err != nil {
		t.Fatal(err)
	}
	if out.Name != "zzz" {
		t.Errorf("got %q", out.Name)
	}
}

func TestWxaAdmin_DoGetRaw_Binary(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("access_token") != "ATOK" {
			t.Errorf("missing access_token")
		}
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = w.Write([]byte{0xFF, 0xD8, 0xFF, 0xE0}) // JPEG magic
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	body, ct, err := w.doGetRaw(context.Background(), "/wxa/fake_binary", url.Values{"path": {"pages/index"}})
	if err != nil {
		t.Fatal(err)
	}
	if ct != "image/jpeg" {
		t.Errorf("content-type: %q", ct)
	}
	if len(body) != 4 || body[0] != 0xFF {
		t.Errorf("body mismatch: %v", body)
	}
}
