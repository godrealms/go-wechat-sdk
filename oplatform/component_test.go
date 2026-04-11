package oplatform

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

func testConfig() Config {
	return Config{
		ComponentAppID:     "wxcomp",
		ComponentAppSecret: "secret",
		Token:              "tk",
		EncodingAESKey:     "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQ", // 43 chars
	}
}

func newTestClient(t *testing.T, baseURL string, opts ...Option) *Client {
	t.Helper()
	opts = append(opts, WithHTTP(utils.NewHTTP(baseURL, utils.WithTimeout(time.Second*3))))
	c, err := NewClient(testConfig(), opts...)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestClient_ComponentAccessToken_LazyAndCaches(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/component/api_component_token") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		atomic.AddInt32(&calls, 1)
		_, _ = w.Write([]byte(`{"component_access_token":"CTOK","expires_in":7200}`))
	}))
	defer srv.Close()

	store := NewMemoryStore()
	_ = store.SetVerifyTicket(context.Background(), "TICKET")
	c := newTestClient(t, srv.URL, WithStore(store))

	for i := 0; i < 3; i++ {
		tok, err := c.ComponentAccessToken(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if tok != "CTOK" {
			t.Errorf("got %q", tok)
		}
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("expected 1 fetch, got %d", got)
	}
}

func TestClient_ComponentAccessToken_MissingTicket(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("should not reach the server when ticket is missing: %s", r.URL.Path)
	}))
	defer srv.Close()
	c := newTestClient(t, srv.URL)

	_, err := c.ComponentAccessToken(context.Background())
	if !errors.Is(err, ErrVerifyTicketMissing) {
		t.Errorf("expected ErrVerifyTicketMissing, got %v", err)
	}
}

func TestClient_ComponentAccessToken_WeixinError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":40013,"errmsg":"invalid appid"}`))
	}))
	defer srv.Close()
	store := NewMemoryStore()
	_ = store.SetVerifyTicket(context.Background(), "TICKET")
	c := newTestClient(t, srv.URL, WithStore(store))

	_, err := c.ComponentAccessToken(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	var werr *WeixinError
	if !errors.As(err, &werr) || werr.ErrCode != 40013 {
		t.Errorf("expected WeixinError 40013, got %v", err)
	}
}

func TestClient_RefreshComponentToken_ForcesFetch(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		_, _ = w.Write([]byte(`{"component_access_token":"CTOK","expires_in":7200}`))
	}))
	defer srv.Close()
	store := NewMemoryStore()
	_ = store.SetVerifyTicket(context.Background(), "TICKET")
	c := newTestClient(t, srv.URL, WithStore(store))

	if _, err := c.ComponentAccessToken(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := c.RefreshComponentToken(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Errorf("expected 2 fetches after forced refresh, got %d", got)
	}
}
