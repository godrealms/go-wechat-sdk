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

func TestGenerateScheme(t *testing.T) {
	srv := httptest.NewServer(tokenHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/wxa/generatescheme" {
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
		if _, ok := req["jump_wxa"]; !ok {
			t.Error("body missing 'jump_wxa' field")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok","openlink":"weixin://dl/business/?t=xxx"}`))
	}))
	defer srv.Close()

	c, err := NewClient(Config{AppId: "wx", AppSecret: "sec"},
		WithHTTP(utils.NewHTTP(srv.URL, utils.WithTimeout(time.Second*3))))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := c.GenerateScheme(context.Background(), &GenerateSchemeReq{
		JumpWxa: &JumpWxa{Path: "pages/index/index", Query: "a=1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.OpenLink != "weixin://dl/business/?t=xxx" {
		t.Errorf("unexpected openlink: %q", resp.OpenLink)
	}
}

func TestGenerateUrlLink(t *testing.T) {
	srv := httptest.NewServer(tokenHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/wxa/generate_urllink" {
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
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok","url_link":"https://wxaurl.cn/xxx"}`))
	}))
	defer srv.Close()

	c, err := NewClient(Config{AppId: "wx", AppSecret: "sec"},
		WithHTTP(utils.NewHTTP(srv.URL, utils.WithTimeout(time.Second*3))))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := c.GenerateUrlLink(context.Background(), &GenerateUrlLinkReq{
		Path:  "pages/index/index",
		Query: "a=1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.URLLink != "https://wxaurl.cn/xxx" {
		t.Errorf("unexpected url_link: %q", resp.URLLink)
	}
}

func TestGenerateShortLink(t *testing.T) {
	srv := httptest.NewServer(tokenHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/wxa/genwxashortlink" {
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
		if _, ok := req["page_url"]; !ok {
			t.Error("body missing 'page_url' field")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok","link":"https://wxmpurl.cn/xxx"}`))
	}))
	defer srv.Close()

	c, err := NewClient(Config{AppId: "wx", AppSecret: "sec"},
		WithHTTP(utils.NewHTTP(srv.URL, utils.WithTimeout(time.Second*3))))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := c.GenerateShortLink(context.Background(), &GenerateShortLinkReq{
		PageURL:   "pages/index/index",
		PageTitle: "首页",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Link != "https://wxmpurl.cn/xxx" {
		t.Errorf("unexpected link: %q", resp.Link)
	}
}
