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

// fakePNG is a minimal fake PNG byte sequence for testing.
var fakePNG = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

// tokenHandler returns a handler that seeds access_token for /cgi-bin/token
// and routes other paths to the provided apiHandler.
func tokenHandler(t *testing.T, apiHandler http.HandlerFunc) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/token" {
			_, _ = w.Write([]byte(`{"access_token":"TOK","expires_in":7200}`))
			return
		}
		apiHandler(w, r)
	}
}

func TestGetWxaCode(t *testing.T) {
	srv := httptest.NewServer(tokenHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/wxa/getwxacode" {
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
		if _, ok := req["path"]; !ok {
			t.Error("body missing 'path' field")
		}
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(fakePNG)
	}))
	defer srv.Close()

	c, err := NewClient(Config{AppId: "wx", AppSecret: "sec"},
		WithHTTP(utils.NewHTTP(srv.URL, utils.WithTimeout(time.Second*3))))
	if err != nil {
		t.Fatal(err)
	}
	got, err := c.GetWxaCode(context.Background(), &GetWxaCodeReq{Path: "pages/index/index"})
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(fakePNG) {
		t.Errorf("unexpected bytes: %v", got)
	}
}

func TestGetWxaCodeUnlimit(t *testing.T) {
	srv := httptest.NewServer(tokenHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/wxa/getwxacodeunlimit" {
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
		if _, ok := req["scene"]; !ok {
			t.Error("body missing 'scene' field")
		}
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(fakePNG)
	}))
	defer srv.Close()

	c, err := NewClient(Config{AppId: "wx", AppSecret: "sec"},
		WithHTTP(utils.NewHTTP(srv.URL, utils.WithTimeout(time.Second*3))))
	if err != nil {
		t.Fatal(err)
	}
	got, err := c.GetWxaCodeUnlimit(context.Background(), &GetWxaCodeUnlimitReq{Scene: "a=1"})
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(fakePNG) {
		t.Errorf("unexpected bytes: %v", got)
	}
}

func TestCreateQRCode(t *testing.T) {
	srv := httptest.NewServer(tokenHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/wxaapp/createwxaqrcode" {
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
		if _, ok := req["path"]; !ok {
			t.Error("body missing 'path' field")
		}
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(fakePNG)
	}))
	defer srv.Close()

	c, err := NewClient(Config{AppId: "wx", AppSecret: "sec"},
		WithHTTP(utils.NewHTTP(srv.URL, utils.WithTimeout(time.Second*3))))
	if err != nil {
		t.Fatal(err)
	}
	got, err := c.CreateQRCode(context.Background(), &CreateQRCodeReq{Path: "pages/index/index"})
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(fakePNG) {
		t.Errorf("unexpected bytes: %v", got)
	}
}
