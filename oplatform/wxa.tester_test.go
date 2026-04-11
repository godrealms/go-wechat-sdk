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

func TestWxaAdmin_BindTester(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/bind_tester") {
			t.Errorf("path: %s", r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0,"userstr":"USER_STR_1"}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	resp, err := w.BindTester(context.Background(), "cool_wechat")
	if err != nil {
		t.Fatal(err)
	}
	if resp.UserStr != "USER_STR_1" {
		t.Errorf("userstr: %q", resp.UserStr)
	}
	if body["wechatid"] != "cool_wechat" {
		t.Errorf("body: %+v", body)
	}
}

func TestWxaAdmin_UnbindTester_ByWechatID(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/unbind_tester") {
			t.Errorf("path: %s", r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	if err := w.UnbindTester(context.Background(), "cool_wechat", ""); err != nil {
		t.Fatal(err)
	}
	if body["wechatid"] != "cool_wechat" {
		t.Errorf("body: %+v", body)
	}
	if _, ok := body["userstr"]; ok {
		t.Errorf("userstr should be omitted when empty")
	}
}

func TestWxaAdmin_UnbindTester_ByUserStr(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	if err := w.UnbindTester(context.Background(), "", "USER_STR_1"); err != nil {
		t.Fatal(err)
	}
	if body["userstr"] != "USER_STR_1" {
		t.Errorf("body: %+v", body)
	}
	if _, ok := body["wechatid"]; ok {
		t.Errorf("wechatid should be omitted when empty")
	}
}

func TestWxaAdmin_ListTesters(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/memberauth") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"members":[{"userstr":"U1"},{"userstr":"U2"}]}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	resp, err := w.ListTesters(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Members) != 2 || resp.Members[0].UserStr != "U1" {
		t.Errorf("unexpected: %+v", resp)
	}
}
