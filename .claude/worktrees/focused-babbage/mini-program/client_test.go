package mini_program

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/godrealms/go-wechat-sdk/core"
)

func TestNewClient_UsesPostForToken(t *testing.T) {
	var gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "mini-token",
			"expires_in":   7200,
		})
	}))
	defer srv.Close()

	cfg := &Config{BaseConfig: core.BaseConfig{AppId: "app1", AppSecret: "sec1"}}
	base := core.NewBaseClient(context.Background(), &cfg.BaseConfig, srv.URL, "/stable_token", "POST")
	c := &Client{BaseClient: base}

	token := c.GetAccessToken()
	if token != "mini-token" {
		t.Errorf("expected mini-token, got %s", token)
	}
	if gotMethod != "POST" {
		t.Errorf("expected POST token method, got %s", gotMethod)
	}
}

func TestCode2Session_UsesAppidSecret(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/sns/jscode2session" {
			if r.URL.Query().Get("appid") == "" {
				http.Error(w, "missing appid", 400)
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{
				"openid":      "test-openid",
				"session_key": "test-session-key",
				"errcode":     0,
				"errmsg":      "ok",
			})
		} else {
			// token endpoint
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "tok", "expires_in": 7200,
			})
		}
	}))
	defer srv.Close()

	cfg := &Config{BaseConfig: core.BaseConfig{AppId: "app1", AppSecret: "sec1"}}
	base := core.NewBaseClient(context.Background(), &cfg.BaseConfig, srv.URL, "/stable_token", "POST")
	c := &Client{BaseClient: base}

	result, err := c.Code2Session("test-code")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.OpenId != "test-openid" {
		t.Errorf("expected test-openid, got %s", result.OpenId)
	}
}
