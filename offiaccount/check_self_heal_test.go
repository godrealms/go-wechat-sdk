package offiaccount

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

// TestIsTokenExpired_TableDriven covers all four errcodes recognised as
// expired-token plus the negative cases (other errcodes, non-WeixinError,
// nil error, wrapped *WeixinError).
func TestIsTokenExpired_TableDriven(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"plain string error", errors.New("boom"), false},
		{"WeixinError 40001", &WeixinError{ErrCode: 40001}, true},
		{"WeixinError 40014", &WeixinError{ErrCode: 40014}, true},
		{"WeixinError 42001", &WeixinError{ErrCode: 42001}, true},
		{"WeixinError 42007", &WeixinError{ErrCode: 42007}, true},
		{"WeixinError 40013 (invalid appid, NOT token)", &WeixinError{ErrCode: 40013}, false},
		{"WeixinError 0 (success)", &WeixinError{ErrCode: 0}, false},
		{"wrapped 40001",
			&wrappedErr{inner: &WeixinError{ErrCode: 40001}},
			true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsTokenExpired(tt.err); got != tt.want {
				t.Errorf("IsTokenExpired(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

type wrappedErr struct{ inner error }

func (w *wrappedErr) Error() string { return "wrapped: " + w.inner.Error() }
func (w *wrappedErr) Unwrap() error { return w.inner }

// TestClient_Invalidate_ClearsInProcessCache verifies Invalidate resets the
// internal token+expiresAt so the next AccessTokenE call hits upstream again.
func TestClient_Invalidate_ClearsInProcessCache(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		_, _ = w.Write([]byte(`{"access_token":"T","expires_in":7200}`))
	}))
	defer srv.Close()

	c := newClientWithBaseURL(srv.URL, &Config{AppId: "a", AppSecret: "b"})
	if _, err := c.AccessTokenE(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected 1 call before invalidate, got %d", got)
	}
	c.Invalidate()
	if _, err := c.AccessTokenE(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Errorf("expected 2 calls after invalidate, got %d", got)
	}
}

// invalidatableTokenSource is a TokenSource that also implements Invalidator.
type invalidatableTokenSource struct {
	token            string
	calls            int32
	invalidateCalls  int32
	tokensPerSession []string // optional rotation for retry tests
}

func (s *invalidatableTokenSource) AccessToken(ctx context.Context) (string, error) {
	idx := atomic.AddInt32(&s.calls, 1) - 1
	if len(s.tokensPerSession) > 0 {
		i := int(idx)
		if i >= len(s.tokensPerSession) {
			i = len(s.tokensPerSession) - 1
		}
		return s.tokensPerSession[i], nil
	}
	return s.token, nil
}

func (s *invalidatableTokenSource) Invalidate() {
	atomic.AddInt32(&s.invalidateCalls, 1)
}

// TestClient_Invalidate_DelegatesToTokenSource verifies that when a TokenSource
// implements Invalidator, Client.Invalidate delegates to it instead of clearing
// its internal cache (which is bypassed by the TokenSource anyway).
func TestClient_Invalidate_DelegatesToTokenSource(t *testing.T) {
	src := &invalidatableTokenSource{token: "T"}
	c := NewClient(context.Background(), &Config{AppId: "a"}, WithTokenSource(src))
	c.Invalidate()
	if got := atomic.LoadInt32(&src.invalidateCalls); got != 1 {
		t.Errorf("expected 1 Invalidate call on token source, got %d", got)
	}
}

// TestClient_Invalidate_TokenSourceWithoutInvalidator must not panic and must
// NOT touch the internal cache when the source can't be invalidated (since
// the internal cache is bypassed when a TokenSource is set anyway).
func TestClient_Invalidate_TokenSourceWithoutInvalidator(t *testing.T) {
	src := &fakeTokenSource{token: "T"}
	c := NewClient(context.Background(), &Config{AppId: "a"}, WithTokenSource(src))
	c.Invalidate() // must not panic
}

