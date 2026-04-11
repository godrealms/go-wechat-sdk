package oplatform

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	mini_program "github.com/godrealms/go-wechat-sdk/mini-program"
	"github.com/godrealms/go-wechat-sdk/offiaccount"
	"github.com/godrealms/go-wechat-sdk/utils"
)

func TestAuthorizerClient_AccessToken_LazyAndCaches(t *testing.T) {
	var refreshCalls int32
	mux := http.NewServeMux()
	mux.HandleFunc("/cgi-bin/component/api_component_token", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"component_access_token":"CTOK","expires_in":7200}`))
	})
	mux.HandleFunc("/cgi-bin/component/api_authorizer_token", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&refreshCalls, 1)
		_, _ = w.Write([]byte(`{"authorizer_access_token":"A1","authorizer_refresh_token":"R1","expires_in":7200}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	store := NewMemoryStore()
	_ = store.SetVerifyTicket(context.Background(), "TICKET")
	_ = store.SetAuthorizer(context.Background(), "wxA", AuthorizerTokens{
		AccessToken:  "old",
		RefreshToken: "R0",
		ExpireAt:     time.Now().Add(-time.Minute),
	})
	c := newTestClient(t, srv.URL, WithStore(store))
	auth := c.Authorizer("wxA")

	for i := 0; i < 3; i++ {
		tok, err := auth.AccessToken(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if tok != "A1" {
			t.Errorf("got %q", tok)
		}
	}
	if got := atomic.LoadInt32(&refreshCalls); got != 1 {
		t.Errorf("expected 1 refresh, got %d", got)
	}
}

func TestAuthorizerClient_AccessToken_RefreshRevoked(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/cgi-bin/component/api_component_token", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"component_access_token":"CTOK","expires_in":7200}`))
	})
	mux.HandleFunc("/cgi-bin/component/api_authorizer_token", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":61023,"errmsg":"invalid refresh_token"}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	store := NewMemoryStore()
	_ = store.SetVerifyTicket(context.Background(), "TICKET")
	_ = store.SetAuthorizer(context.Background(), "wxA", AuthorizerTokens{
		AccessToken: "old", RefreshToken: "Rbad", ExpireAt: time.Now().Add(-time.Minute),
	})
	c := newTestClient(t, srv.URL, WithStore(store))
	auth := c.Authorizer("wxA")

	_, err := auth.AccessToken(context.Background())
	if !errors.Is(err, ErrAuthorizerRevoked) {
		t.Errorf("expected ErrAuthorizerRevoked, got %v", err)
	}
}

func TestAuthorizerClient_Refresh_ForcesFetch(t *testing.T) {
	var refreshCalls int32
	mux := http.NewServeMux()
	mux.HandleFunc("/cgi-bin/component/api_component_token", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"component_access_token":"CTOK","expires_in":7200}`))
	})
	mux.HandleFunc("/cgi-bin/component/api_authorizer_token", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&refreshCalls, 1)
		_, _ = w.Write([]byte(`{"authorizer_access_token":"A2","authorizer_refresh_token":"R2","expires_in":7200}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	store := NewMemoryStore()
	_ = store.SetVerifyTicket(context.Background(), "TICKET")
	_ = store.SetAuthorizer(context.Background(), "wxA", AuthorizerTokens{
		AccessToken: "cached", RefreshToken: "R0", ExpireAt: time.Now().Add(time.Hour),
	})
	c := newTestClient(t, srv.URL, WithStore(store))
	auth := c.Authorizer("wxA")

	tok, err := auth.AccessToken(context.Background())
	if err != nil || tok != "cached" {
		t.Fatalf("expected cached, got %q err=%v", tok, err)
	}
	if err := auth.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := atomic.LoadInt32(&refreshCalls); got != 1 {
		t.Errorf("expected 1 refresh, got %d", got)
	}
	stored, _ := store.GetAuthorizer(context.Background(), "wxA")
	if stored.AccessToken != "A2" || stored.RefreshToken != "R2" {
		t.Errorf("store mismatch: %+v", stored)
	}
}

