package channels

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

// TestDoPost_SelfHealsOn40001 exercises the utils.DoWithTokenRetry wiring added
// for audit P1-2: a 40001 response invalidates the cached access_token, a fresh
// one is fetched, and the request is retried once and succeeds. This flat module
// is representative — channels, mini-program, mini-game, mini-store, aispeech,
// and xiaowei all share the identical helper shape.
func TestDoPost_SelfHealsOn40001(t *testing.T) {
	var apiCalls, tokenCalls int32
	tokens := []string{"T1", "T2"}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/cgi-bin/token") {
			i := atomic.AddInt32(&tokenCalls, 1) - 1
			if int(i) >= len(tokens) {
				i = int32(len(tokens)) - 1
			}
			_, _ = w.Write([]byte(`{"access_token":"` + tokens[i] + `","expires_in":7200}`))
			return
		}
		n := atomic.AddInt32(&apiCalls, 1)
		tok := r.URL.Query().Get("access_token")
		if n == 1 {
			if tok != "T1" {
				t.Errorf("first call token=%q, want T1", tok)
			}
			_, _ = w.Write([]byte(`{"errcode":40001,"errmsg":"invalid credential"}`))
			return
		}
		if tok != "T2" {
			t.Errorf("retry token=%q, want T2 (fresh token after invalidate)", tok)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	if err := c.doPost(context.Background(), "/some/path", map[string]any{"k": "v"}, nil); err != nil {
		t.Fatalf("doPost should self-heal on 40001, got %v", err)
	}
	if got := atomic.LoadInt32(&apiCalls); got != 2 {
		t.Errorf("apiCalls = %d, want 2 (initial 40001 + successful retry)", got)
	}
	if got := atomic.LoadInt32(&tokenCalls); got != 2 {
		t.Errorf("tokenCalls = %d, want 2 (initial fetch + refetch after invalidate)", got)
	}
}

// TestDoPost_NoRetryOnNonTokenErrcode verifies non-token errcodes are NOT
// retried — only 40001/40014/42001/42007 trigger the self-heal, so an ordinary
// business error surfaces immediately after a single request.
func TestDoPost_NoRetryOnNonTokenErrcode(t *testing.T) {
	var apiCalls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/cgi-bin/token") {
			_, _ = w.Write([]byte(`{"access_token":"T","expires_in":7200}`))
			return
		}
		atomic.AddInt32(&apiCalls, 1)
		_, _ = w.Write([]byte(`{"errcode":48001,"errmsg":"api unauthorized"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	err := c.doPost(context.Background(), "/some/path", map[string]any{"k": "v"}, nil)
	var ae *APIError
	if !errors.As(err, &ae) || ae.ErrCode != 48001 {
		t.Fatalf("want *APIError 48001, got %v", err)
	}
	if got := atomic.LoadInt32(&apiCalls); got != 1 {
		t.Errorf("apiCalls = %d, want 1 (no retry on non-token errcode)", got)
	}
}
