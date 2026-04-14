package offiaccount

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

func newAnalysisTestClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	h := utils.NewHTTP(srv.URL, utils.WithTimeout(3*time.Second))
	return NewClient(context.Background(), &Config{AppId: "test", AppSecret: "secret"},
		WithHTTPClient(h),
		WithTokenSource(fixedToken{"FAKE_TOKEN"}),
	)
}

func TestGetUserSummary_Success(t *testing.T) {
	body := `{"list":[{"ref_date":"2024-01-01","user_source":0,"new_user":100,"cancel_user":5}]}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "access_token=FAKE_TOKEN") {
			t.Errorf("missing access_token in request URL: %s", r.URL.String())
		}
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	c := newAnalysisTestClient(t, srv)
	result, err := c.GetUserSummary("2024-01-01", "2024-01-07")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestGetUserSummary_NetworkError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "access_token=FAKE_TOKEN") {
			t.Errorf("missing access_token in request URL: %s", r.URL.String())
		}
	}))
	srv.Close()
	c := newAnalysisTestClient(t, srv)
	_, err := c.GetUserSummary("2024-01-01", "2024-01-07")
	if err == nil {
		t.Error("expected network error")
	}
}

func TestGetUserSummary_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "access_token=FAKE_TOKEN") {
			t.Errorf("missing access_token in request URL: %s", r.URL.String())
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`not-json`))
	}))
	defer srv.Close()

	c := newAnalysisTestClient(t, srv)
	_, err := c.GetUserSummary("2024-01-01", "2024-01-07")
	if err == nil {
		t.Error("expected unmarshal error for invalid JSON")
	}
}

func TestGetUserCumulate_Success(t *testing.T) {
	body := `{"list":[{"ref_date":"2024-01-01","cumulate_user":5000}]}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "access_token=FAKE_TOKEN") {
			t.Errorf("missing access_token in request URL: %s", r.URL.String())
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	c := newAnalysisTestClient(t, srv)
	result, err := c.GetUserCumulate("2024-01-01", "2024-01-07")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestGetUserCumulate_NetworkError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "access_token=FAKE_TOKEN") {
			t.Errorf("missing access_token in request URL: %s", r.URL.String())
		}
	}))
	srv.Close()
	c := newAnalysisTestClient(t, srv)
	_, err := c.GetUserCumulate("2024-01-01", "2024-01-07")
	if err == nil {
		t.Error("expected network error")
	}
}
