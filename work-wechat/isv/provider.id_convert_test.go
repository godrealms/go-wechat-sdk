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

// newTestISVClientWithProvider 构造启用了 provider 字段的 Client。
func newTestISVClientWithProvider(t *testing.T, baseURL string) *Client {
	t.Helper()
	cfg := testConfig()
	cfg.ProviderCorpID = "wxprov"
	cfg.ProviderSecret = "PSEC"
	c, err := NewClient(cfg, WithBaseURL(baseURL))
	if err != nil {
		t.Fatal(err)
	}
	_ = c.store.PutSuiteTicket(context.Background(), "suite1", "TICKET")
	return c
}

func TestCorpIDToOpenCorpID(t *testing.T) {
	var providerHits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/service/get_provider_token":
			atomic.AddInt32(&providerHits, 1)
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["corpid"] != "wxprov" || body["provider_secret"] != "PSEC" {
				t.Errorf("provider token body: %+v", body)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"provider_access_token": "PTOK",
				"expires_in":            7200,
			})
		case "/cgi-bin/service/corpid_to_opencorpid":
			if got := r.URL.Query().Get("provider_access_token"); got != "PTOK" {
				t.Errorf("token query: %q", got)
			}
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["corpid"] != "wxcorp1" {
				t.Errorf("body: %+v", body)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"open_corpid": "openWx1",
			})
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := newTestISVClientWithProvider(t, srv.URL)

	got, err := c.CorpIDToOpenCorpID(context.Background(), "wxcorp1")
	if err != nil || got != "openWx1" {
		t.Fatalf("got %q err=%v", got, err)
	}
	// second call — provider token cached
	got2, _ := c.CorpIDToOpenCorpID(context.Background(), "wxcorp1")
	if got2 != "openWx1" || providerHits != 1 {
		t.Errorf("provider cache failed: hits=%d", providerHits)
	}
}

func TestUserIDToOpenUserID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/service/get_provider_token":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"provider_access_token": "PTOK",
				"expires_in":            7200,
			})
		case "/cgi-bin/service/batch/userid_to_openuserid":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"open_userid_list": []map[string]string{
					{"userid": "u1", "open_userid": "o1"},
				},
				"invalid_userid_list": []string{"u_bad"},
			})
		}
	}))
	defer srv.Close()

	c := newTestISVClientWithProvider(t, srv.URL)
	resp, err := c.UserIDToOpenUserID(context.Background(), "wxcorp1", []string{"u1", "u_bad"})
	if err != nil || len(resp.OpenUserIDList) != 1 || resp.InvalidUserIDList[0] != "u_bad" {
		t.Fatalf("got %+v err=%v", resp, err)
	}
}

func TestProviderNotConfigured(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	_, err := c.CorpIDToOpenCorpID(context.Background(), "wxcorp1")
	if !errors.Is(err, ErrProviderCorpIDMissing) && !errors.Is(err, ErrProviderSecretMissing) {
		t.Fatalf("want provider missing error, got %v", err)
	}
}

func TestProviderTokenExpiredRefresh(t *testing.T) {
	var providerHits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/service/get_provider_token" {
			atomic.AddInt32(&providerHits, 1)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"provider_access_token": "FRESH",
				"expires_in":            7200,
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"open_corpid": "o1"})
	}))
	defer srv.Close()

	c := newTestISVClientWithProvider(t, srv.URL)
	_ = c.store.PutProviderToken(context.Background(), "suite1", "STALE", time.Now().Add(-1*time.Second))

	if _, err := c.CorpIDToOpenCorpID(context.Background(), "wxcorp1"); err != nil {
		t.Fatal(err)
	}
	if providerHits != 1 {
		t.Errorf("want 1 provider hit, got %d", providerHits)
	}
}
