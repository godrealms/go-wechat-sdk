package oplatform

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

func newQRTestClient(baseURL string) *QRLoginClient {
	return NewQRLoginClient("wxqr", "qrsecret",
		WithQRLoginHTTP(utils.NewHTTP(baseURL, utils.WithTimeout(time.Second*3))))
}

func TestQRLoginClient_AuthorizeURL(t *testing.T) {
	q := NewQRLoginClient("wxqr", "sec")
	u := q.AuthorizeURL("https://example.com/cb", "snsapi_login", "state1")
	if !strings.Contains(u, "appid=wxqr") {
		t.Errorf("appid: %s", u)
	}
	if !strings.Contains(u, "scope=snsapi_login") {
		t.Errorf("scope: %s", u)
	}
	if !strings.Contains(u, "state=state1") {
		t.Errorf("state: %s", u)
	}
	if !strings.Contains(u, "redirect_uri=https%3A%2F%2Fexample.com%2Fcb") {
		t.Errorf("redirect_uri: %s", u)
	}
	if !strings.HasSuffix(u, "#wechat_redirect") {
		t.Errorf("missing fragment: %s", u)
	}
}

func TestQRLoginClient_Code2Token(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/sns/oauth2/access_token") {
			t.Errorf("path: %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("appid") != "wxqr" || q.Get("secret") != "qrsecret" || q.Get("code") != "CODE" || q.Get("grant_type") != "authorization_code" {
			t.Errorf("unexpected query: %v", q)
		}
		_, _ = w.Write([]byte(`{"access_token":"A","expires_in":7200,"refresh_token":"R","openid":"O","scope":"snsapi_login","unionid":"U"}`))
	}))
	defer srv.Close()
	q := newQRTestClient(srv.URL)

	tok, err := q.Code2Token(context.Background(), "CODE")
	if err != nil {
		t.Fatal(err)
	}
	if tok.AccessToken != "A" || tok.OpenID != "O" || tok.UnionID != "U" {
		t.Errorf("unexpected: %+v", tok)
	}
}

func TestQRLoginClient_RefreshToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/sns/oauth2/refresh_token") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"access_token":"A2","expires_in":7200,"refresh_token":"R2","openid":"O","scope":"snsapi_login"}`))
	}))
	defer srv.Close()
	q := newQRTestClient(srv.URL)

	tok, err := q.RefreshToken(context.Background(), "RX")
	if err != nil {
		t.Fatal(err)
	}
	if tok.AccessToken != "A2" || tok.RefreshToken != "R2" {
		t.Errorf("unexpected: %+v", tok)
	}
}

func TestQRLoginClient_UserInfo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/sns/userinfo") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"openid":"O","nickname":"N","sex":1,"country":"CN","unionid":"U"}`))
	}))
	defer srv.Close()
	q := newQRTestClient(srv.URL)

	info, err := q.UserInfo(context.Background(), "TOK", "O")
	if err != nil {
		t.Fatal(err)
	}
	if info.Nickname != "N" || info.Country != "CN" {
		t.Errorf("unexpected: %+v", info)
	}
}

func TestQRLoginClient_Auth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/sns/auth") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()
	q := newQRTestClient(srv.URL)

	if err := q.Auth(context.Background(), "TOK", "O"); err != nil {
		t.Fatal(err)
	}
}

func TestQRLoginClient_Code2Token_Errcode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":40029,"errmsg":"invalid code"}`))
	}))
	defer srv.Close()
	q := newQRTestClient(srv.URL)
	if _, err := q.Code2Token(context.Background(), "BAD"); err == nil {
		t.Error("expected error")
	}
}
