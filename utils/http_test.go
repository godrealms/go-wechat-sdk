package utils

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func startServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
}

func newHTTP(t *testing.T, srv *httptest.Server) *HTTP {
	t.Helper()
	return NewHTTP(srv.URL, WithTimeout(3*time.Second))
}

func TestHTTP_Get_Success(t *testing.T) {
	type Resp struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	srv := startServer(t, 200, `{"errcode":0,"errmsg":"ok"}`)
	defer srv.Close()

	h := newHTTP(t, srv)
	var result Resp
	if err := h.Get(context.Background(), "/test", nil, &result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ErrCode != 0 {
		t.Errorf("expected errcode 0, got %d", result.ErrCode)
	}
}

func TestHTTP_Get_WithQueryParams(t *testing.T) {
	var captured url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = r.URL.Query()
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	h := newHTTP(t, srv)
	q := url.Values{"access_token": {"TOKEN"}, "openid": {"oUser123"}}
	if err := h.Get(context.Background(), "/user/info", q, &struct{}{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured.Get("access_token") != "TOKEN" {
		t.Errorf("expected access_token=TOKEN, got %q", captured.Get("access_token"))
	}
}

func TestHTTP_Get_Non2xxStatus(t *testing.T) {
	srv := startServer(t, 500, `internal server error`)
	defer srv.Close()

	h := newHTTP(t, srv)
	err := h.Get(context.Background(), "/fail", nil, &struct{}{})
	if err == nil {
		t.Fatal("expected error for 500 status")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected 500 in error, got: %v", err)
	}
}

func TestHTTP_Get_NetworkError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	u := srv.URL
	srv.Close()

	h := NewHTTP(u, WithTimeout(time.Second))
	err := h.Get(context.Background(), "/path", nil, &struct{}{})
	if err == nil {
		t.Fatal("expected network error")
	}
}

func TestHTTP_Get_InvalidJSON(t *testing.T) {
	srv := startServer(t, 200, `not-json`)
	defer srv.Close()

	h := newHTTP(t, srv)
	type Resp struct{ ErrCode int `json:"errcode"` }
	var result Resp
	err := h.Get(context.Background(), "/bad-json", nil, &result)
	if err == nil {
		t.Fatal("expected unmarshal error")
	}
	if !strings.Contains(err.Error(), "unmarshal") {
		t.Errorf("expected unmarshal in error, got: %v", err)
	}
}

func TestHTTP_Post_Success(t *testing.T) {
	type ReqBody struct{ Foo string `json:"foo"` }
	type Resp struct{ Bar int `json:"bar"` }

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %q", ct)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"bar":42}`))
	}))
	defer srv.Close()

	h := newHTTP(t, srv)
	var result Resp
	if err := h.Post(context.Background(), "/post", ReqBody{Foo: "hello"}, &result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Bar != 42 {
		t.Errorf("expected bar=42, got %d", result.Bar)
	}
}

func TestHTTP_Post_Non2xxStatus(t *testing.T) {
	srv := startServer(t, 400, `{"errcode":1,"errmsg":"bad request"}`)
	defer srv.Close()

	h := newHTTP(t, srv)
	err := h.Post(context.Background(), "/fail", map[string]string{"k": "v"}, &struct{}{})
	if err == nil {
		t.Fatal("expected error for 400 status")
	}
}

func TestHTTP_Post_NetworkError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	u := srv.URL
	srv.Close()

	h := NewHTTP(u, WithTimeout(time.Second))
	err := h.Post(context.Background(), "/path", map[string]string{}, &struct{}{})
	if err == nil {
		t.Fatal("expected network error")
	}
}

