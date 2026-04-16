package offiaccount

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ---------- Template message ----------

func TestSendTemplateMessage_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		if !strings.HasSuffix(r.URL.Path, "/cgi-bin/message/template/send") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["touser"] != "o1" || body["template_id"] != "tmpl1" {
			t.Errorf("unexpected body: %+v", body)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok","msg_id":12345}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.SendTemplateMessage(context.Background(), &SubscribeMessageRequest{
		ToUser:     "o1",
		TemplateID: "tmpl1",
		URL:        "https://example.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.MsgId != 12345 {
		t.Errorf("expected msgid 12345, got %d", resp.MsgId)
	}
}

func TestAddTemplate_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok","template_id":"tmpl_abc"}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.AddTemplate(context.Background(), "TM001", []string{"keyword1", "keyword2"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.TemplateId != "tmpl_abc" {
		t.Errorf("expected tmpl_abc, got %q", resp.TemplateId)
	}
}

func TestDeleteTemplate_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["template_id"] != "tmpl1" {
			t.Errorf("unexpected template_id: %v", body["template_id"])
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.DeleteTemplate(context.Background(), "tmpl1")
	if err != nil {
		t.Fatal(err)
	}
	if resp.ErrCode != 0 {
		t.Errorf("expected errcode 0, got %d", resp.ErrCode)
	}
}

func TestGetAllTemplates_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		_, _ = w.Write([]byte(`{"template_list":[{"template_id":"tmpl1","title":"Test"}]}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.GetAllTemplates(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.TemplateList) != 1 {
		t.Errorf("expected 1 template, got %d", len(resp.TemplateList))
	}
}

func TestGetIndustry_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		_, _ = w.Write([]byte(`{"primary_industry":{"first_class":"IT","second_class":"Software"},"secondary_industry":{"first_class":"Finance","second_class":"Banking"}}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.GetIndustry(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if resp.PrimaryIndustry.FirstClass != "IT" {
		t.Errorf("expected IT, got %q", resp.PrimaryIndustry.FirstClass)
	}
}

func TestSetIndustry_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	err := c.SetIndustry(context.Background(), "1", "2")
	if err != nil {
		t.Fatal(err)
	}
}

// ---------- Notify subscribe ----------

func TestTemplateSubscribe_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		if !strings.HasSuffix(r.URL.Path, "/cgi-bin/message/template/subscribe") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	err := c.TemplateSubscribe(context.Background(), &TemplateSubscribeReq{
		ToUser:     "o1",
		TemplateId: "tmpl1",
		Scene:      "1000",
		Title:      "test",
	})
	if err != nil {
		t.Fatal(err)
	}
}

// ---------- Notify new subscribe template ----------

func TestSendNewSubscribeMsg_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		if !strings.HasSuffix(r.URL.Path, "/cgi-bin/message/subscribe/bizsend") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	err := c.SendNewSubscribeMsg(context.Background(), &SubscribeMsg{
		ToUser:     "o1",
		TemplateId: "tmpl1",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetCategory_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		_, _ = w.Write([]byte(`{"data":[{"id":616,"name":"IT科技"}]}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.GetCategory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Data) != 1 {
		t.Errorf("expected 1 category, got %d", len(resp.Data))
	}
}

// ---------- Notify auto replies ----------

func TestGetCurrentAutoReplyInfo_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		_, _ = w.Write([]byte(`{"is_add_friend_reply_open":1,"is_autoreply_open":1}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.GetCurrentAutoReplyInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if resp.IsAddFriendReplyOpen != 1 {
		t.Errorf("expected is_add_friend_reply_open=1, got %d", resp.IsAddFriendReplyOpen)
	}
}

// ---------- Notify mass message ----------

func TestGetSpeed_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		_, _ = w.Write([]byte(`{"speed":3,"realspeed":15}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.GetSpeed(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if resp.Speed != 3 {
		t.Errorf("expected speed 3, got %d", resp.Speed)
	}
}

func TestSetSpeed_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.SetSpeed(context.Background(), 2)
	if err != nil {
		t.Fatal(err)
	}
	if resp.ErrCode != 0 {
		t.Errorf("expected errcode 0, got %d", resp.ErrCode)
	}
}

func TestGetMassMsg_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		_, _ = w.Write([]byte(`{"msg_id":1234,"msg_status":"SEND_SUCCESS"}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.GetMassMsg(context.Background(), 1234)
	if err != nil {
		t.Fatal(err)
	}
	if resp.MsgStatus != "SEND_SUCCESS" {
		t.Errorf("expected SEND_SUCCESS, got %q", resp.MsgStatus)
	}
}

func TestDeleteMassMsg_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	err := c.DeleteMassMsg(context.Background(), &DeleteMassMsgRequest{MsgId: 1234})
	if err != nil {
		t.Fatal(err)
	}
}

func TestMassSend_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		if !strings.HasSuffix(r.URL.Path, "/cgi-bin/message/mass/send") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok","msg_id":5678,"msg_data_id":9012}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.MassSend(context.Background(), &MassSendRequest{
		ToUser:  []string{"o1", "o2"},
		MsgType: "text",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.MsgId != 5678 {
		t.Errorf("expected msg_id 5678, got %d", resp.MsgId)
	}
}

func TestSendAll_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok","msg_id":1111,"msg_data_id":2222}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.SendAll(context.Background(), &MassSendByTagRequest{
		MsgType: "text",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.MsgID != 1111 {
		t.Errorf("expected msg_id 1111, got %d", resp.MsgID)
	}
}
