package offiaccount

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

func newClientWithBaseURL(baseURL string, cfg *Config) *Client {
	c := NewClient(context.Background(), cfg)
	c.Https = utils.NewHTTP(baseURL, utils.WithTimeout(time.Second*5))
	return c
}

func TestClient_AccessTokenE_CachesAndRefreshes(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/token") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"TOKEN_X","expires_in":7200}`))
	}))
	defer srv.Close()

	c := newClientWithBaseURL(srv.URL, &Config{AppId: "appid", AppSecret: "secret"})
	tok, err := c.AccessTokenE(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if tok != "TOKEN_X" {
		t.Errorf("unexpected token: %s", tok)
	}
	// 第二次应使用缓存
	if _, err := c.AccessTokenE(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("expected 1 refresh, got %d", got)
	}
}

func TestClient_AccessTokenE_ReturnsWeixinError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errcode":40013,"errmsg":"invalid appid"}`))
	}))
	defer srv.Close()

	c := newClientWithBaseURL(srv.URL, &Config{AppId: "bad", AppSecret: "secret"})
	_, err := c.AccessTokenE(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	var werr *WeixinError
	if !errors.As(err, &werr) {
		t.Fatalf("expected *WeixinError, got %T: %v", err, err)
	}
	if werr.ErrCode != 40013 {
		t.Errorf("unexpected errcode: %d", werr.ErrCode)
	}
}

func TestClient_GetAccessToken_BackwardsCompatible(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"access_token":"COMPAT","expires_in":7200}`))
	}))
	defer srv.Close()

	c := newClientWithBaseURL(srv.URL, &Config{AppId: "appid", AppSecret: "secret"})
	if got := c.GetAccessToken(); got != "COMPAT" {
		t.Errorf("expected COMPAT, got %q", got)
	}
}

func TestCheckResp(t *testing.T) {
	if err := CheckResp(&Resp{ErrCode: 0}); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	err := CheckResp(&Resp{ErrCode: 40001, ErrMsg: "invalid credential"})
	if err == nil {
		t.Fatal("expected error")
	}
	var werr *WeixinError
	if !errors.As(err, &werr) || werr.ErrCode != 40001 {
		t.Errorf("expected WeixinError 40001, got %v", err)
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

func TestClient_AccessTokenE_UsesInjectedTokenSource(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("/cgi-bin/token must NOT be called when TokenSource is injected: %s", r.URL.Path)
	}))
	defer srv.Close()

	fake := &fakeTokenSource{token: "INJECTED"}
	c := NewClient(context.Background(), &Config{AppId: "appid"},
		WithHTTPClient(utils.NewHTTP(srv.URL, utils.WithTimeout(time.Second*3))),
		WithTokenSource(fake),
	)
	tok, err := c.AccessTokenE(context.Background())
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
