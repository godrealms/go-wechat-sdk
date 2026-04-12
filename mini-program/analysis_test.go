package mini_program

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

func TestGetDailySummary(t *testing.T) {
	srv := httptest.NewServer(tokenHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/datacube/getweanalysisappiddailysummarytrend" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Query().Get("access_token") != "TOK" {
			t.Errorf("missing or wrong access_token: %q", r.URL.Query().Get("access_token"))
		}
		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("invalid JSON body: %v", err)
		}
		for _, field := range []string{"begin_date", "end_date"} {
			if _, ok := req[field]; !ok {
				t.Errorf("body missing %q field", field)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"list": [
				{"ref_date": "20240101", "visit_total": 100, "share_pv": 20, "share_uv": 10}
			]
		}`))
	}))
	defer srv.Close()

	c, err := NewClient(Config{AppId: "wx", AppSecret: "sec"},
		WithHTTP(utils.NewHTTP(srv.URL, utils.WithTimeout(time.Second*3))))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := c.GetDailySummary(context.Background(), "20240101", "20240101")
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.List) != 1 {
		t.Fatalf("expected 1 item, got %d", len(resp.List))
	}
	item := resp.List[0]
	if item.RefDate != "20240101" {
		t.Errorf("unexpected ref_date: %q", item.RefDate)
	}
	if item.VisitTotal != 100 {
		t.Errorf("unexpected visit_total: %d", item.VisitTotal)
	}
	if item.SharePV != 20 {
		t.Errorf("unexpected share_pv: %d", item.SharePV)
	}
	if item.ShareUV != 10 {
		t.Errorf("unexpected share_uv: %d", item.ShareUV)
	}
}

func TestGetVisitPage(t *testing.T) {
	srv := httptest.NewServer(tokenHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/datacube/getweanalysisappidvisitpage" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Query().Get("access_token") != "TOK" {
			t.Errorf("missing or wrong access_token: %q", r.URL.Query().Get("access_token"))
		}
		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("invalid JSON body: %v", err)
		}
		for _, field := range []string{"begin_date", "end_date"} {
			if _, ok := req[field]; !ok {
				t.Errorf("body missing %q field", field)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"ref_date": "20240101",
			"list": [
				{"page_path": "pages/index/index", "page_visit_pv": 500, "page_visit_uv": 300}
			]
		}`))
	}))
	defer srv.Close()

	c, err := NewClient(Config{AppId: "wx", AppSecret: "sec"},
		WithHTTP(utils.NewHTTP(srv.URL, utils.WithTimeout(time.Second*3))))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := c.GetVisitPage(context.Background(), "20240101", "20240101")
	if err != nil {
		t.Fatal(err)
	}
	if resp.RefDate != "20240101" {
		t.Errorf("unexpected ref_date: %q", resp.RefDate)
	}
	if len(resp.List) != 1 {
		t.Fatalf("expected 1 item, got %d", len(resp.List))
	}
	item := resp.List[0]
	if item.PagePath != "pages/index/index" {
		t.Errorf("unexpected page_path: %q", item.PagePath)
	}
	if item.PageVisitPV != 500 {
		t.Errorf("unexpected page_visit_pv: %d", item.PageVisitPV)
	}
	if item.PageVisitUV != 300 {
		t.Errorf("unexpected page_visit_uv: %d", item.PageVisitUV)
	}
}

func TestGetDailyVisitTrend(t *testing.T) {
	srv := httptest.NewServer(tokenHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/datacube/getweanalysisappiddailyvisittrend" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Query().Get("access_token") != "TOK" {
			t.Errorf("missing or wrong access_token: %q", r.URL.Query().Get("access_token"))
		}
		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("invalid JSON body: %v", err)
		}
		for _, field := range []string{"begin_date", "end_date"} {
			if _, ok := req[field]; !ok {
				t.Errorf("body missing %q field", field)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"list": [
				{"ref_date": "20240101", "session_cnt": 200, "visit_pv": 1000, "visit_uv": 800, "visit_uv_new": 100, "stay_time_uv": 3.5, "stay_time_session": 2.1, "visit_depth": 1.8}
			]
		}`))
	}))
	defer srv.Close()

	c, err := NewClient(Config{AppId: "wx", AppSecret: "sec"},
		WithHTTP(utils.NewHTTP(srv.URL, utils.WithTimeout(time.Second*3))))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := c.GetDailyVisitTrend(context.Background(), "20240101", "20240101")
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.List) != 1 {
		t.Fatalf("expected 1 item, got %d", len(resp.List))
	}
	item := resp.List[0]
	if item.SessionCnt != 200 {
		t.Errorf("unexpected session_cnt: %d", item.SessionCnt)
	}
	if item.VisitPV != 1000 {
		t.Errorf("unexpected visit_pv: %d", item.VisitPV)
	}
	if item.VisitUV != 800 {
		t.Errorf("unexpected visit_uv: %d", item.VisitUV)
	}
}
