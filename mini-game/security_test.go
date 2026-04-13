package mini_game

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMsgSecCheck_ErrcodeError(t *testing.T) {
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
	_, err := c.MsgSecCheck(context.Background(), &MsgSecCheckReq{
		Content: "hello", Version: 2, Scene: 1, OpenID: "oABC123",
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

func TestMsgSecCheck(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/token":
			_, _ = w.Write([]byte(`{"access_token":"TEST_TOKEN","expires_in":7200}`))
		case "/wxa/msg_sec_check":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Query().Get("access_token") != "TEST_TOKEN" {
				t.Errorf("missing access_token")
			}
			var req MsgSecCheckReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatal(err)
			}
			if req.Content != "hello world" {
				t.Errorf("expected content hello world, got %s", req.Content)
			}
			if req.OpenID != "oABC123" {
				t.Errorf("expected openid oABC123, got %s", req.OpenID)
			}
			_, _ = w.Write([]byte(`{"result":{"suggest":"pass","label":100}}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	resp, err := c.MsgSecCheck(context.Background(), &MsgSecCheckReq{
		Content: "hello world",
		Version: 2,
		Scene:   1,
		OpenID:  "oABC123",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Result.Suggest != "pass" {
		t.Errorf("unexpected suggest: %q", resp.Result.Suggest)
	}
	if resp.Result.Label != 100 {
		t.Errorf("unexpected label: %d", resp.Result.Label)
	}
}
