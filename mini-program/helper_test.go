package mini_program

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

// helperFakeServer wires a fake WeChat backend that always serves an
// access_token for /cgi-bin/token and routes everything else through the
// supplied handler.
func helperFakeServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/cgi-bin/token") {
			_, _ = w.Write([]byte(`{"access_token":"TOK","expires_in":7200}`))
			return
		}
		handler(w, r)
	}))
}

func helperTestClient(t *testing.T, baseURL string) *Client {
	t.Helper()
	c, err := NewClient(Config{AppId: "wx", AppSecret: "sec"},
		WithHTTP(utils.NewHTTP(baseURL, utils.WithTimeout(3*time.Second))))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// === doPost ================================================================

func TestDoPost_ReturnsTypedAPIError(t *testing.T) {
	srv := helperFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":40001,"errmsg":"invalid credential"}`))
	})
	defer srv.Close()

	c := helperTestClient(t, srv.URL)
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

func TestDoPost_FailsLoudOnNonJSONBody(t *testing.T) {
	srv := helperFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<html><body>502 Bad Gateway</body></html>`))
	})
	defer srv.Close()

	c := helperTestClient(t, srv.URL)
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

func TestDoPost_DecodesSuccessIntoOut(t *testing.T) {
	srv := helperFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"openid":"o1","session_key":"sk"}`))
	})
	defer srv.Close()

	c := helperTestClient(t, srv.URL)
	var out struct {
		OpenID     string `json:"openid"`
		SessionKey string `json:"session_key"`
	}
	if err := c.doPost(context.Background(), "/x", nil, &out); err != nil {
		t.Fatalf("doPost: %v", err)
	}
	if out.OpenID != "o1" || out.SessionKey != "sk" {
		t.Errorf("unexpected out: %+v", out)
	}
}

// === doGet ================================================================
//
// doGet previously delegated straight to c.http.Get, which json.Unmarshal'd
// the response into the caller's struct. WeChat error envelopes
// (`{"errcode":N,"errmsg":"..."}`) silently decode to a zero-valued out
// struct and return nil, hiding the failure. These tests pin the fix.

func TestDoGet_ReturnsTypedAPIError(t *testing.T) {
	srv := helperFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":40001,"errmsg":"invalid credential"}`))
	})
	defer srv.Close()

	c := helperTestClient(t, srv.URL)
	var out map[string]any
	err := c.doGet(context.Background(), "/some/path", nil, &out)
	if err == nil {
		t.Fatal("expected error")
	}
	var ae *APIError
	if !errors.As(err, &ae) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if ae.ErrCode != 40001 || ae.Path != "/some/path" {
		t.Errorf("unexpected APIError: %+v", ae)
	}
}

func TestDoGet_FailsLoudOnNonJSONBody(t *testing.T) {
	srv := helperFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<html><body>upstream broken</body></html>`))
	})
	defer srv.Close()

	c := helperTestClient(t, srv.URL)
	var out map[string]any
	err := c.doGet(context.Background(), "/some/path", nil, &out)
	if err == nil {
		t.Fatal("expected error for non-JSON body, got nil")
	}
	if !strings.Contains(err.Error(), "decode envelope") {
		t.Errorf("error should mention decode envelope: %v", err)
	}
}

func TestDoGet_DecodesSuccessIntoOut(t *testing.T) {
	srv := helperFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"openid":"o1","session_key":"sk"}`))
	})
	defer srv.Close()

	c := helperTestClient(t, srv.URL)
	var out struct {
		OpenID     string `json:"openid"`
		SessionKey string `json:"session_key"`
	}
	if err := c.doGet(context.Background(), "/x", nil, &out); err != nil {
		t.Fatalf("doGet: %v", err)
	}
	if out.OpenID != "o1" || out.SessionKey != "sk" {
		t.Errorf("unexpected out: %+v", out)
	}
}
