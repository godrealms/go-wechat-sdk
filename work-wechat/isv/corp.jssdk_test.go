package isv

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetJSAPITicket(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method: %s", r.Method)
		}
		if r.URL.Path != "/cgi-bin/get_jsapi_ticket" {
			t.Errorf("path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"errcode":    0,
			"errmsg":     "ok",
			"ticket":     "kgt8ON7yVITDhtdwci0qeT1D",
			"expires_in": 7200,
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	resp, err := cc.GetJSAPITicket(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if resp.Ticket != "kgt8ON7yVITDhtdwci0qeT1D" {
		t.Errorf("Ticket: %q", resp.Ticket)
	}
	if resp.ExpiresIn != 7200 {
		t.Errorf("ExpiresIn: %d", resp.ExpiresIn)
	}
}

func TestGetAgentConfigTicket(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method: %s", r.Method)
		}
		if r.URL.Path != "/cgi-bin/ticket/get" {
			t.Errorf("path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if got := r.URL.Query().Get("type"); got != "agent_config" {
			t.Errorf("type: %q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"errcode":    0,
			"errmsg":     "ok",
			"ticket":     "Hk5MBi7_bfGSG",
			"expires_in": 7200,
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	resp, err := cc.GetAgentConfigTicket(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if resp.Ticket != "Hk5MBi7_bfGSG" {
		t.Errorf("Ticket: %q", resp.Ticket)
	}
	if resp.ExpiresIn != 7200 {
		t.Errorf("ExpiresIn: %d", resp.ExpiresIn)
	}
}

func TestSignJSAPI(t *testing.T) {
	ticket := "sM4AOVdWfPE4DxkXGEs8VMCPGGVi4C3VM0P37wVUCFvkVAy_90u5h9nbSlYy3-Sl-HhTdfl2fzFy1AOcHKP7qg"
	nonceStr := "Wm3WZYTPz0wzccnW"
	timestamp := "1414587457"
	pageURL := "http://mp.weixin.qq.com?params=value"
	want := "0f9de62fce790f9a083d5c99e95740ceb90c27ed"

	got := SignJSAPI(ticket, nonceStr, timestamp, pageURL)
	if got != want {
		t.Errorf("SignJSAPI = %q, want %q", got, want)
	}
}
