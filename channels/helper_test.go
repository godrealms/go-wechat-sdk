package channels

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func fakeServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/cgi-bin/token") {
			_, _ = w.Write([]byte(`{"access_token":"TOK","expires_in":7200}`))
			return
		}
		handler(w, r)
	}))
}

// TestDoPost_ReturnsTypedAPIError ensures doPost surfaces a typed *APIError
// when WeChat returns a non-zero errcode, so callers can errors.As() it.
func TestDoPost_ReturnsTypedAPIError(t *testing.T) {
	srv := fakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":40001,"errmsg":"invalid credential"}`))
	})
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	err := c.doPost(context.Background(), "/some/path", map[string]any{"k": "v"}, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var ae *APIError
	if !errors.As(err, &ae) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if ae.ErrCode != 40001 || ae.ErrMsg != "invalid credential" || ae.Path != "/some/path" {
		t.Errorf("unexpected APIError: %+v", ae)
	}
}

// TestDoPost_FailsLoudOnNonJSONBody verifies a malformed envelope (e.g. HTML
// error page from a proxy) returns a wrapped error rather than silently
// treating the body as success.
func TestDoPost_FailsLoudOnNonJSONBody(t *testing.T) {
	srv := fakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<html><body>502 Bad Gateway</body></html>`))
	})
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	var out map[string]any
	err := c.doPost(context.Background(), "/some/path", nil, &out)
	if err == nil {
		t.Fatal("expected error for non-JSON body, got nil")
	}
	if !strings.Contains(err.Error(), "decode envelope") {
		t.Errorf("error should mention decode envelope: %v", err)
	}
	var ae *APIError
	if errors.As(err, &ae) {
		t.Errorf("non-JSON body should NOT decode to *APIError, got %+v", ae)
	}
}

// TestDoPost_DecodesSuccessIntoOut verifies the success path decodes the
// JSON body into the caller's out struct.
func TestDoPost_DecodesSuccessIntoOut(t *testing.T) {
	srv := fakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"finder_username":"u1","display_name":"d1"}`))
	})
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	var out struct {
		FinderUsername string `json:"finder_username"`
		DisplayName    string `json:"display_name"`
	}
	if err := c.doPost(context.Background(), "/x", nil, &out); err != nil {
		t.Fatalf("doPost: %v", err)
	}
	if out.FinderUsername != "u1" || out.DisplayName != "d1" {
		t.Errorf("unexpected out: %+v", out)
	}
}
