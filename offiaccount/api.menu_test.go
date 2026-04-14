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

// fixedToken is a TokenSource that always returns a constant token.
type fixedToken struct{ tok string }

func (f fixedToken) AccessToken(_ context.Context) (string, error) { return f.tok, nil }

func newMenuTestClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	h := utils.NewHTTP(srv.URL, utils.WithTimeout(3*time.Second))
	return NewClient(context.Background(), &Config{AppId: "test", AppSecret: "secret"},
		WithHTTPClient(h),
		WithTokenSource(fixedToken{"FAKE_TOKEN"}),
	)
}

func TestCreateCustomMenu(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		status  int
		wantErr bool
		errMsg  string
	}{
		{
			name:    "success",
			body:    `{"errcode":0,"errmsg":"ok"}`,
			status:  200,
			wantErr: false,
		},
		{
			name:    "wechat errcode error",
			body:    `{"errcode":65001,"errmsg":"invalid menu type"}`,
			status:  200,
			wantErr: true,
			errMsg:  "invalid menu type",
		},
		{
			name:    "network error",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var srv *httptest.Server
			if tc.name == "network error" {
				srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if !strings.Contains(r.URL.RawQuery, "access_token=FAKE_TOKEN") {
						t.Errorf("missing access_token in request URL: %s", r.URL.String())
					}
				}))
				srv.Close()
			} else {
				srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if !strings.Contains(r.URL.RawQuery, "access_token=FAKE_TOKEN") {
						t.Errorf("missing access_token in request URL: %s", r.URL.String())
					}
					w.WriteHeader(tc.status)
					_, _ = w.Write([]byte(tc.body))
				}))
				defer srv.Close()
			}
			c := newMenuTestClient(t, srv)
			err := c.CreateCustomMenu(context.Background(), &CreateMenuButton{})
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tc.errMsg != "" && err != nil && !strings.Contains(err.Error(), tc.errMsg) {
				t.Errorf("expected error containing %q, got %v", tc.errMsg, err)
			}
		})
	}
}

func TestGetMenu(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{"success", `{"menu":{"button":[]}}`, false},
		{"network error", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var srv *httptest.Server
			if tc.name == "network error" {
				srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if !strings.Contains(r.URL.RawQuery, "access_token=FAKE_TOKEN") {
						t.Errorf("missing access_token in request URL: %s", r.URL.String())
					}
				}))
				srv.Close()
			} else {
				srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if !strings.Contains(r.URL.RawQuery, "access_token=FAKE_TOKEN") {
						t.Errorf("missing access_token in request URL: %s", r.URL.String())
					}
					w.WriteHeader(200)
					_, _ = w.Write([]byte(tc.body))
				}))
				defer srv.Close()
			}
			c := newMenuTestClient(t, srv)
			result, err := c.GetMenu(context.Background())
			if tc.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result == nil {
				t.Error("expected non-nil result")
			}
		})
	}
}

func TestDeleteMenu(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{"success", `{"errcode":0,"errmsg":"ok"}`, false},
		{"errcode error", `{"errcode":40001,"errmsg":"invalid credential"}`, true},
		{"network error", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var srv *httptest.Server
			if tc.name == "network error" {
				srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if !strings.Contains(r.URL.RawQuery, "access_token=FAKE_TOKEN") {
						t.Errorf("missing access_token in request URL: %s", r.URL.String())
					}
				}))
				srv.Close()
			} else {
				srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if !strings.Contains(r.URL.RawQuery, "access_token=FAKE_TOKEN") {
						t.Errorf("missing access_token in request URL: %s", r.URL.String())
					}
					w.WriteHeader(200)
					_, _ = w.Write([]byte(tc.body))
				}))
				defer srv.Close()
			}
			c := newMenuTestClient(t, srv)
			err := c.DeleteMenu(context.Background())
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestGetCurrentSelfMenuInfo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "access_token=FAKE_TOKEN") {
			t.Errorf("missing access_token in request URL: %s", r.URL.String())
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"is_menu_open":1,"selfmenu_info":{"button":[]}}`))
	}))
	defer srv.Close()

	c := newMenuTestClient(t, srv)
	result, err := c.GetCurrentSelfMenuInfo(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestAddConditionalMenu(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{"success", `{"errcode":0,"errmsg":"ok","menuid":"502394"}`, false},
		{"errcode error", `{"errcode":65301,"errmsg":"menu count limit"}`, true},
		{"network error", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var srv *httptest.Server
			if tc.name == "network error" {
				srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if !strings.Contains(r.URL.RawQuery, "access_token=FAKE_TOKEN") {
						t.Errorf("missing access_token in request URL: %s", r.URL.String())
					}
				}))
				srv.Close()
			} else {
				srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if !strings.Contains(r.URL.RawQuery, "access_token=FAKE_TOKEN") {
						t.Errorf("missing access_token in request URL: %s", r.URL.String())
					}
					w.WriteHeader(200)
					_, _ = w.Write([]byte(tc.body))
				}))
				defer srv.Close()
			}
			c := newMenuTestClient(t, srv)
			_, err := c.AddConditionalMenu(context.Background(), &ConditionalMenu{})
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestDeleteConditionalMenu(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "access_token=FAKE_TOKEN") {
			t.Errorf("missing access_token in request URL: %s", r.URL.String())
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()

	c := newMenuTestClient(t, srv)
	_, err := c.DeleteConditionalMenu(context.Background(), "502394")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTryMatchMenu(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "access_token=FAKE_TOKEN") {
			t.Errorf("missing access_token in request URL: %s", r.URL.String())
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"menu":{"button":[]}}`))
	}))
	defer srv.Close()

	c := newMenuTestClient(t, srv)
	result, err := c.TryMatchMenu(context.Background(), "oUser123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}
