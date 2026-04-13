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

func TestMsgSecCheck(t *testing.T) {
	srv := httptest.NewServer(tokenHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/wxa/msg_sec_check" {
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
		for _, field := range []string{"content", "version", "scene", "openid"} {
			if _, ok := req[field]; !ok {
				t.Errorf("body missing %q field", field)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"trace_id": "TR123",
			"result": {"suggest": "pass", "label": 100},
			"detail": [{"strategy": "content_model", "errcode": 0, "suggest": "pass", "label": 100, "prob": 90, "keyword": ""}]
		}`))
	}))
	defer srv.Close()

	c, err := NewClient(Config{AppId: "wx", AppSecret: "sec"},
		WithHTTP(utils.NewHTTP(srv.URL, utils.WithTimeout(time.Second*3))))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := c.MsgSecCheck(context.Background(), &MsgSecCheckReq{
		Content: "hello world",
		Version: 2,
		Scene:   1,
		OpenID:  "oABC123",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.TraceID != "TR123" {
		t.Errorf("unexpected trace_id: %q", resp.TraceID)
	}
	if resp.Result.Suggest != "pass" {
		t.Errorf("unexpected suggest: %q", resp.Result.Suggest)
	}
	if resp.Result.Label != 100 {
		t.Errorf("unexpected label: %d", resp.Result.Label)
	}
	if len(resp.Detail) != 1 {
		t.Errorf("expected 1 detail, got %d", len(resp.Detail))
	}
}

func TestMediaCheckAsync(t *testing.T) {
	srv := httptest.NewServer(tokenHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/wxa/media_check_async" {
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
		for _, field := range []string{"media_url", "media_type"} {
			if _, ok := req[field]; !ok {
				t.Errorf("body missing %q field", field)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"trace_id":"MEDIA_TR456"}`))
	}))
	defer srv.Close()

	c, err := NewClient(Config{AppId: "wx", AppSecret: "sec"},
		WithHTTP(utils.NewHTTP(srv.URL, utils.WithTimeout(time.Second*3))))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := c.MediaCheckAsync(context.Background(), &MediaCheckAsyncReq{
		MediaURL:  "https://example.com/image.jpg",
		MediaType: 2,
		Version:   2,
		Scene:     1,
		OpenID:    "oABC123",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.TraceID != "MEDIA_TR456" {
		t.Errorf("unexpected trace_id: %q", resp.TraceID)
	}
}
