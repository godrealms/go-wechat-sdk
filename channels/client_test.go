package channels

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

func newTestClient(t *testing.T, baseURL string) *Client {
	t.Helper()
	c, err := NewClient(Config{AppId: "wx", AppSecret: "sec"},
		WithHTTP(utils.NewHTTP(baseURL, utils.WithTimeout(time.Second*3))))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestNewClient(t *testing.T) {
	// empty AppId
	if _, err := NewClient(Config{}); err == nil {
		t.Error("expected error for empty AppId")
	}
	// empty AppSecret without TokenSource
	if _, err := NewClient(Config{AppId: "wx"}); err == nil {
		t.Error("expected error for empty AppSecret")
	}
	// valid config
	c, err := NewClient(Config{AppId: "wx", AppSecret: "sec"})
	if err != nil {
		t.Fatal(err)
	}
	if c.HTTP() == nil {
		t.Error("HTTP() should not be nil")
	}
}

func TestAccessToken(t *testing.T) {
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

func TestAccessToken_HandlesShortExpiresIn(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"access_token":"X","expires_in":10}`))
	}))
	defer srv.Close()
	c := newTestClient(t, srv.URL)
	tok, err := c.AccessToken(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if tok != "X" {
		t.Fatalf("got %q", tok)
	}
	// expiresAt must be in the future, not the past.
	if !c.expiresAt.After(time.Now()) {
		t.Errorf("expiresAt is in the past: %v", c.expiresAt)
	}
}

func TestAccessToken_UsesInjectedTokenSource(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("/cgi-bin/token must NOT be called when TokenSource is injected: %s", r.URL.Path)
	}))
	defer srv.Close()

	fake := &fakeTokenSource{token: "INJECTED"}
	c, err := NewClient(Config{AppId: "wx", AppSecret: "sec"},
		WithHTTP(utils.NewHTTP(srv.URL, utils.WithTimeout(time.Second*3))),
		WithTokenSource(fake),
	)
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
		t.Errorf("expected 1 call, got %d", fake.calls)
	}
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
