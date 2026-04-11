package oplatform

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
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
