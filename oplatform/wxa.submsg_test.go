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

func TestWxaAdmin_GetSubscribeCategory(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxaapi/newtmpl/getcategory") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"data":[{"id":1,"name":"工具"},{"id":2,"name":"教育"}]}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	resp, err := w.GetSubscribeCategory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Data) != 2 || resp.Data[0].Name != "工具" {
		t.Errorf("unexpected: %+v", resp)
	}
}

func TestWxaAdmin_GetPubTemplateTitles(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxaapi/newtmpl/getpubtemplatetitles") {
			t.Errorf("path: %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("ids") != "2-3" {
			t.Errorf("ids: %q", q.Get("ids"))
		}
		if q.Get("start") != "0" || q.Get("limit") != "30" {
			t.Errorf("pagination: %v", q)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"count":1,"data":[{"tid":99,"title":"订单已发货","type":2,"categoryId":"2"}]}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	resp, err := w.GetPubTemplateTitles(context.Background(), "2-3", 0, 30)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Count != 1 || len(resp.Data) != 1 || resp.Data[0].TID != 99 {
		t.Errorf("unexpected: %+v", resp)
	}
}

func TestWxaAdmin_GetPubTemplateKeywords(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxaapi/newtmpl/getpubtemplatekeywords") {
			t.Errorf("path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("tid") != "99" {
			t.Errorf("tid: %q", r.URL.Query().Get("tid"))
		}
		_, _ = w.Write([]byte(`{"errcode":0,"count":2,"data":[{"kid":1,"name":"订单号","example":"1234","rule":"thing"},{"kid":2,"name":"金额","example":"10","rule":"amount"}]}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	resp, err := w.GetPubTemplateKeywords(context.Background(), "99")
	if err != nil {
		t.Fatal(err)
	}
	if resp.Count != 2 || len(resp.Data) != 2 || resp.Data[0].Name != "订单号" {
		t.Errorf("unexpected: %+v", resp)
	}
}

func TestWxaAdmin_AddSubscribeTemplate(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxaapi/newtmpl/addtemplate") {
			t.Errorf("path: %s", r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0,"priTmplId":"PTID_1"}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	resp, err := w.AddSubscribeTemplate(context.Background(), &WxaAddSubscribeTemplateReq{
		TID:       "99",
		KidList:   []int{1, 2},
		SceneDesc: "订单通知",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.PriTmplID != "PTID_1" {
		t.Errorf("priTmplId: %q", resp.PriTmplID)
	}
	if body["tid"] != "99" {
		t.Errorf("body tid: %+v", body)
	}
	kids, _ := body["kidList"].([]any)
	if len(kids) != 2 {
		t.Errorf("body kidList: %+v", body)
	}
}

func TestWxaAdmin_DeleteSubscribeTemplate(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxaapi/newtmpl/deltemplate") {
			t.Errorf("path: %s", r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	if err := w.DeleteSubscribeTemplate(context.Background(), "PTID_1"); err != nil {
		t.Fatal(err)
	}
	if body["priTmplId"] != "PTID_1" {
		t.Errorf("body: %+v", body)
	}
}

func TestWxaAdmin_ListSubscribeTemplates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxaapi/newtmpl/gettemplate") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"data":[{"priTmplId":"P1","title":"订单通知","content":"{{c1.DATA}}","example":"xxx","type":2}]}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	list, err := w.ListSubscribeTemplates(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Data) != 1 || list.Data[0].PriTmplID != "P1" {
		t.Errorf("unexpected: %+v", list)
	}
}

func TestWxaAdmin_SendSubscribeMessage(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/message/subscribe/send") {
			t.Errorf("path: %s", r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	err := w.SendSubscribeMessage(context.Background(), &WxaSendSubscribeReq{
		ToUser:     "OPENID_1",
		TemplateID: "P1",
		Page:       "pages/order/detail",
		Data: map[string]WxaSubscribeTemplateDataField{
			"c1": {Value: "12345"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if body["touser"] != "OPENID_1" || body["template_id"] != "P1" {
		t.Errorf("body: %+v", body)
	}
	data, _ := body["data"].(map[string]any)
	c1, _ := data["c1"].(map[string]any)
	if c1["value"] != "12345" {
		t.Errorf("body.data: %+v", body)
	}
}
