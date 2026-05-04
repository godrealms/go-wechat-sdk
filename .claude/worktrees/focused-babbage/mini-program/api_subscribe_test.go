package mini_program

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/godrealms/go-wechat-sdk/core"
)

func TestSendSubscribeMessage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/message/subscribe/send" {
			json.NewEncoder(w).Encode(map[string]interface{}{"errcode": 0, "errmsg": "ok"})
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{"access_token": "tok", "expires_in": 7200})
		}
	}))
	defer srv.Close()

	cfg := &Config{BaseConfig: core.BaseConfig{AppId: "app1", AppSecret: "sec1"}}
	base := core.NewBaseClient(context.Background(), &cfg.BaseConfig, srv.URL, "/token", "POST")
	c := &Client{BaseClient: base}

	req := &SendSubscribeMessageRequest{
		ToUser:     "openid123",
		TemplateId: "tmpl001",
		Data: map[string]*SubscribeMessageValue{
			"thing1": {Value: "test value"},
		},
	}
	err := c.SendSubscribeMessage(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMsgSecCheck(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/wxa/msg_sec_check" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"errcode":  0,
				"errmsg":   "ok",
				"trace_id": "trace001",
				"result":   map[string]interface{}{"suggest": "pass", "label": 100},
			})
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{"access_token": "tok", "expires_in": 7200})
		}
	}))
	defer srv.Close()

	cfg := &Config{BaseConfig: core.BaseConfig{AppId: "app1", AppSecret: "sec1"}}
	base := core.NewBaseClient(context.Background(), &cfg.BaseConfig, srv.URL, "/token", "POST")
	c := &Client{BaseClient: base}

	req := &MsgSecCheckRequest{
		Content: "hello world",
		Version: 2,
		Scene:   2,
		Openid:  "openid123",
	}
	result, err := c.MsgSecCheck(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TraceId != "trace001" {
		t.Errorf("expected trace001, got %s", result.TraceId)
	}
}
