package mini_game

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetDailySummary_ErrcodeError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/token":
			_, _ = w.Write([]byte(`{"access_token":"TEST_TOKEN","expires_in":7200}`))
		default:
			_, _ = w.Write([]byte(`{"errcode":40001,"errmsg":"invalid token"}`))
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetDailySummary(context.Background(), &AnalysisDateReq{BeginDate: "20240101", EndDate: "20240101"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.Code() != 40001 {
		t.Errorf("expected Code() == 40001, got %d", apiErr.Code())
	}
}

func TestGetDailySummary(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/token":
			_, _ = w.Write([]byte(`{"access_token":"TEST_TOKEN","expires_in":7200}`))
		case "/datacube/getweanalysisappiddailysummarytrend":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Query().Get("access_token") != "TEST_TOKEN" {
				t.Errorf("missing access_token")
			}
			var req AnalysisDateReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatal(err)
			}
			if req.BeginDate != "20240101" {
				t.Errorf("expected begin_date 20240101, got %s", req.BeginDate)
			}
			if req.EndDate != "20240101" {
				t.Errorf("expected end_date 20240101, got %s", req.EndDate)
			}
			_, _ = w.Write([]byte(`{"list":[{"ref_date":"20240101","visit_total":100,"share_pv":20,"share_uv":10}]}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	resp, err := c.GetDailySummary(context.Background(), &AnalysisDateReq{
		BeginDate: "20240101",
		EndDate:   "20240101",
	})
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

func TestGetDailyRetain(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/token":
			_, _ = w.Write([]byte(`{"access_token":"TEST_TOKEN","expires_in":7200}`))
		case "/datacube/getweanalysisappiddailyretaininfo":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Query().Get("access_token") != "TEST_TOKEN" {
				t.Errorf("missing access_token")
			}
			var req AnalysisDateReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatal(err)
			}
			if req.BeginDate != "20240101" {
				t.Errorf("expected begin_date 20240101, got %s", req.BeginDate)
			}
			_, _ = w.Write([]byte(`{"ref_date":"20240101","visit_uv_new":[{"date_key":"1","value":50}],"visit_uv":[{"date_key":"1","value":200}]}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	resp, err := c.GetDailyRetain(context.Background(), &AnalysisDateReq{
		BeginDate: "20240101",
		EndDate:   "20240101",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.RefDate != "20240101" {
		t.Errorf("unexpected ref_date: %q", resp.RefDate)
	}
	if len(resp.VisitUVNew) != 1 {
		t.Fatalf("expected 1 visit_uv_new item, got %d", len(resp.VisitUVNew))
	}
	if resp.VisitUVNew[0].Value != 50 {
		t.Errorf("unexpected visit_uv_new value: %d", resp.VisitUVNew[0].Value)
	}
	if len(resp.VisitUV) != 1 {
		t.Fatalf("expected 1 visit_uv item, got %d", len(resp.VisitUV))
	}
	if resp.VisitUV[0].Value != 200 {
		t.Errorf("unexpected visit_uv value: %d", resp.VisitUV[0].Value)
	}
}
