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

func TestWxaAdmin_SendCustomerMessage_Text(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/message/custom/send") {
			t.Errorf("path: %s", r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	err := w.SendCustomerMessage(context.Background(), &WxaSendCustomerMessageReq{
		ToUser:  "OPENID_1",
		MsgType: "text",
		Text:    &WxaCustomerTextPayload{Content: "hello"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if body["touser"] != "OPENID_1" || body["msgtype"] != "text" {
		t.Errorf("body top: %+v", body)
	}
	text, _ := body["text"].(map[string]any)
	if text["content"] != "hello" {
		t.Errorf("body.text: %+v", body)
	}
	if _, ok := body["image"]; ok {
		t.Errorf("image should be omitted when nil")
	}
}

func TestWxaAdmin_SendCustomerMessage_Image(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	err := w.SendCustomerMessage(context.Background(), &WxaSendCustomerMessageReq{
		ToUser:  "OPENID_2",
		MsgType: "image",
		Image:   &WxaCustomerImagePayload{MediaID: "MEDIA_X"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if body["msgtype"] != "image" {
		t.Errorf("msgtype: %+v", body)
	}
	img, _ := body["image"].(map[string]any)
	if img["media_id"] != "MEDIA_X" {
		t.Errorf("body.image: %+v", body)
	}
}

func TestWxaAdmin_SendCustomerMessage_MiniProgramPage(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	err := w.SendCustomerMessage(context.Background(), &WxaSendCustomerMessageReq{
		ToUser:  "OPENID_3",
		MsgType: "miniprogrampage",
		MiniProgramPage: &WxaCustomerMiniProgramPagePayload{
			Title:        "订单详情",
			Pagepath:     "pages/order/detail?id=123",
			ThumbMediaID: "THUMB_1",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if body["msgtype"] != "miniprogrampage" {
		t.Errorf("msgtype: %+v", body)
	}
	mpp, _ := body["miniprogrampage"].(map[string]any)
	if mpp["title"] != "订单详情" || mpp["pagepath"] != "pages/order/detail?id=123" {
		t.Errorf("body.miniprogrampage: %+v", body)
	}
}

func TestWxaAdmin_SendCustomerTyping(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/message/custom/typing") {
			t.Errorf("path: %s", r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	if err := w.SendCustomerTyping(context.Background(), "OPENID_1", "Typing"); err != nil {
		t.Fatal(err)
	}
	if body["touser"] != "OPENID_1" || body["command"] != "Typing" {
		t.Errorf("body: %+v", body)
	}
}
