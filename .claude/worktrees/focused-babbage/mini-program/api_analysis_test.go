package mini_program

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/godrealms/go-wechat-sdk/core"
)

func TestGetDailyVisitTrend(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/datacube/getweanalysisappiddailyvisittrend" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"errcode": 0,
				"errmsg":  "ok",
				"list": []map[string]interface{}{
					{"ref_date": "20240101", "session_cnt": 100, "visit_pv": 200, "visit_uv": 50},
				},
			})
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{"access_token": "tok", "expires_in": 7200})
		}
	}))
	defer srv.Close()

	cfg := &Config{BaseConfig: core.BaseConfig{AppId: "app1", AppSecret: "sec1"}}
	base := core.NewBaseClient(context.Background(), &cfg.BaseConfig, srv.URL, "/token", "POST")
	c := &Client{BaseClient: base}

	result, err := c.GetDailyVisitTrend(&AnalysisDateRequest{
		BeginDate: "20240101",
		EndDate:   "20240101",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.List) != 1 {
		t.Errorf("expected 1 item, got %d", len(result.List))
	}
	if result.List[0].RefDate != "20240101" {
		t.Errorf("expected 20240101, got %s", result.List[0].RefDate)
	}
}

func TestGetAllDelivery(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/express/business/delivery/getall" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"errcode": 0,
				"errmsg":  "ok",
				"count":   2,
				"data": []map[string]interface{}{
					{"delivery_id": "SF", "delivery_name": "顺丰速递"},
					{"delivery_id": "ZTO", "delivery_name": "中通快递"},
				},
			})
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{"access_token": "tok", "expires_in": 7200})
		}
	}))
	defer srv.Close()

	cfg := &Config{BaseConfig: core.BaseConfig{AppId: "app1", AppSecret: "sec1"}}
	base := core.NewBaseClient(context.Background(), &cfg.BaseConfig, srv.URL, "/token", "POST")
	c := &Client{BaseClient: base}

	result, err := c.GetAllDelivery()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Count != 2 {
		t.Errorf("expected 2 deliveries, got %d", result.Count)
	}
	if result.Data[0].DeliveryId != "SF" {
		t.Errorf("expected SF, got %s", result.Data[0].DeliveryId)
	}
}
