package offiaccount_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/godrealms/go-wechat-sdk/offiaccount"
	"github.com/godrealms/go-wechat-sdk/utils"
)

func newTestClient(srv *httptest.Server) *offiaccount.Client {
	cfg := &offiaccount.Config{AppId: "appid", AppSecret: "secret"}
	c := offiaccount.NewClient(context.Background(), cfg,
		offiaccount.WithHTTPClient(utils.NewHTTP(srv.URL, utils.WithTimeout(3*time.Second))),
		offiaccount.WithTokenSource(&staticToken{token: "FAKE_TOKEN"}),
	)
	return c
}

type staticToken struct{ token string }

func (s *staticToken) AccessToken(_ context.Context) (string, error) { return s.token, nil }

func TestGetCallbackIp_ReturnsWeixinError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errcode":40001,"errmsg":"invalid credential","ip_list":null}`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.GetCallbackIp()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var werr *offiaccount.WeixinError
	if !errors.As(err, &werr) {
		t.Fatalf("expected *WeixinError, got %T: %v", err, err)
	}
	if werr.Code() != 40001 {
		t.Errorf("Code() = %d, want 40001", werr.Code())
	}

	var apiErr utils.WechatAPIError
	if !errors.As(err, &apiErr) {
		t.Fatal("expected utils.WechatAPIError")
	}
}

func TestGetApiDomainIP_ReturnsWeixinError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errcode":48001,"errmsg":"api unauthorized","ip_list":null}`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.GetApiDomainIP()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var werr *offiaccount.WeixinError
	if !errors.As(err, &werr) {
		t.Fatalf("expected *WeixinError, got %T: %v", err, err)
	}
	if werr.Code() != 48001 {
		t.Errorf("Code() = %d, want 48001", werr.Code())
	}
}
