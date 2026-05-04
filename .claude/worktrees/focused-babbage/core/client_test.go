package core

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"
)

func newTestTokenServer(t *testing.T, callCount *int32) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(callCount, 1)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "test-token-12345",
			"expires_in":   7200,
		})
	}))
}

func TestBaseClient_GetAccessToken_CachesToken(t *testing.T) {
	var callCount int32
	srv := newTestTokenServer(t, &callCount)
	defer srv.Close()

	cfg := &BaseConfig{AppId: "wx_app_id", AppSecret: "app_secret"}
	client := NewBaseClient(context.Background(), cfg, srv.URL, "/token", "GET")

	token1 := client.GetAccessToken()
	token2 := client.GetAccessToken()

	if token1 == "" {
		t.Fatal("expected non-empty token on first call")
	}
	if token1 != token2 {
		t.Errorf("expected same token on second call; got %q and %q", token1, token2)
	}
	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("expected 1 HTTP call, got %d", atomic.LoadInt32(&callCount))
	}
}

func TestBaseClient_GetAccessToken_RefreshesExpiredToken(t *testing.T) {
	var callCount int32
	srv := newTestTokenServer(t, &callCount)
	defer srv.Close()

	cfg := &BaseConfig{AppId: "wx_app_id", AppSecret: "app_secret"}
	client := NewBaseClient(context.Background(), cfg, srv.URL, "/token", "GET")

	// Pre-set an already-expired token
	client.SetAccessToken(&AccessToken{
		AccessToken: "expired-token",
		ExpiresIn:   time.Now().Unix() - 60,
	})

	token := client.GetAccessToken()
	if token == "expired-token" {
		t.Error("expected token to be refreshed, but got the stale token")
	}
	if token == "" {
		t.Fatal("expected a new non-empty token after refresh")
	}
	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("expected 1 HTTP call for refresh, got %d", atomic.LoadInt32(&callCount))
	}
}

func TestBaseClient_PostTokenMethod(t *testing.T) {
	var gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "post-token-67890",
			"expires_in":   7200,
		})
	}))
	defer srv.Close()

	cfg := &BaseConfig{AppId: "wx_app_id", AppSecret: "app_secret"}
	client := NewBaseClient(context.Background(), cfg, srv.URL, "/token", "POST")

	token := client.GetAccessToken()
	if token == "" {
		t.Fatal("expected non-empty token from POST method client")
	}
	if gotMethod != http.MethodPost {
		t.Errorf("expected HTTP method POST, got %s", gotMethod)
	}
}

func TestBaseClient_TokenQuery_MergesExtra(t *testing.T) {
	var callCount int32
	srv := newTestTokenServer(t, &callCount)
	defer srv.Close()

	cfg := &BaseConfig{AppId: "wx_app_id", AppSecret: "app_secret"}
	client := NewBaseClient(context.Background(), cfg, srv.URL, "/token", "GET")

	q := client.TokenQuery(url.Values{
		"openid": {"user_open_id_123"},
		"lang":   {"zh_CN"},
	})

	if len(q["access_token"]) == 0 || q["access_token"][0] == "" {
		t.Error("expected non-empty access_token in query values")
	}
	if q.Get("openid") != "user_open_id_123" {
		t.Errorf("expected openid=user_open_id_123, got %q", q.Get("openid"))
	}
	if q.Get("lang") != "zh_CN" {
		t.Errorf("expected lang=zh_CN, got %q", q.Get("lang"))
	}
}

func TestBaseClient_SetAccessToken(t *testing.T) {
	var callCount int32
	srv := newTestTokenServer(t, &callCount)
	defer srv.Close()

	cfg := &BaseConfig{AppId: "wx_app_id", AppSecret: "app_secret"}
	client := NewBaseClient(context.Background(), cfg, srv.URL, "/token", "GET")

	client.SetAccessToken(&AccessToken{
		AccessToken: "manually-set-token",
		ExpiresIn:   time.Now().Unix() + 7000,
	})

	token := client.GetAccessToken()
	if token != "manually-set-token" {
		t.Errorf("expected manually-set-token, got %q", token)
	}
	if atomic.LoadInt32(&callCount) != 0 {
		t.Errorf("expected 0 HTTP calls (token was pre-set), got %d", atomic.LoadInt32(&callCount))
	}
}
