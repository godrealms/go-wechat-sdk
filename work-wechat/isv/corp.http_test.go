package isv

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// newTestCorpClient creates a CorpClient with a pre-seeded valid corp_access_token.
// No httptest mock needed for token fetch — the token is directly in the Store.
func newTestCorpClient(t *testing.T, baseURL string) *CorpClient {
	t.Helper()
	c := newTestISVClient(t, baseURL)
	ctx := context.Background()
	_ = c.store.PutSuiteToken(ctx, "suite1", "STOK", time.Now().Add(time.Hour))
	_ = c.store.PutAuthorizer(ctx, "suite1", "wxcorp1", &AuthorizerTokens{
		CorpID:            "wxcorp1",
		PermanentCode:     "PERM",
		CorpAccessToken:   "CTOK",
		CorpTokenExpireAt: time.Now().Add(time.Hour),
	})
	return c.CorpClient("wxcorp1")
}

func TestCorpClient_DoPost_TokenInjection(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["key"] != "val" {
			t.Errorf("body: %+v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"result": "ok",
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	var resp struct {
		Result string `json:"result"`
	}
	err := cc.doPost(context.Background(), "/test/post", map[string]string{"key": "val"}, &resp)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Result != "ok" {
		t.Errorf("resp: %+v", resp)
	}
}

func TestCorpClient_DoGet_TokenInjection(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if got := r.URL.Query().Get("extra"); got != "123" {
			t.Errorf("extra: %q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": "hello",
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	var resp struct {
		Data string `json:"data"`
	}
	extra := map[string][]string{"extra": {"123"}}
	err := cc.doGet(context.Background(), "/test/get", extra, &resp)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Data != "hello" {
		t.Errorf("resp: %+v", resp)
	}
}