func TestAuthorizerClient_OffiaccountClient_UsesAuthorizerToken(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/cgi-bin/component/api_component_token", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"component_access_token":"CTOK","expires_in":7200}`))
	})
	mux.HandleFunc("/cgi-bin/component/api_authorizer_token", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"authorizer_access_token":"AUTH_A","authorizer_refresh_token":"R","expires_in":7200}`))
	})
	mux.HandleFunc("/cgi-bin/token", func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("cgi-bin/token should not be called through TokenSource")
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	store := NewMemoryStore()
	_ = store.SetVerifyTicket(context.Background(), "TICKET")
	_ = store.SetAuthorizer(context.Background(), "wxBiz", AuthorizerTokens{
		AccessToken: "x", RefreshToken: "R", ExpireAt: time.Now().Add(-time.Minute),
	})
	c := newTestClient(t, srv.URL, WithStore(store))
	auth := c.Authorizer("wxBiz")

	off := auth.OffiaccountClient(offiaccount.WithHTTPClient(utils.NewHTTP(srv.URL, utils.WithTimeout(time.Second*3))))
	tok, err := off.AccessTokenE(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if tok != "AUTH_A" {
		t.Errorf("offiaccount token should come from oplatform, got %q", tok)
	}
}

func TestAuthorizerClient_MiniProgramClient_UsesAuthorizerToken(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/cgi-bin/component/api_component_token", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"component_access_token":"CTOK","expires_in":7200}`))
	})
	mux.HandleFunc("/cgi-bin/component/api_authorizer_token", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"authorizer_access_token":"AUTH_MP","authorizer_refresh_token":"R","expires_in":7200}`))
	})
	mux.HandleFunc("/cgi-bin/token", func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("cgi-bin/token should not be called")
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	store := NewMemoryStore()
	_ = store.SetVerifyTicket(context.Background(), "TICKET")
	_ = store.SetAuthorizer(context.Background(), "wxMP", AuthorizerTokens{
		AccessToken: "x", RefreshToken: "R", ExpireAt: time.Now().Add(-time.Minute),
	})
	c := newTestClient(t, srv.URL, WithStore(store))
	auth := c.Authorizer("wxMP")

	mp, err := auth.MiniProgramClient(mini_program.WithHTTP(utils.NewHTTP(srv.URL, utils.WithTimeout(time.Second*3))))
	if err != nil {
		t.Fatal(err)
	}
	tok, err := mp.AccessToken(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if tok != "AUTH_MP" {
		t.Errorf("mini-program token should come from oplatform, got %q", tok)
	}
}

func TestClient_RefreshAll(t *testing.T) {
	var refreshCalls int32
	mux := http.NewServeMux()
	mux.HandleFunc("/cgi-bin/component/api_component_token", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"component_access_token":"CTOK","expires_in":7200}`))
	})
	mux.HandleFunc("/cgi-bin/component/api_authorizer_token", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&refreshCalls, 1)
		_, _ = w.Write([]byte(`{"authorizer_access_token":"NEW","authorizer_refresh_token":"R","expires_in":7200}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	store := NewMemoryStore()
	_ = store.SetVerifyTicket(context.Background(), "TICKET")
	_ = store.SetAuthorizer(context.Background(), "wxA", AuthorizerTokens{
		AccessToken: "old", RefreshToken: "R", ExpireAt: time.Now().Add(time.Hour),
	})
	_ = store.SetAuthorizer(context.Background(), "wxB", AuthorizerTokens{
		AccessToken: "old", RefreshToken: "R", ExpireAt: time.Now().Add(time.Hour),
	})
	c := newTestClient(t, srv.URL, WithStore(store))

	if err := c.RefreshAll(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := atomic.LoadInt32(&refreshCalls); got != 2 {
		t.Errorf("expected 2 refreshes, got %d", got)
	}
}
