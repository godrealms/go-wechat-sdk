package mini_store

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

func newTestClient(t *testing.T, baseURL string) *Client {
	t.Helper()
	c, err := NewClient(Config{AppId: "wx", AppSecret: "sec"},
		WithHTTP(utils.NewHTTP(baseURL, utils.WithTimeout(3*time.Second))))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

type fakeTokenSource struct {
	token string
	err   error
	calls int
}

func (f *fakeTokenSource) AccessToken(_ context.Context) (string, error) {
	f.calls++
	return f.token, f.err
}

func TestNewClient(t *testing.T) {
	if _, err := NewClient(Config{}); err == nil {
		t.Error("expected error for empty AppId")
	}
	if _, err := NewClient(Config{AppId: "wx"}); err == nil {
		t.Error("expected error for empty AppSecret without TokenSource")
	}
	c, err := NewClient(Config{AppId: "wx", AppSecret: "sec"})
	if err != nil {
		t.Fatal(err)
	}
	if c.HTTP() == nil {
		t.Error("HTTP() must not be nil")
	}
}

func TestAccessToken_Caches(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		_, _ = w.Write([]byte(`{"access_token":"TOK","expires_in":7200}`))
	}))
	defer srv.Close()
	c := newTestClient(t, srv.URL)
	for i := 0; i < 3; i++ {
		tok, err := c.AccessToken(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if tok != "TOK" {
			t.Errorf("got %q, want TOK", tok)
		}
	}
	if calls != 1 {
		t.Errorf("expected 1 fetch, got %d", calls)
	}
}

// TestAccessToken_ErrcodeTyped verifies token-fetch errcode is surfaced as a
// typed *APIError, so callers can errors.As-distinguish it.
func TestAccessToken_ErrcodeTyped(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":40013,"errmsg":"invalid appid"}`))
	}))
	defer srv.Close()
	c := newTestClient(t, srv.URL)
	_, err := c.AccessToken(context.Background())
	if err == nil {
		t.Fatal("expected errcode error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.ErrCode != 40013 {
		t.Errorf("ErrCode = %d, want 40013", apiErr.ErrCode)
	}
}

// TestAccessToken_TTLClamp verifies tiny/zero expires_in is floored.
func TestAccessToken_TTLClamp(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		_, _ = w.Write([]byte(`{"access_token":"TOK","expires_in":0}`))
	}))
	defer srv.Close()
	c := newTestClient(t, srv.URL)
	if _, err := c.AccessToken(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := c.AccessToken(context.Background()); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Errorf("expected 1 fetch with TTL clamp, got %d", calls)
	}
}

func TestAccessToken_UsesInjectedTokenSource(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("must not call /cgi-bin/token when TokenSource injected: %s", r.URL.Path)
	}))
	defer srv.Close()
	fake := &fakeTokenSource{token: "INJECTED"}
	c, err := NewClient(Config{AppId: "wx", AppSecret: "sec"},
		WithHTTP(utils.NewHTTP(srv.URL, utils.WithTimeout(3*time.Second))),
		WithTokenSource(fake))
	if err != nil {
		t.Fatal(err)
	}
	tok, err := c.AccessToken(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if tok != "INJECTED" {
		t.Errorf("got %q, want INJECTED", tok)
	}
	if fake.calls != 1 {
		t.Errorf("expected TokenSource called 1 time, got %d", fake.calls)
	}
}
