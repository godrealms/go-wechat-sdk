package offiaccount

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/godrealms/go-wechat-sdk/core"
)

func TestNewClient_SetsFields(t *testing.T) {
	cfg := &Config{
		BaseConfig:     core.BaseConfig{AppId: "app1", AppSecret: "sec1"},
		Token:          "tok",
		EncodingAESKey: "key",
	}
	c := NewClient(context.Background(), cfg)
	if c.Token != "tok" {
		t.Errorf("expected tok, got %s", c.Token)
	}
	if c.EncodingAESKey != "key" {
		t.Errorf("expected key, got %s", c.EncodingAESKey)
	}
	if c.Config.AppId != "app1" {
		t.Errorf("expected app1, got %s", c.Config.AppId)
	}
}

func TestClient_GetAccessToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "test-token-123",
			"expires_in":   7200,
		})
	}))
	defer srv.Close()

	cfg := &Config{
		BaseConfig: core.BaseConfig{AppId: "app1", AppSecret: "sec1"},
	}
	base := core.NewBaseClient(context.Background(), &cfg.BaseConfig, srv.URL, "/token", "GET")
	c := &Client{BaseClient: base}

	token := c.GetAccessToken()
	if token != "test-token-123" {
		t.Errorf("expected test-token-123, got %s", token)
	}
}