// TestDoGet_RetriesOnce40001 exercises the core self-heal path: first call
// returns errcode=40001, doGet invalidates the cached token, fetches a fresh
// one, replaces access_token in the query, and retries successfully.
func TestDoGet_RetriesOnce40001(t *testing.T) {
	var apiCalls int32
	var tokenCalls int32
	tokens := []string{"T1", "T2"}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/cgi-bin/token"):
			i := atomic.AddInt32(&tokenCalls, 1) - 1
			if int(i) >= len(tokens) {
				i = int32(len(tokens)) - 1
			}
			_, _ = w.Write([]byte(`{"access_token":"` + tokens[i] + `","expires_in":7200}`))
		case strings.HasSuffix(r.URL.Path, "/cgi-bin/test"):
			n := atomic.AddInt32(&apiCalls, 1)
			tok := r.URL.Query().Get("access_token")
			if n == 1 {
				if tok != "T1" {
					t.Errorf("first call should use T1, got %q", tok)
				}
				_, _ = w.Write([]byte(`{"errcode":40001,"errmsg":"invalid credential"}`))
				return
			}
			if tok != "T2" {
				t.Errorf("retry should use T2, got %q", tok)
			}
			_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := newClientWithBaseURL(srv.URL, &Config{AppId: "a", AppSecret: "b"})
	tok, err := c.AccessTokenE(context.Background())
	if err != nil || tok != "T1" {
		t.Fatalf("setup AccessTokenE: tok=%q err=%v", tok, err)
	}
	var result Resp
	err = c.doGet(context.Background(), "/cgi-bin/test", url.Values{"access_token": {"T1"}}, &result)
	if err != nil {
		t.Fatalf("doGet should self-heal, got %v", err)
	}
	if got := atomic.LoadInt32(&apiCalls); got != 2 {
		t.Errorf("expected 2 API calls (40001 + retry), got %d", got)
	}
	if got := atomic.LoadInt32(&tokenCalls); got != 2 {
		t.Errorf("expected 2 token fetches (initial + invalidate-refresh), got %d", got)
	}
}

// TestDoPost_RetriesOnce40001 verifies the path-embedded access_token is
// correctly replaced on retry. doPost callers stuff token into the URL query
// string before calling, unlike doGet which passes it as a separate url.Values.
func TestDoPost_RetriesOnce40001(t *testing.T) {
	var apiCalls int32
	var tokenCalls int32
	tokens := []string{"T1", "T2"}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/cgi-bin/token"):
			i := atomic.AddInt32(&tokenCalls, 1) - 1
			if int(i) >= len(tokens) {
				i = int32(len(tokens)) - 1
			}
			_, _ = w.Write([]byte(`{"access_token":"` + tokens[i] + `","expires_in":7200}`))
		case strings.HasSuffix(r.URL.Path, "/cgi-bin/test"):
			n := atomic.AddInt32(&apiCalls, 1)
			tok := r.URL.Query().Get("access_token")
			if n == 1 {
				if tok != "T1" {
					t.Errorf("first call token=%q want T1", tok)
				}
				_, _ = w.Write([]byte(`{"errcode":42001,"errmsg":"access_token expired"}`))
				return
			}
			if tok != "T2" {
				t.Errorf("retry token=%q want T2", tok)
			}
			_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
		}
	}))
	defer srv.Close()

	c := newClientWithBaseURL(srv.URL, &Config{AppId: "a", AppSecret: "b"})
	if _, err := c.AccessTokenE(context.Background()); err != nil {
		t.Fatal(err)
	}
	var result Resp
	path := "/cgi-bin/test?access_token=T1"
	err := c.doPost(context.Background(), path, map[string]any{"x": 1}, &result)
	if err != nil {
		t.Fatalf("doPost should self-heal, got %v", err)
	}
	if got := atomic.LoadInt32(&apiCalls); got != 2 {
		t.Errorf("expected 2 API calls, got %d", got)
	}
}

// TestDoGet_NoRetryOnNon40001 verifies that the self-heal path does NOT trigger
// for arbitrary errcodes — only token-expired ones are retried.
func TestDoGet_NoRetryOnNon40001(t *testing.T) {
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

	c := newClientWithBaseURL(srv.URL, &Config{AppId: "a", AppSecret: "b"})
	if _, err := c.AccessTokenE(context.Background()); err != nil {
		t.Fatal(err)
	}
	var result Resp
	err := c.doGet(context.Background(), "/cgi-bin/test", url.Values{"access_token": {"T"}}, &result)
	if err == nil {
		t.Fatal("expected error")
	}
	var werr *WeixinError
	if !errors.As(err, &werr) || werr.ErrCode != 48001 {
		t.Errorf("expected WeixinError 48001, got %v", err)
	}
	if got := atomic.LoadInt32(&apiCalls); got != 1 {
		t.Errorf("expected exactly 1 API call (no retry), got %d", got)
	}
}

// TestDoGet_NoRetryWhenSecondAlsoFails verifies that if the second attempt
// (after invalidate) ALSO returns 40001, the error is propagated and we do
// not retry indefinitely.
func TestDoGet_NoRetryWhenSecondAlsoFails(t *testing.T) {
	var apiCalls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/cgi-bin/token") {
			_, _ = w.Write([]byte(`{"access_token":"T","expires_in":7200}`))
			return
		}
		atomic.AddInt32(&apiCalls, 1)
		_, _ = w.Write([]byte(`{"errcode":40001,"errmsg":"invalid credential"}`))
	}))
	defer srv.Close()

	c := newClientWithBaseURL(srv.URL, &Config{AppId: "a", AppSecret: "b"})
	if _, err := c.AccessTokenE(context.Background()); err != nil {
		t.Fatal(err)
	}
	var result Resp
	err := c.doGet(context.Background(), "/cgi-bin/test", url.Values{"access_token": {"T"}}, &result)
	if err == nil {
		t.Fatal("expected error")
	}
	var werr *WeixinError
	if !errors.As(err, &werr) || werr.ErrCode != 40001 {
		t.Errorf("expected WeixinError 40001, got %v", err)
	}
	if got := atomic.LoadInt32(&apiCalls); got != 2 {
		t.Errorf("expected exactly 2 API calls (initial + 1 retry), got %d", got)
	}
}

