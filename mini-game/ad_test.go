package mini_game

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetGameAdData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/token":
			_, _ = w.Write([]byte(`{"access_token":"TEST_TOKEN","expires_in":7200}`))
		case "/wxa/game/getgameaddata":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Query().Get("access_token") != "TEST_TOKEN" {
				t.Errorf("missing access_token")
			}
			var req GetGameAdDataReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatal(err)
			}
			if req.StartDate != "20240101" {
				t.Errorf("expected start_date 20240101, got %s", req.StartDate)
			}
			if req.EndDate != "20240107" {
				t.Errorf("expected end_date 20240107, got %s", req.EndDate)
			}
			_, _ = w.Write([]byte(`{"items":[{"date":"20240101","ad_unit_id":"ad_001","req_count":1000,"show_count":800,"click_count":50,"income":5000}]}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	resp, err := c.GetGameAdData(context.Background(), &GetGameAdDataReq{
		StartDate: "20240101",
		EndDate:   "20240107",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(resp.Items))
	}
	item := resp.Items[0]
	if item.Date != "20240101" {
		t.Errorf("unexpected date: %s", item.Date)
	}
	if item.AdUnitID != "ad_001" {
		t.Errorf("unexpected ad_unit_id: %s", item.AdUnitID)
	}
	if item.ReqCount != 1000 {
		t.Errorf("unexpected req_count: %d", item.ReqCount)
	}
	if item.ShowCount != 800 {
		t.Errorf("unexpected show_count: %d", item.ShowCount)
	}
	if item.ClickCount != 50 {
		t.Errorf("unexpected click_count: %d", item.ClickCount)
	}
	if item.Income != 5000 {
		t.Errorf("unexpected income: %d", item.Income)
	}
}
