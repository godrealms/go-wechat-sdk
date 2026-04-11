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

func TestWxaAdmin_GetCategory(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/wxopen/getcategory") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"categories_list":[{"first":1,"second":2,"first_name":"工具","second_name":"办公"}]}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	resp, err := w.GetCategory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.CategoriesList) != 1 || resp.CategoriesList[0].First != 1 {
		t.Errorf("unexpected: %+v", resp)
	}
}

func TestWxaAdmin_GetAllCategories(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/wxopen/getallcategories") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"categories_list":[{"id":1,"name":"root","level":1,"father":0,"children":[2,3]}]}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	resp, err := w.GetAllCategories(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.CategoriesList) != 1 || resp.CategoriesList[0].Name != "root" {
		t.Errorf("unexpected: %+v", resp)
	}
}

func TestWxaAdmin_AddCategory(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/wxopen/addcategory") {
			t.Errorf("path: %s", r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	err := w.AddCategory(context.Background(), &WxaAddCategoryReq{
		Categories: []WxaCategoryItem{{First: 1, Second: 2}},
	})
	if err != nil {
		t.Fatal(err)
	}
	cats, _ := body["categories"].([]any)
	if len(cats) != 1 {
		t.Errorf("body categories: %+v", body)
	}
}

func TestWxaAdmin_DeleteCategory(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/wxopen/deletecategory") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	if err := w.DeleteCategory(context.Background(), 1, 2); err != nil {
		t.Fatal(err)
	}
}

func TestWxaAdmin_ModifyCategory(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/wxopen/modifycategory") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	err := w.ModifyCategory(context.Background(), &WxaModifyCategoryReq{First: 1, Second: 2})
	if err != nil {
		t.Fatal(err)
	}
}

func TestWxaAdmin_ModifyCategory_Errcode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":85003,"errmsg":"too frequent"}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	err := w.ModifyCategory(context.Background(), &WxaModifyCategoryReq{First: 1, Second: 2})
	var werr *WeixinError
	if !errors.As(err, &werr) || werr.ErrCode != 85003 {
		t.Errorf("expected WeixinError 85003, got %v", err)
	}
}