func TestHTTP_Post_NilBody(t *testing.T) {
	srv := startServer(t, 200, `{"ok":true}`)
	defer srv.Close()

	h := newHTTP(t, srv)
	if err := h.Post(context.Background(), "/nil-body", nil, &struct{}{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHTTP_Put_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	h := newHTTP(t, srv)
	if err := h.Put(context.Background(), "/resource/1", map[string]string{"name": "v"}, &struct{}{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHTTP_Put_Non2xxStatus(t *testing.T) {
	srv := startServer(t, 404, `not found`)
	defer srv.Close()

	h := newHTTP(t, srv)
	err := h.Put(context.Background(), "/notfound", map[string]string{}, &struct{}{})
	if err == nil {
		t.Fatal("expected error for 404 status")
	}
}

func TestHTTP_WithHeaders(t *testing.T) {
	var gotHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("X-Custom")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	h := NewHTTP(srv.URL,
		WithTimeout(3*time.Second),
		WithHeaders(map[string]string{"X-Custom": "myvalue"}),
	)
	_ = h.Get(context.Background(), "/", nil, &struct{}{})
	if gotHeader != "myvalue" {
		t.Errorf("expected X-Custom=myvalue, got %q", gotHeader)
	}
}

func TestHTTP_SetBaseURL(t *testing.T) {
	srv := startServer(t, 200, `{}`)
	defer srv.Close()

	h := NewHTTP("http://old.example.com")
	h.SetBaseURL(srv.URL)
	if h.BaseURL != srv.URL {
		t.Errorf("SetBaseURL did not update BaseURL")
	}
}

func TestHTTP_Patch_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"updated":true}`))
	}))
	defer srv.Close()

	h := newHTTP(t, srv)
	type Resp struct {
		Updated bool `json:"updated"`
	}
	var result Resp
	if err := h.Patch(context.Background(), "/resource/1", map[string]string{"field": "value"}, &result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Updated {
		t.Error("expected updated=true")
	}
}

func TestHTTP_Patch_Non2xxStatus(t *testing.T) {
	srv := startServer(t, 422, `unprocessable entity`)
	defer srv.Close()

	h := newHTTP(t, srv)
	err := h.Patch(context.Background(), "/invalid", map[string]string{}, &struct{}{})
	if err == nil {
		t.Fatal("expected error for 422 status")
	}
	if !strings.Contains(err.Error(), "422") {
		t.Errorf("expected 422 in error, got: %v", err)
	}
}

func TestHTTP_Delete_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	h := newHTTP(t, srv)
	if err := h.Delete(context.Background(), "/resource/1", &struct{}{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHTTP_Delete_Non2xxStatus(t *testing.T) {
	srv := startServer(t, 404, `not found`)
	defer srv.Close()

	h := newHTTP(t, srv)
	err := h.Delete(context.Background(), "/missing", &struct{}{})
	if err == nil {
		t.Fatal("expected error for 404 status")
	}
}

func TestHTTP_PostForm_Success(t *testing.T) {
	var gotContentType string
	var gotFormValue string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		_ = r.ParseForm()
		gotFormValue = r.FormValue("grant_type")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"access_token":"TOKEN"}`))
	}))
	defer srv.Close()

	h := newHTTP(t, srv)
	form := url.Values{"grant_type": {"client_credentials"}, "appid": {"wx123"}}
	type Resp struct {
		AccessToken string `json:"access_token"`
	}
	var result Resp
	if err := h.PostForm(context.Background(), "/token", form, &result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AccessToken != "TOKEN" {
		t.Errorf("expected access_token=TOKEN, got %q", result.AccessToken)
	}
	if !strings.Contains(gotContentType, "application/x-www-form-urlencoded") {
		t.Errorf("expected form content-type, got %q", gotContentType)
	}
	if gotFormValue != "client_credentials" {
		t.Errorf("expected grant_type=client_credentials, got %q", gotFormValue)
	}
}

func TestHTTP_PostForm_Non2xxStatus(t *testing.T) {
	srv := startServer(t, 400, `bad request`)
	defer srv.Close()

	h := newHTTP(t, srv)
	err := h.PostForm(context.Background(), "/form", url.Values{"k": {"v"}}, &struct{}{})
	if err == nil {
		t.Fatal("expected error for 400 status")
	}
}

func TestHTTP_PostForm_NetworkError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	u := srv.URL
	srv.Close()

	h := NewHTTP(u, WithTimeout(time.Second))
	err := h.PostForm(context.Background(), "/form", url.Values{}, &struct{}{})
	if err == nil {
		t.Fatal("expected network error")
	}
}