// TestPatchAccessToken covers the helper directly: query-only path, path-embedded
// token (start, middle, end of query), and no-token case (no replacement).
func TestPatchAccessToken(t *testing.T) {
	t.Run("params replaces value", func(t *testing.T) {
		params := url.Values{"access_token": {"OLD"}, "other": {"x"}}
		var p *string
		ok := patchAccessToken(p, &params, "NEW")
		if !ok {
			t.Fatal("expected ok")
		}
		if got := params.Get("access_token"); got != "NEW" {
			t.Errorf("got %q want NEW", got)
		}
		if got := params.Get("other"); got != "x" {
			t.Errorf("other field clobbered: %q", got)
		}
	})
	t.Run("path replaces value at start", func(t *testing.T) {
		path := "/cgi-bin/x?access_token=OLD&other=1"
		ok := patchAccessToken(&path, nil, "NEW")
		if !ok {
			t.Fatal("expected ok")
		}
		u, _ := url.Parse(path)
		if got := u.Query().Get("access_token"); got != "NEW" {
			t.Errorf("got %q want NEW", got)
		}
		if got := u.Query().Get("other"); got != "1" {
			t.Errorf("other clobbered: %q", got)
		}
	})
	t.Run("path replaces value at end", func(t *testing.T) {
		path := "/cgi-bin/x?other=1&access_token=OLD"
		ok := patchAccessToken(&path, nil, "NEW")
		if !ok {
			t.Fatal("expected ok")
		}
		u, _ := url.Parse(path)
		if got := u.Query().Get("access_token"); got != "NEW" {
			t.Errorf("got %q want NEW", got)
		}
	})
	t.Run("no token site returns false", func(t *testing.T) {
		path := "/cgi-bin/x?other=1"
		ok := patchAccessToken(&path, nil, "NEW")
		if ok {
			t.Error("expected false when no access_token to patch")
		}
	})
	t.Run("nil params and nil path returns false", func(t *testing.T) {
		ok := patchAccessToken(nil, nil, "NEW")
		if ok {
			t.Error("expected false")
		}
	})
	t.Run("does not mutate caller's url.Values", func(t *testing.T) {
		// patchAccessToken clones the map so callers' input is untouched.
		original := url.Values{"access_token": {"OLD"}}
		ref := original
		patchAccessToken(nil, &ref, "NEW")
		if got := original.Get("access_token"); got != "OLD" {
			t.Errorf("caller's params mutated: got %q want OLD", got)
		}
		if got := ref.Get("access_token"); got != "NEW" {
			t.Errorf("ref not updated: %q", got)
		}
	})
}

// TestDoGet_RetryWithTokenSource verifies the self-heal path delegates to
// the TokenSource's Invalidator implementation when one is present, rather
// than clearing the Client's internal cache (which is bypassed anyway).
func TestDoGet_RetryWithTokenSource(t *testing.T) {
	var apiCalls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&apiCalls, 1)
		tok := r.URL.Query().Get("access_token")
		if n == 1 {
			if tok != "T1" {
				t.Errorf("first call tok=%q want T1", tok)
			}
			_, _ = w.Write([]byte(`{"errcode":40001,"errmsg":"invalid"}`))
			return
		}
		if tok != "T2" {
			t.Errorf("retry tok=%q want T2", tok)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()

	src := &invalidatableTokenSource{tokensPerSession: []string{"T1", "T2"}}
	c := NewClient(context.Background(), &Config{AppId: "a"},
		WithHTTPClient(utils.NewHTTP(srv.URL, utils.WithTimeout(time.Second*3))),
		WithTokenSource(src),
	)

	var result Resp
	tok, _ := c.AccessTokenE(context.Background())
	err := c.doGet(context.Background(), "/cgi-bin/test", url.Values{"access_token": {tok}}, &result)
	if err != nil {
		t.Fatalf("doGet: %v", err)
	}
	if got := atomic.LoadInt32(&src.invalidateCalls); got != 1 {
		t.Errorf("expected 1 Invalidate on tokensource, got %d", got)
	}
	if got := atomic.LoadInt32(&apiCalls); got != 2 {
		t.Errorf("expected 2 API calls, got %d", got)
	}
}
