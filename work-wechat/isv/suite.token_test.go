package isv

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// newTestISVClient 建立一个指向 baseURL 的 Client,store 中预种子 suite_ticket。
func newTestISVClient(t *testing.T, baseURL string) *Client {
	t.Helper()
	c, err := NewClient(testConfig(), WithBaseURL(baseURL))
	if err != nil {
		t.Fatal(err)
	}
	_ = c.store.PutSuiteTicket(context.Background(), "suite1", "TICKET")
	return c
}

func TestGetSuiteAccessToken_FirstAndCached(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		if r.URL.Path != "/cgi-bin/service/get_suite_token" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["suite_id"] != "suite1" || body["suite_ticket"] != "TICKET" {
			t.Errorf("unexpected body: %+v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"suite_access_token": "STOK",
			"expires_in":         7200,
		})
	}))
	defer srv.Close()

	c := newTestISVClient(t, srv.URL)
	ctx := context.Background()

	tok, err := c.GetSuiteAccessToken(ctx)
	if err != nil || tok != "STOK" {
		t.Fatalf("got %q err=%v", tok, err)
	}
	// second call — cached, no HTTP hit
	tok2, err := c.GetSuiteAccessToken(ctx)
	if err != nil || tok2 != "STOK" {
		t.Fatalf("got %q err=%v", tok2, err)
	}
	if hits != 1 {
		t.Errorf("want 1 HTTP hit, got %d", hits)
	}
}

func TestGetSuiteAccessToken_MissingTicket(t *testing.T) {
	c, err := NewClient(testConfig())
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.GetSuiteAccessToken(context.Background())
	if !errors.Is(err, ErrSuiteTicketMissing) {
		t.Fatalf("want ErrSuiteTicketMissing, got %v", err)
	}
}

func TestGetSuiteAccessToken_ExpiredRefreshes(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"suite_access_token": "FRESH",
			"expires_in":         7200,
		})
	}))
	defer srv.Close()

	c := newTestISVClient(t, srv.URL)
	// seed an expired token
	_ = c.store.PutSuiteToken(context.Background(), "suite1", "STALE", time.Now().Add(-1*time.Second))

	tok, err := c.GetSuiteAccessToken(context.Background())
	if err != nil || tok != "FRESH" {
		t.Fatalf("got %q err=%v", tok, err)
	}
	if hits != 1 {
		t.Errorf("want 1 HTTP hit, got %d", hits)
	}
}

func TestRefreshSuiteToken_Explicit(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"suite_access_token": "NEW",
			"expires_in":         7200,
		})
	}))
	defer srv.Close()

	c := newTestISVClient(t, srv.URL)
	ctx := context.Background()
	// pre-seed a still-valid token; RefreshSuiteToken should ignore cache.
	_ = c.store.PutSuiteToken(ctx, "suite1", "OLD", time.Now().Add(time.Hour))

	if err := c.RefreshSuiteToken(ctx); err != nil {
		t.Fatal(err)
	}
	tok, _, _ := c.store.GetSuiteToken(ctx, "suite1")
	if tok != "NEW" {
		t.Errorf("want NEW, got %q", tok)
	}
	if hits != 1 {
		t.Errorf("want 1 HTTP hit, got %d", hits)
	}
}

func TestGetSuiteAccessToken_WeixinError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"errcode": 40001,
			"errmsg":  "invalid credential",
		})
	}))
	defer srv.Close()

	c := newTestISVClient(t, srv.URL)
	_, err := c.GetSuiteAccessToken(context.Background())
	var we *WeixinError
	if !errors.As(err, &we) || we.ErrCode != 40001 {
		t.Fatalf("want *WeixinError 40001, got %v", err)
	}
}
