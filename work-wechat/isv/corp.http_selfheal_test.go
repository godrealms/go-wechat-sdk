package isv

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

// TestCorpClient_DoPost_SelfHealsOn40001 exercises the audit P1-2 self-heal for
// the ISV corp path. Unlike the flat modules (whose token cache is process-local),
// the corp_access_token is Store-backed, so Invalidate sets a per-handle
// forceRefresh flag that makes the next AccessToken re-fetch via permanent_code
// and write the fresh token back to the shared Store.
func TestCorpClient_DoPost_SelfHealsOn40001(t *testing.T) {
	var apiCalls, refreshCalls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/cgi-bin/service/get_corp_token") {
			atomic.AddInt32(&refreshCalls, 1)
			_, _ = w.Write([]byte(`{"access_token":"CTOK2","expires_in":7200}`))
			return
		}
		n := atomic.AddInt32(&apiCalls, 1)
		tok := r.URL.Query().Get("access_token")
		if n == 1 {
			if tok != "CTOK" {
				t.Errorf("first call token=%q, want seeded CTOK", tok)
			}
			_, _ = w.Write([]byte(`{"errcode":40001,"errmsg":"invalid credential"}`))
			return
		}
		if tok != "CTOK2" {
			t.Errorf("retry token=%q, want refreshed CTOK2", tok)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	if err := cc.doPost(context.Background(), "/test/post", map[string]string{"k": "v"}, nil); err != nil {
		t.Fatalf("doPost should self-heal on 40001, got %v", err)
	}
	if got := atomic.LoadInt32(&apiCalls); got != 2 {
		t.Errorf("apiCalls = %d, want 2 (initial 40001 + successful retry)", got)
	}
	if got := atomic.LoadInt32(&refreshCalls); got != 1 {
		t.Errorf("get_corp_token calls = %d, want 1 (permanent_code refresh after invalidate)", got)
	}
	// The refreshed token must be persisted back to the shared Store so other
	// instances benefit — the multi-instance point of the Store-backed path.
	toks, err := cc.parent.store.GetAuthorizer(context.Background(), "suite1", "wxcorp1")
	if err != nil {
		t.Fatalf("GetAuthorizer: %v", err)
	}
	if toks.CorpAccessToken != "CTOK2" {
		t.Errorf("Store CorpAccessToken = %q, want CTOK2 (refresh must write through)", toks.CorpAccessToken)
	}
}
