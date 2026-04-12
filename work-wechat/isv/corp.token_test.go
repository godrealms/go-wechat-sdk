package isv

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestGetCorpToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/service/get_corp_token" {
			t.Errorf("path: %s", r.URL.Path)
		}
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["auth_corpid"] != "wxcorp1" || body["permanent_code"] != "PERM" {
			t.Errorf("body: %+v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "CTOK1",
			"expires_in":   7200,
		})
	}))
	defer srv.Close()

	c := newTestISVClient(t, srv.URL)
	_ = c.store.PutSuiteToken(context.Background(), "suite1", "STOK", time.Now().Add(time.Hour))

	resp, err := c.GetCorpToken(context.Background(), "wxcorp1", "PERM")
	if err != nil || resp.AccessToken != "CTOK1" {
		t.Fatalf("got %+v err=%v", resp, err)
	}
}

func TestCorpClient_AccessTokenLifecycle(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/service/get_corp_token" {
			atomic.AddInt32(&hits, 1)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "CTOK_FRESH",
				"expires_in":   7200,
			})
		}
	}))
	defer srv.Close()

	c := newTestISVClient(t, srv.URL)
	ctx := context.Background()
	_ = c.store.PutSuiteToken(ctx, "suite1", "STOK", time.Now().Add(time.Hour))
	_ = c.store.PutAuthorizer(ctx, "suite1", "wxcorp1", &AuthorizerTokens{
		CorpID:            "wxcorp1",
		PermanentCode:     "PERM",
		CorpAccessToken:   "STALE",
		CorpTokenExpireAt: time.Now().Add(-1 * time.Second), // already expired
	})

	cc := c.CorpClient("wxcorp1")
	tok, err := cc.AccessToken(ctx)
	if err != nil || tok != "CTOK_FRESH" {
		t.Fatalf("got %q err=%v", tok, err)
	}
	// cached on second call
	tok2, _ := cc.AccessToken(ctx)
	if tok2 != "CTOK_FRESH" || hits != 1 {
		t.Errorf("cache miss: hits=%d", hits)
	}
}

func TestCorpClient_AccessTokenSingleFlight(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/service/get_corp_token" {
			atomic.AddInt32(&hits, 1)
			time.Sleep(20 * time.Millisecond) // simulate network latency
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "CTOK",
				"expires_in":   7200,
			})
		}
	}))
	defer srv.Close()

	c := newTestISVClient(t, srv.URL)
	ctx := context.Background()
	_ = c.store.PutSuiteToken(ctx, "suite1", "STOK", time.Now().Add(time.Hour))
	_ = c.store.PutAuthorizer(ctx, "suite1", "wxcorp1", &AuthorizerTokens{
		CorpID:        "wxcorp1",
		PermanentCode: "PERM",
	})

	cc := c.CorpClient("wxcorp1")
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = cc.AccessToken(ctx)
		}()
	}
	wg.Wait()

	if hits != 1 {
		t.Errorf("want 1 HTTP hit (single-flight), got %d", hits)
	}
}

func TestCorpClient_Refresh(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/service/get_corp_token" {
			atomic.AddInt32(&hits, 1)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "CTOK_NEW",
				"expires_in":   7200,
			})
		}
	}))
	defer srv.Close()

	c := newTestISVClient(t, srv.URL)
	ctx := context.Background()
	_ = c.store.PutSuiteToken(ctx, "suite1", "STOK", time.Now().Add(time.Hour))
	_ = c.store.PutAuthorizer(ctx, "suite1", "wxcorp1", &AuthorizerTokens{
		CorpID:            "wxcorp1",
		PermanentCode:     "PERM",
		CorpAccessToken:   "OLD",
		CorpTokenExpireAt: time.Now().Add(time.Hour), // still valid
	})

	if err := c.CorpClient("wxcorp1").Refresh(ctx); err != nil {
		t.Fatal(err)
	}
	got, _ := c.store.GetAuthorizer(ctx, "suite1", "wxcorp1")
	if got.CorpAccessToken != "CTOK_NEW" || hits != 1 {
		t.Errorf("refresh failed: tok=%q hits=%d", got.CorpAccessToken, hits)
	}
}

func TestRefreshAll(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/service/get_corp_token" {
			atomic.AddInt32(&hits, 1)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "X",
				"expires_in":   7200,
			})
		}
	}))
	defer srv.Close()

	c := newTestISVClient(t, srv.URL)
	ctx := context.Background()
	_ = c.store.PutSuiteToken(ctx, "suite1", "STOK", time.Now().Add(time.Hour))
	_ = c.store.PutAuthorizer(ctx, "suite1", "corpA", &AuthorizerTokens{CorpID: "corpA", PermanentCode: "P1"})
	_ = c.store.PutAuthorizer(ctx, "suite1", "corpB", &AuthorizerTokens{CorpID: "corpB", PermanentCode: "P2"})

	if err := c.RefreshAll(ctx); err != nil {
		t.Fatal(err)
	}
	if hits != 2 {
		t.Errorf("want 2 HTTP hits, got %d", hits)
	}
}

func TestCorpClient_ImplementsTokenSource(t *testing.T) {
	var _ TokenSource = (*CorpClient)(nil)
}
