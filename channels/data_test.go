package channels

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetFinderLiveDataList_ErrcodeError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/token":
			_, _ = w.Write([]byte(`{"access_token":"TEST_TOKEN","expires_in":7200}`))
		default:
			_, _ = w.Write([]byte(`{"errcode":40001,"errmsg":"invalid token"}`))
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, err := c.GetFinderLiveDataList(context.Background(), &GetFinderLiveDataListReq{
		StartDate: "2024-01-01",
		EndDate:   "2024-01-31",
	})
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

func TestGetFinderLiveDataList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/token":
			_, _ = w.Write([]byte(`{"access_token":"TEST_TOKEN","expires_in":7200}`))
		case "/channels/ec/basics/getfinderlivedata":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Query().Get("access_token") != "TEST_TOKEN" {
				t.Errorf("missing access_token")
			}
			var req GetFinderLiveDataListReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatal(err)
			}
			if req.StartDate != "2024-01-01" {
				t.Errorf("expected start_date 2024-01-01, got %s", req.StartDate)
			}
			_, _ = w.Write([]byte(`{"items":[{"date":"2024-01-01","view_count":100,"like_count":10,"share_count":5}],"total":1}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	resp, err := c.GetFinderLiveDataList(context.Background(), &GetFinderLiveDataListReq{
		StartDate: "2024-01-01",
		EndDate:   "2024-01-31",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Total != 1 {
		t.Errorf("got total=%d, want 1", resp.Total)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("got %d items, want 1", len(resp.Items))
	}
	if resp.Items[0].ViewCount != 100 {
		t.Errorf("got view_count=%d, want 100", resp.Items[0].ViewCount)
	}
}

func TestGetFinderList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/token":
			_, _ = w.Write([]byte(`{"access_token":"TEST_TOKEN","expires_in":7200}`))
		case "/channels/ec/basics/getfinderlist":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			_, _ = w.Write([]byte(`{"items":[{"finder_id":"finder_001","nickname":"TestFinder"}],"total":1}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	limit := 10
	resp, err := c.GetFinderList(context.Background(), &GetFinderListReq{Limit: &limit})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Total != 1 {
		t.Errorf("got total=%d, want 1", resp.Total)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("got %d items, want 1", len(resp.Items))
	}
	if resp.Items[0].FinderID != "finder_001" {
		t.Errorf("got finder_id=%s, want finder_001", resp.Items[0].FinderID)
	}
	if resp.Items[0].Nickname != "TestFinder" {
		t.Errorf("got nickname=%s, want TestFinder", resp.Items[0].Nickname)
	}
}
