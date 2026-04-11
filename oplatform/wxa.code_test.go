package oplatform

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWxaAdmin_Commit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/commit") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	err := w.Commit(context.Background(), &WxaCommitReq{
		TemplateID:  1,
		UserVersion: "1.0.0",
		UserDesc:    "initial",
		ExtJSON:     `{"extEnable":true}`,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestWxaAdmin_GetPage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/get_page") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"page_list":["pages/index/index","pages/detail/detail"]}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	resp, err := w.GetPage(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.PageList) != 2 {
		t.Errorf("unexpected: %+v", resp)
	}
}

func TestWxaAdmin_GetQrcode_Binary(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/get_qrcode") {
			t.Errorf("path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("path") != "pages/index" {
			t.Errorf("path query: %q", r.URL.Query().Get("path"))
		}
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = w.Write([]byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10})
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	body, ct, err := w.GetQrcode(context.Background(), "pages/index")
	if err != nil {
		t.Fatal(err)
	}
	if ct != "image/jpeg" {
		t.Errorf("content-type: %q", ct)
	}
	if len(body) != 6 || body[0] != 0xFF {
		t.Errorf("body: %v", body)
	}
}

func TestWxaAdmin_GetQrcode_Errcode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errcode":85024,"errmsg":"no test version"}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	_, _, err := w.GetQrcode(context.Background(), "pages/index")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestWxaAdmin_GetCodeCategory(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/get_category") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"category_list":[{"first_class":"A","second_class":"B","first_id":1,"second_id":2}]}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	resp, err := w.GetCodeCategory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.CategoryList) != 1 || resp.CategoryList[0].FirstClass != "A" {
		t.Errorf("unexpected: %+v", resp)
	}
}
