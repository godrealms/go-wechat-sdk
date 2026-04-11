package oplatform

import (
	"context"
	"errors"
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
