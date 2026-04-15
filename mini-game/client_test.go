package mini_game

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

// newTestClient 创建一个指向 httptest.Server 的测试客户端。
func newTestClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	c, err := NewClient(Config{AppId: "wx_test", AppSecret: "secret_test"},
		WithHTTP(utils.NewHTTP(srv.URL, utils.WithTimeout(time.Second*3))))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestNewClient_Validation(t *testing.T) {
	// AppId 必填
	if _, err := NewClient(Config{}); err == nil {
		t.Error("expected error for empty AppId")
	}
	// AppSecret 必填（无 TokenSource）
	if _, err := NewClient(Config{AppId: "wx"}); err == nil {
		t.Error("expected error for empty AppSecret without TokenSource")
	}
	// 有 TokenSource 时 AppSecret 可为空
	fake := &fakeTokenSource{token: "T"}
	c, err := NewClient(Config{AppId: "wx"}, WithTokenSource(fake))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("client should not be nil")
	}
}

func TestAccessToken_FetchAndCache(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/token") {
			t.Errorf("unexpected path: %s", r.URL.Path)
			return
		}
		if r.URL.Query().Get("grant_type") != "client_credential" {
			t.Errorf("unexpected grant_type: %s", r.URL.Query().Get("grant_type"))
		}
		calls++
		_, _ = w.Write([]byte(`{"access_token":"TOK","expires_in":7200}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)

	// 多次调用只应触发一次网络请求（缓存生效）
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

func TestAccessToken_Errcode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":40013,"errmsg":"invalid appid"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.AccessToken(context.Background())
	if err == nil {
		t.Fatal("expected error for errcode response")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.ErrCode != 40013 {
		t.Errorf("ErrCode = %d, want 40013", apiErr.ErrCode)
	}
}

// TestAccessToken_TTLClamp verifies that a tiny/zero expires_in is floored so
// the cache doesn't collapse into a refresh storm.
func TestAccessToken_TTLClamp(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		_, _ = w.Write([]byte(`{"access_token":"TOK","expires_in":0}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if _, err := c.AccessToken(context.Background()); err != nil {
		t.Fatal(err)
	}
	// Second call within ~a second should hit the cache, not the server,
	// because the floor keeps expiresAt well in the future.
	if _, err := c.AccessToken(context.Background()); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Errorf("expected 1 fetch with TTL clamp, got %d", calls)
	}
}

func TestCode2Session(t *testing.T) {
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

	c := newTestClient(t, srv)
	resp, err := c.Code2Session(context.Background(), "CODE")
	if err != nil {
		t.Fatal(err)
	}
	if resp.OpenId != "o1" || resp.SessionKey != "sk" || resp.UnionId != "u1" {
		t.Errorf("unexpected resp: %+v", resp)
	}
}

func TestCode2Session_EmptyCode(t *testing.T) {
	c, _ := NewClient(Config{AppId: "wx", AppSecret: "sec"})
	if _, err := c.Code2Session(context.Background(), ""); err == nil {
		t.Error("expected error for empty jsCode")
	}
}

func TestCode2Session_Errcode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":40029,"errmsg":"invalid code"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if _, err := c.Code2Session(context.Background(), "BAD"); err == nil {
		t.Error("expected error")
	}
}

// --- helpers ---

type fakeTokenSource struct {
	token string
	err   error
	calls int
}

func (f *fakeTokenSource) AccessToken(_ context.Context) (string, error) {
	f.calls++
	return f.token, f.err
}
