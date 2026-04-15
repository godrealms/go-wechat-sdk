package offiaccount

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

func newMsgTestClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	h := utils.NewHTTP(srv.URL, utils.WithTimeout(3*time.Second))
	return NewClient(context.Background(), &Config{AppId: "test", AppSecret: "secret"},
		WithHTTPClient(h),
		WithTokenSource(fixedToken{"FAKE_TOKEN"}),
	)
}

func msgJsonServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
}

func TestSendTemplateMessage(t *testing.T) {
	tests := []struct {
		name    string
		resp    string
		wantErr bool
	}{
		{"success", `{"errcode":0,"errmsg":"ok","msgid":123456}`, false},
		{"errcode error", `{"errcode":40003,"errmsg":"invalid openid"}`, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := msgJsonServer(t, 200, tc.resp)
			defer srv.Close()
			c := newMsgTestClient(t, srv)
			req := &SubscribeMessageRequest{
				ToUser:     "oUser123",
				TemplateID: "tplId",
			}
			result, err := c.SendTemplateMessage(context.Background(), req)
			if tc.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result == nil {
				t.Error("expected non-nil result")
			}
		})
	}
}

func TestSendTemplateMessage_NetworkError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
	}))
	srv.Close()
	c := newMsgTestClient(t, srv)
	_, err := c.SendTemplateMessage(context.Background(), &SubscribeMessageRequest{})
	if err == nil {
		t.Error("expected network error")
	}
}

func TestAddTemplate(t *testing.T) {
	srv := msgJsonServer(t, 200, `{"errcode":0,"errmsg":"ok","template_id":"tpl123"}`)
	defer srv.Close()
	c := newMsgTestClient(t, srv)
	result, err := c.AddTemplate(context.Background(), "shortId", []string{"keyword1", "keyword2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestAddTemplate_ErrCode(t *testing.T) {
	srv := msgJsonServer(t, 200, `{"errcode":40015,"errmsg":"invalid template id"}`)
	defer srv.Close()
	c := newMsgTestClient(t, srv)
	_, err := c.AddTemplate(context.Background(), "bad", nil)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestDeleteTemplate(t *testing.T) {
	srv := msgJsonServer(t, 200, `{"errcode":0,"errmsg":"ok"}`)
	defer srv.Close()
	c := newMsgTestClient(t, srv)
	_, err := c.DeleteTemplate(context.Background(), "tplId")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetAllTemplates(t *testing.T) {
	srv := msgJsonServer(t, 200, `{"errcode":0,"errmsg":"ok","template_list":[]}`)
	defer srv.Close()
	c := newMsgTestClient(t, srv)
	result, err := c.GetAllTemplates(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestSetIndustry(t *testing.T) {
	srv := msgJsonServer(t, 200, `{"errcode":0,"errmsg":"ok"}`)
	defer srv.Close()
	c := newMsgTestClient(t, srv)
	if err := c.SetIndustry(context.Background(), "1", "2"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTemplateSubscribe(t *testing.T) {
	tests := []struct {
		name    string
		resp    string
		wantErr bool
	}{
		{"success", `{"errcode":0,"errmsg":"ok"}`, false},
		{"errcode error", `{"errcode":43101,"errmsg":"user refuse to accept the msg"}`, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := msgJsonServer(t, 200, tc.resp)
			defer srv.Close()
			c := newMsgTestClient(t, srv)
			err := c.TemplateSubscribe(context.Background(), &TemplateSubscribeReq{})
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestDeleteMassMsg(t *testing.T) {
	srv := msgJsonServer(t, 200, `{"errcode":0,"errmsg":"ok"}`)
	defer srv.Close()
	c := newMsgTestClient(t, srv)
	if err := c.DeleteMassMsg(context.Background(), &DeleteMassMsgRequest{MsgId: 1234567}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMassSend(t *testing.T) {
	srv := msgJsonServer(t, 200, `{"errcode":0,"errmsg":"send job submission success","msg_id":34182}`)
	defer srv.Close()
	c := newMsgTestClient(t, srv)
	result, err := c.MassSend(context.Background(), &MassSendRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestPreview(t *testing.T) {
	srv := msgJsonServer(t, 200, `{"errcode":0,"errmsg":"ok"}`)
	defer srv.Close()
	c := newMsgTestClient(t, srv)
	result, err := c.Preview(context.Background(), &MassSendRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestSendAll(t *testing.T) {
	srv := msgJsonServer(t, 200, `{"errcode":0,"errmsg":"send job submission success","msg_id":34183}`)
	defer srv.Close()
	c := newMsgTestClient(t, srv)
	result, err := c.SendAll(context.Background(), &MassSendByTagRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestSetSpeed(t *testing.T) {
	srv := msgJsonServer(t, 200, `{"errcode":0,"errmsg":"ok"}`)
	defer srv.Close()
	c := newMsgTestClient(t, srv)
	result, err := c.SetSpeed(context.Background(), 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}
