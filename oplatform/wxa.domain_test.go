package oplatform

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWxaAdmin_ModifyServerDomain(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/modify_domain") {
			t.Errorf("path: %s", r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0,"requestdomain":["https://a.example.com"]}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	resp, err := w.ModifyServerDomain(context.Background(), &WxaModifyServerDomainReq{
		Action:        "add",
		Requestdomain: []string{"https://a.example.com"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if body["action"] != "add" {
		t.Errorf("body action: %+v", body)
	}
	if len(resp.Requestdomain) != 1 {
		t.Errorf("resp: %+v", resp)
	}
}

func TestWxaAdmin_SetWebviewDomain(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/setwebviewdomain") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	err := w.SetWebviewDomain(context.Background(), &WxaSetWebviewDomainReq{
		Action:        "set",
		Webviewdomain: []string{"https://webview.example.com"},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestWxaAdmin_GetDomainConfirmFile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/get_webview_confirmfile") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"file_name":"abc.txt","file_content":"xyz"}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	file, err := w.GetDomainConfirmFile(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if file.FileName != "abc.txt" || file.FileContent != "xyz" {
		t.Errorf("unexpected: %+v", file)
	}
}

func TestWxaAdmin_ModifyDomainDirectly(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/modify_domain_directly") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"requestdomain":["https://a.example.com"]}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	resp, err := w.ModifyDomainDirectly(context.Background(), &WxaModifyDomainDirectlyReq{
		Action:        "set",
		Requestdomain: []string{"https://a.example.com"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Requestdomain) != 1 {
		t.Errorf("resp: %+v", resp)
	}
}
