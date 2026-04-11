package oplatform

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWxaAdmin_ApplyPlugin(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/plugin") {
			t.Errorf("path: %s", r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	if err := w.ApplyPlugin(context.Background(), "wxPLUG"); err != nil {
		t.Fatal(err)
	}
	if body["action"] != "apply" || body["plugin_appid"] != "wxPLUG" {
		t.Errorf("body: %+v", body)
	}
}

func TestWxaAdmin_ApplyPlugin_Errcode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":89236,"errmsg":"duplicate apply"}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	err := w.ApplyPlugin(context.Background(), "wxPLUG")
	var werr *WeixinError
	if !errors.As(err, &werr) || werr.ErrCode != 89236 {
		t.Errorf("expected 89236, got %v", err)
	}
}

func TestWxaAdmin_ListPlugins(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0,"plugin_list":[{"appid":"wxA","status":2,"nickname":"N1"},{"appid":"wxB","status":1}]}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	list, err := w.ListPlugins(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if body["action"] != "list" {
		t.Errorf("body: %+v", body)
	}
	if len(list.PluginList) != 2 || list.PluginList[0].AppID != "wxA" {
		t.Errorf("list: %+v", list)
	}
}

func TestWxaAdmin_UnbindPlugin(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	if err := w.UnbindPlugin(context.Background(), "wxPLUG"); err != nil {
		t.Fatal(err)
	}
	if body["action"] != "unbind" || body["plugin_appid"] != "wxPLUG" {
		t.Errorf("body: %+v", body)
	}
}

func TestWxaAdmin_GetPluginDevApplyList(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/devplugin") {
			t.Errorf("path: %s", r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0,"apply_list":[{"appid":"wxUSER","status":1,"nickname":"U1"}]}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	list, err := w.GetPluginDevApplyList(context.Background(), 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if body["action"] != "dev_apply_list" {
		t.Errorf("body: %+v", body)
	}
	if int(body["page"].(float64)) != 0 || int(body["num"].(float64)) != 10 {
		t.Errorf("page/num: %+v", body)
	}
	if len(list.ApplyList) != 1 || list.ApplyList[0].AppID != "wxUSER" {
		t.Errorf("list: %+v", list)
	}
}

func TestWxaAdmin_AgreeDevPlugin(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	if err := w.AgreeDevPlugin(context.Background(), "wxUSER"); err != nil {
		t.Fatal(err)
	}
	if body["action"] != "dev_agree" || body["appid"] != "wxUSER" {
		t.Errorf("body: %+v", body)
	}
}

func TestWxaAdmin_RefuseDevPlugin(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	if err := w.RefuseDevPlugin(context.Background(), "违反规范"); err != nil {
		t.Fatal(err)
	}
	if body["action"] != "dev_refuse" || body["reason"] != "违反规范" {
		t.Errorf("body: %+v", body)
	}
}

func TestWxaAdmin_DeleteDevPlugin(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	if err := w.DeleteDevPlugin(context.Background(), "wxUSER"); err != nil {
		t.Fatal(err)
	}
	if body["action"] != "dev_delete" || body["appid"] != "wxUSER" {
		t.Errorf("body: %+v", body)
	}
}
