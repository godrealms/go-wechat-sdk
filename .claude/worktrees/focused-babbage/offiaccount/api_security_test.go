package offiaccount

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/godrealms/go-wechat-sdk/core"
)

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

	cfg := &Config{
		BaseConfig: core.BaseConfig{AppId: "app1", AppSecret: "sec1"},
	}
	base := core.NewBaseClient(context.Background(), &cfg.BaseConfig, srv.URL, "/token", "GET")
	c := &Client{BaseClient: base}

	result, err := c.MsgSecCheck(&MsgSecCheckRequest{
		Content: "test content",
		Version: 2,
		Scene:   2,
		Openid:  "openid123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TraceId != "trace001" {
		t.Errorf("expected trace001, got %s", result.TraceId)
	}
	if result.Result.Suggest != "pass" {
		t.Errorf("expected pass, got %s", result.Result.Suggest)
	}
}

func TestCreateCard(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/card/create" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"errcode": 0,
				"errmsg":  "ok",
				"card_id": "pFS7Fjg8kV1IdDz01r4jqycQZtVk",
			})
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{"access_token": "tok", "expires_in": 7200})
		}
	}))
	defer srv.Close()

	cfg := &Config{
		BaseConfig: core.BaseConfig{AppId: "app1", AppSecret: "sec1"},
	}
	base := core.NewBaseClient(context.Background(), &cfg.BaseConfig, srv.URL, "/token", "GET")
	c := &Client{BaseClient: base}

	result, err := c.CardCreate(&CardCreateRequest{
		Card: &CardSpec{
			CardType: "DISCOUNT",
			Discount: &DiscountCard{
				BaseInfo: &CardBaseInfo{
					LogoUrl:     "http://example.com/logo.png",
					BrandName:   "TestBrand",
					CodeType:    "CODE_TYPE_TEXT",
					Title:       "测试折扣券",
					Color:       "Color010",
					Notice:      "出示此码享受折扣",
					Description: "测试描述",
					TimeInfo:    &TimeInfo{Type: "DATE_TYPE_FIX_TIME_RANGE", BeginTimestamp: 1700000000, EndTimestamp: 1800000000},
					Sku:         &Sku{Quantity: 100},
				},
				Discount: 70,
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.CardId != "pFS7Fjg8kV1IdDz01r4jqycQZtVk" {
		t.Errorf("expected card id, got %s", result.CardId)
	}
}
