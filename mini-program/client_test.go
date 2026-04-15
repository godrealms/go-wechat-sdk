package mini_program

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

func TestClient_Code2Session(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/sns/jscode2session") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("js_code") != "CODE" {
			t.Errorf("missing js_code")
		}
		_, _ = w.Write([]byte(`{"openid":"o1","session_key":"sk","unionid":"u1"}`))
	}))
	defer srv.Close()

	c, err := NewClient(Config{AppId: "wx", AppSecret: "sec"},
		WithHTTP(utils.NewHTTP(srv.URL, utils.WithTimeout(time.Second*3))))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := c.Code2Session(context.Background(), "CODE")
	if err != nil {
		t.Fatal(err)
	}
	if resp.OpenId != "o1" || resp.SessionKey != "sk" {
		t.Errorf("unexpected resp: %+v", resp)
	}
}

func TestClient_AccessTokenCaches(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		_, _ = w.Write([]byte(`{"access_token":"TOK","expires_in":7200}`))
	}))
	defer srv.Close()
	c, _ := NewClient(Config{AppId: "wx", AppSecret: "sec"},
		WithHTTP(utils.NewHTTP(srv.URL, utils.WithTimeout(time.Second*3))))
	for i := 0; i < 3; i++ {
		tok, err := c.AccessToken(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if tok != "TOK" {
			t.Errorf("got %q", tok)
		}
	}
	if calls != 1 {
		t.Errorf("expected 1 fetch, got %d", calls)
	}
}

func TestClient_Code2Session_Errcode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":40029,"errmsg":"invalid code"}`))
	}))
	defer srv.Close()
	c, _ := NewClient(Config{AppId: "wx", AppSecret: "sec"},
		WithHTTP(utils.NewHTTP(srv.URL, utils.WithTimeout(time.Second*3))))
	if _, err := c.Code2Session(context.Background(), "BAD"); err == nil {
		t.Error("expected error")
	}
}

type fakeTokenSource struct {
	token string
	err   error
	calls int
}

func (f *fakeTokenSource) AccessToken(ctx context.Context) (string, error) {
	f.calls++
	return f.token, f.err
}

func TestAccessToken_HandlesShortExpiresIn(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"access_token":"X","expires_in":10}`))
	}))
	defer srv.Close()
	c, err := NewClient(Config{AppId: "wx", AppSecret: "sec"},
		WithHTTP(utils.NewHTTP(srv.URL, utils.WithTimeout(time.Second*3))))
	if err != nil {
		t.Fatal(err)
	}
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

func TestClient_AccessToken_UsesInjectedTokenSource(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("/cgi-bin/token must NOT be called when TokenSource is injected: %s", r.URL.Path)
	}))
	defer srv.Close()

	fake := &fakeTokenSource{token: "INJECTED_MP"}
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
	if tok != "INJECTED_MP" {
		t.Errorf("got %q, want INJECTED_MP", tok)
	}
	if fake.calls != 1 {
		t.Errorf("expected 1 call, got %d", fake.calls)
	}
}
