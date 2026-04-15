package mini_store

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestDoPost_ReturnsTypedAPIError ensures doPost surfaces a typed *APIError
// when WeChat returns a non-zero errcode, so callers can errors.As() it.
func TestDoPost_ReturnsTypedAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/cgi-bin/token") {
			_, _ = w.Write([]byte(`{"access_token":"TOK","expires_in":7200}`))
			return
		}
		_, _ = w.Write([]byte(`{"errcode":9101000,"errmsg":"no permission"}`))
	}))
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
	if ae.ErrCode != 9101000 {
		t.Errorf("ErrCode = %d, want 9101000", ae.ErrCode)
	}
	if ae.ErrMsg != "no permission" {
		t.Errorf("ErrMsg = %q, want %q", ae.ErrMsg, "no permission")
	}
	if ae.Path != "/some/path" {
		t.Errorf("Path = %q, want %q", ae.Path, "/some/path")
	}
}
