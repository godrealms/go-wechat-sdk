package oplatform

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

func testConfig() Config {
	return Config{
		ComponentAppID:     "wxcomp",
		ComponentAppSecret: "secret",
		Token:              "tk",
		EncodingAESKey:     "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQ", // 43 chars
	}
}

func newTestClient(t *testing.T, baseURL string, opts ...Option) *Client {
	t.Helper()
	opts = append(opts, WithHTTP(utils.NewHTTP(baseURL, utils.WithTimeout(time.Second*3))))
	c, err := NewClient(testConfig(), opts...)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestClient_ComponentAccessToken_LazyAndCaches(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/component/api_component_token") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		atomic.AddInt32(&calls, 1)
		_, _ = w.Write([]byte(`{"component_access_token":"CTOK","expires_in":7200}`))
	}))
	defer srv.Close()

	store := NewMemoryStore()
	_ = store.SetVerifyTicket(context.Background(), "TICKET")
	c := newTestClient(t, srv.URL, WithStore(store))

	for i := 0; i < 3; i++ {
		tok, err := c.ComponentAccessToken(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if tok != "CTOK" {
			t.Errorf("got %q", tok)
		}
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("expected 1 fetch, got %d", got)
	}
}

func TestClient_ComponentAccessToken_MissingTicket(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("should not reach the server when ticket is missing: %s", r.URL.Path)
	}))
	defer srv.Close()
	c := newTestClient(t, srv.URL)

	_, err := c.ComponentAccessToken(context.Background())
	if !errors.Is(err, ErrVerifyTicketMissing) {
		t.Errorf("expected ErrVerifyTicketMissing, got %v", err)
	}
}

func TestClient_ComponentAccessToken_WeixinError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":40013,"errmsg":"invalid appid"}`))
	}))
	defer srv.Close()
	store := NewMemoryStore()
	_ = store.SetVerifyTicket(context.Background(), "TICKET")
	c := newTestClient(t, srv.URL, WithStore(store))

	_, err := c.ComponentAccessToken(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	var werr *WeixinError
	if !errors.As(err, &werr) || werr.ErrCode != 40013 {
		t.Errorf("expected WeixinError 40013, got %v", err)
	}
}

func TestClient_RefreshComponentToken_ForcesFetch(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		_, _ = w.Write([]byte(`{"component_access_token":"CTOK","expires_in":7200}`))
	}))
	defer srv.Close()
	store := NewMemoryStore()
	_ = store.SetVerifyTicket(context.Background(), "TICKET")
	c := newTestClient(t, srv.URL, WithStore(store))

	if _, err := c.ComponentAccessToken(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := c.RefreshComponentToken(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Errorf("expected 2 fetches after forced refresh, got %d", got)
	}
}

func TestClient_PreAuthCode(t *testing.T) {
	var gotBody string
	mux := http.NewServeMux()
	mux.HandleFunc("/cgi-bin/component/api_component_token", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"component_access_token":"CTOK","expires_in":7200}`))
	})
	mux.HandleFunc("/cgi-bin/component/api_create_preauthcode", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("component_access_token") != "CTOK" {
			t.Errorf("missing component_access_token in query")
		}
		bb, _ := io.ReadAll(r.Body)
		gotBody = string(bb)
		_, _ = w.Write([]byte(`{"pre_auth_code":"PREAUTH","expires_in":600}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	store := NewMemoryStore()
	_ = store.SetVerifyTicket(context.Background(), "TICKET")
	c := newTestClient(t, srv.URL, WithStore(store))

	code, err := c.PreAuthCode(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if code != "PREAUTH" {
		t.Errorf("got %q", code)
	}
	if !strings.Contains(gotBody, `"component_appid":"wxcomp"`) {
		t.Errorf("unexpected body: %s", gotBody)
	}
}

func TestClient_AuthorizeURL(t *testing.T) {
	c, err := NewClient(testConfig())
	if err != nil {
		t.Fatal(err)
	}
	u := c.AuthorizeURL("PC", "https://example.com/cb", 3, "")
	if !strings.Contains(u, "component_appid=wxcomp") {
		t.Errorf("missing component_appid: %s", u)
	}
	if !strings.Contains(u, "pre_auth_code=PC") {
		t.Errorf("missing pre_auth_code: %s", u)
	}
	if !strings.Contains(u, "redirect_uri=https%3A%2F%2Fexample.com%2Fcb") {
		t.Errorf("redirect_uri not encoded: %s", u)
	}
	if !strings.Contains(u, "auth_type=3") {
		t.Errorf("missing auth_type: %s", u)
	}
	if strings.Contains(u, "biz_appid") {
		t.Errorf("biz_appid should be absent when empty: %s", u)
	}
}

func TestClient_AuthorizeURL_WithBizAppid(t *testing.T) {
	c, _ := NewClient(testConfig())
	u := c.AuthorizeURL("PC", "https://x/cb", 1, "wxbiz")
	if !strings.Contains(u, "biz_appid=wxbiz") {
		t.Errorf("biz_appid missing: %s", u)
	}
}

func TestClient_MobileAuthorizeURL(t *testing.T) {
	c, _ := NewClient(testConfig())
	u := c.MobileAuthorizeURL("PC", "https://x/cb", 3, "")
	if !strings.Contains(u, "action=bindcomponent") {
		t.Errorf("missing action=bindcomponent: %s", u)
	}
	if !strings.Contains(u, "pre_auth_code=PC") {
		t.Errorf("missing pre_auth_code: %s", u)
	}
}

func TestClient_QueryAuth_PopulatesStore(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/cgi-bin/component/api_component_token", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"component_access_token":"CTOK","expires_in":7200}`))
	})
	mux.HandleFunc("/cgi-bin/component/api_query_auth", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
  "authorization_info": {
    "authorizer_appid": "wxAuthed",
    "authorizer_access_token": "ATOK",
    "expires_in": 7200,
    "authorizer_refresh_token": "RTOK",
    "func_info": [{"funcscope_category": {"id": 1}}]
  }
}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	store := NewMemoryStore()
	_ = store.SetVerifyTicket(context.Background(), "TICKET")
	c := newTestClient(t, srv.URL, WithStore(store))

	info, err := c.QueryAuth(context.Background(), "AUTHCODE")
	if err != nil {
		t.Fatal(err)
	}
	if info.AuthorizerAppID != "wxAuthed" {
		t.Errorf("appid mismatch: %+v", info)
	}
	got, err := store.GetAuthorizer(context.Background(), "wxAuthed")
	if err != nil {
		t.Fatal(err)
	}
	if got.AccessToken != "ATOK" || got.RefreshToken != "RTOK" {
		t.Errorf("store mismatch: %+v", got)
	}
	if !got.ExpireAt.After(time.Now()) {
		t.Errorf("expire_at should be future, got %v", got.ExpireAt)
	}
}

