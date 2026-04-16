package offiaccount

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ---------- Customer message ----------

func TestGetKFMsgList_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		if !strings.HasSuffix(r.URL.Path, "/customservice/msgrecord/getmsglist") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["starttime"] != float64(1000) || body["endtime"] != float64(2000) {
			t.Errorf("unexpected body: %+v", body)
		}
		_, _ = w.Write([]byte(`{"recordlist":[{"openid":"o1","text":"hi"}],"number":1,"msgid":100}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.GetKFMsgList(context.Background(), &KFGetMsgListRequest{
		StartTime: 1000, EndTime: 2000, MsgID: 0, Number: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.RecordList) != 1 {
		t.Errorf("expected 1 record, got %d", len(resp.RecordList))
	}
}

func TestSetKFTyping_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.SetKFTyping(context.Background(), &KFTypingRequest{
		ToUser:  "o1",
		Command: "Typing",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.ErrCode != 0 {
		t.Errorf("expected errcode 0, got %d", resp.ErrCode)
	}
}

func TestSendKFMessage_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		if !strings.HasSuffix(r.URL.Path, "/cgi-bin/message/custom/send") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["touser"] != "o1" || body["msgtype"] != "text" {
			t.Errorf("unexpected body: %+v", body)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.SendKFMessage(context.Background(), &KFMessage{
		ToUser:  "o1",
		MsgType: "text",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.ErrCode != 0 {
		t.Errorf("expected errcode 0, got %d", resp.ErrCode)
	}
}

// ---------- Customer session control ----------

func TestCreateKFSession_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		if !strings.HasSuffix(r.URL.Path, "/customservice/kfsession/create") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["kf_account"] != "kf1@test" || body["openid"] != "o1" {
			t.Errorf("unexpected body: %+v", body)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.CreateKFSession(context.Background(), "kf1@test", "o1")
	if err != nil {
		t.Fatal(err)
	}
	if resp.ErrCode != 0 {
		t.Errorf("expected errcode 0, got %d", resp.ErrCode)
	}
}

func TestCloseKFSession_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.CloseKFSession(context.Background(), "kf1@test", "o1")
	if err != nil {
		t.Fatal(err)
	}
	if resp.ErrCode != 0 {
		t.Errorf("expected errcode 0, got %d", resp.ErrCode)
	}
}

func TestGetKFCustomerSession_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		if !strings.HasSuffix(r.URL.Path, "/customservice/kfsession/getsession") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("openid") != "o1" {
			t.Errorf("expected openid=o1, got %q", r.URL.Query().Get("openid"))
		}
		_, _ = w.Write([]byte(`{"kf_account":"kf1@test","createtime":1234567890}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.GetKFCustomerSession(context.Background(), "o1")
	if err != nil {
		t.Fatal(err)
	}
	if resp.KfAccount != "kf1@test" {
		t.Errorf("expected kf1@test, got %q", resp.KfAccount)
	}
}

func TestGetKFSessionList_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		_, _ = w.Write([]byte(`{"sessionlist":[{"openid":"o1","kf_account":"kf1@test","createtime":1234567890}]}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.GetKFSessionList(context.Background(), "kf1@test")
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.SessionList) != 1 {
		t.Errorf("expected 1 session, got %d", len(resp.SessionList))
	}
}

func TestGetWaitCaseList_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		_, _ = w.Write([]byte(`{"count":2,"waitcaselist":[{"latest_time":1000,"openid":"o1"},{"latest_time":2000,"openid":"o2"}]}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.GetWaitCaseList(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if resp.Count != 2 {
		t.Errorf("expected count 2, got %d", resp.Count)
	}
}

// ---------- Customer servicer management ----------

func TestAddKFAccount_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		if !strings.HasSuffix(r.URL.Path, "/customservice/kfaccount/add") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.AddKFAccount(context.Background(), "kf1@test", "KF One")
	if err != nil {
		t.Fatal(err)
	}
	if resp.ErrCode != 0 {
		t.Errorf("expected errcode 0, got %d", resp.ErrCode)
	}
}

func TestUpdateKFAccount_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.UpdateKFAccount(context.Background(), "kf1@test", "KF Updated")
	if err != nil {
		t.Fatal(err)
	}
	if resp.ErrCode != 0 {
		t.Errorf("expected errcode 0, got %d", resp.ErrCode)
	}
}

func TestDelKFAccount_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.DelKFAccount(context.Background(), "kf1@test")
	if err != nil {
		t.Fatal(err)
	}
	if resp.ErrCode != 0 {
		t.Errorf("expected errcode 0, got %d", resp.ErrCode)
	}
}

func TestInviteKFWorker_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.InviteKFWorker(context.Background(), "kf1@test", "wxid_abc")
	if err != nil {
		t.Fatal(err)
	}
	if resp.ErrCode != 0 {
		t.Errorf("expected errcode 0, got %d", resp.ErrCode)
	}
}

func TestGetOnlineKFList_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		_, _ = w.Write([]byte(`{"kf_online_list":[{"kf_account":"kf1@test","status":1,"kf_id":"1001"}]}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.GetOnlineKFList(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.KFOnlineList) != 1 {
		t.Errorf("expected 1 online KF, got %d", len(resp.KFOnlineList))
	}
}

func TestGetKFList_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		_, _ = w.Write([]byte(`{"kf_list":[{"kf_account":"kf1@test","kf_nick":"KF One","kf_id":"1001"}]}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.GetKFList(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.KFList) != 1 {
		t.Errorf("expected 1 KF, got %d", len(resp.KFList))
	}
}

func TestUploadKFHeadImg_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		if r.URL.Query().Get("kf_account") != "kf1@test" {
			t.Errorf("expected kf_account=kf1@test, got %q", r.URL.Query().Get("kf_account"))
		}
		ct := r.Header.Get("Content-Type")
		if !strings.Contains(ct, "multipart/form-data") {
			t.Errorf("expected multipart, got %q", ct)
		}
		_ = r.ParseMultipartForm(1 << 20)
		f, fh, err := r.FormFile("media")
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		if fh.Filename != "avatar.jpg" {
			t.Errorf("expected filename avatar.jpg, got %q", fh.Filename)
		}
		data, _ := io.ReadAll(f)
		if string(data) != "fake image" {
			t.Errorf("unexpected data: %q", data)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.UploadKFHeadImg(context.Background(), "kf1@test", "avatar.jpg", strings.NewReader("fake image"))
	if err != nil {
		t.Fatal(err)
	}
	if resp.ErrCode != 0 {
		t.Errorf("expected errcode 0, got %d", resp.ErrCode)
	}
}