func TestClient_GetAuthorizerInfo(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/cgi-bin/component/api_component_token", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"component_access_token":"CTOK","expires_in":7200}`))
	})
	mux.HandleFunc("/cgi-bin/component/api_get_authorizer_info", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
  "authorizer_info": {"nick_name":"biz","user_name":"gh_x","principal_name":"Acme"},
  "authorization_info": {"authorizer_appid":"wxAuthed","authorizer_refresh_token":"RTOK"}
}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	store := NewMemoryStore()
	_ = store.SetVerifyTicket(context.Background(), "TICKET")
	c := newTestClient(t, srv.URL, WithStore(store))

	got, err := c.GetAuthorizerInfo(context.Background(), "wxAuthed")
	if err != nil {
		t.Fatal(err)
	}
	if got.AuthorizerInfo.NickName != "biz" || got.AuthorizerInfo.PrincipalName != "Acme" {
		t.Errorf("unexpected: %+v", got)
	}
}

func TestClient_GetAuthorizerList(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/cgi-bin/component/api_component_token", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"component_access_token":"CTOK","expires_in":7200}`))
	})
	mux.HandleFunc("/cgi-bin/component/api_get_authorizer_list", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
  "total_count": 2,
  "list": [
    {"authorizer_appid":"wxA","refresh_token":"rA","auth_time":1700000000},
    {"authorizer_appid":"wxB","refresh_token":"rB","auth_time":1700000001}
  ]
}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	store := NewMemoryStore()
	_ = store.SetVerifyTicket(context.Background(), "TICKET")
	c := newTestClient(t, srv.URL, WithStore(store))

	list, err := c.GetAuthorizerList(context.Background(), 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if list.TotalCount != 2 || len(list.List) != 2 {
		t.Errorf("unexpected: %+v", list)
	}
}

func TestClient_GetSetAuthorizerOption(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/cgi-bin/component/api_component_token", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"component_access_token":"CTOK","expires_in":7200}`))
	})
	mux.HandleFunc("/cgi-bin/component/api_get_authorizer_option", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"authorizer_appid":"wxA","option_name":"voice_recognize","option_value":"1"}`))
	})
	mux.HandleFunc("/cgi-bin/component/api_set_authorizer_option", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	store := NewMemoryStore()
	_ = store.SetVerifyTicket(context.Background(), "TICKET")
	c := newTestClient(t, srv.URL, WithStore(store))

	opt, err := c.GetAuthorizerOption(context.Background(), "wxA", "voice_recognize")
	if err != nil {
		t.Fatal(err)
	}
	if opt.OptionValue != "1" {
		t.Errorf("unexpected: %+v", opt)
	}
	if err := c.SetAuthorizerOption(context.Background(), "wxA", "voice_recognize", "0"); err != nil {
		t.Fatal(err)
	}
}
