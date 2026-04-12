package isv

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestOAuth2URL_Default(t *testing.T) {
	cfg := testConfig()
	c, err := NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	got := c.OAuth2URL("https://app.example.com/cb?x=1&y=2", "STATE1")

	// Split off the fragment before parsing.
	if !strings.HasSuffix(got, "#wechat_redirect") {
		t.Errorf("missing fragment: %q", got)
	}
	base := strings.TrimSuffix(got, "#wechat_redirect")

	u, err := url.Parse(base)
	if err != nil {
		t.Fatal(err)
	}
	if u.Host != "open.weixin.qq.com" {
		t.Errorf("host: %q", u.Host)
	}
	if u.Path != "/connect/oauth2/authorize" {
		t.Errorf("path: %q", u.Path)
	}
	q := u.Query()
	if q.Get("appid") != cfg.SuiteID {
		t.Errorf("appid: %q", q.Get("appid"))
	}
	if q.Get("redirect_uri") != "https://app.example.com/cb?x=1&y=2" {
		t.Errorf("redirect_uri: %q", q.Get("redirect_uri"))
	}
	if q.Get("response_type") != "code" {
		t.Errorf("response_type: %q", q.Get("response_type"))
	}
	if q.Get("scope") != "snsapi_privateinfo" {
		t.Errorf("scope: %q", q.Get("scope"))
	}
	if q.Get("state") != "STATE1" {
		t.Errorf("state: %q", q.Get("state"))
	}
	if q.Get("agentid") != "" {
		t.Errorf("agentid should be absent by default, got %q", q.Get("agentid"))
	}
}

func TestOAuth2URL_WithOptions(t *testing.T) {
	cfg := testConfig()
	c, err := NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	got := c.OAuth2URL(
		"https://app.example.com/cb",
		"STATE2",
		WithOAuth2Scope("snsapi_base"),
		WithOAuth2AgentID(1000001),
	)
	base := strings.TrimSuffix(got, "#wechat_redirect")
	u, err := url.Parse(base)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	if q.Get("scope") != "snsapi_base" {
		t.Errorf("scope: %q", q.Get("scope"))
	}
	if q.Get("agentid") != "1000001" {
		t.Errorf("agentid: %q", q.Get("agentid"))
	}
}

func TestGetUserInfo3rd_Member(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/service/get_provider_token":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"provider_access_token": "PTOK",
				"expires_in":            7200,
			})
		case "/cgi-bin/service/auth/getuserinfo3rd":
			if r.Method != http.MethodGet {
				t.Errorf("method: %s", r.Method)
			}
			if got := r.URL.Query().Get("provider_access_token"); got != "PTOK" {
				t.Errorf("token query: %q", got)
			}
			if got := r.URL.Query().Get("code"); got != "AUTH1" {
				t.Errorf("code query: %q", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"CorpId":      "wxcorp1",
				"UserId":      "u1",
				"DeviceId":    "dev1",
				"user_ticket": "TICKET1",
				"expires_in":  1800,
				"open_userid": "ou1",
			})
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := newTestISVClientWithProvider(t, srv.URL)
	resp, err := c.GetUserInfo3rd(context.Background(), "AUTH1")
	if err != nil {
		t.Fatal(err)
	}
	if resp.CorpID != "wxcorp1" || resp.UserID != "u1" || resp.DeviceID != "dev1" {
		t.Errorf("member fields: %+v", resp)
	}
	if resp.UserTicket != "TICKET1" || resp.ExpiresIn != 1800 {
		t.Errorf("ticket: %+v", resp)
	}
	if resp.OpenUserID != "ou1" {
		t.Errorf("open_userid: %q", resp.OpenUserID)
	}
	if resp.OpenID != "" {
		t.Errorf("OpenID should be empty for member, got %q", resp.OpenID)
	}
}

func TestGetUserInfo3rd_NonMember(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/service/get_provider_token":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"provider_access_token": "PTOK",
				"expires_in":            7200,
			})
		case "/cgi-bin/service/auth/getuserinfo3rd":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"CorpId": "wxcorp1",
				"OpenId": "oAbCdEf",
			})
		}
	}))
	defer srv.Close()

	c := newTestISVClientWithProvider(t, srv.URL)
	resp, err := c.GetUserInfo3rd(context.Background(), "AUTH2")
	if err != nil {
		t.Fatal(err)
	}
	if resp.UserID != "" {
		t.Errorf("UserID should be empty for non-member, got %q", resp.UserID)
	}
	if resp.OpenID != "oAbCdEf" {
		t.Errorf("OpenID: %q", resp.OpenID)
	}
}

func TestGetUserDetail3rd(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/service/get_provider_token":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"provider_access_token": "PTOK",
				"expires_in":            7200,
			})
		case "/cgi-bin/service/auth/getuserdetail3rd":
			if r.Method != http.MethodPost {
				t.Errorf("method: %s", r.Method)
			}
			if got := r.URL.Query().Get("provider_access_token"); got != "PTOK" {
				t.Errorf("token query: %q", got)
			}
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["user_ticket"] != "TICKET1" {
				t.Errorf("body: %+v", body)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"corpid":   "wxcorp1",
				"userid":   "u1",
				"gender":   "1",
				"avatar":   "http://img/a.png",
				"mobile":   "13800000000",
				"email":    "u1@example.com",
				"biz_mail": "u1@biz.example.com",
				"address":  "Beijing",
			})
		}
	}))
	defer srv.Close()

	c := newTestISVClientWithProvider(t, srv.URL)
	resp, err := c.GetUserDetail3rd(context.Background(), "TICKET1")
	if err != nil {
		t.Fatal(err)
	}
	if resp.CorpID != "wxcorp1" || resp.UserID != "u1" {
		t.Errorf("ids: %+v", resp)
	}
	if resp.Mobile != "13800000000" || resp.Email != "u1@example.com" || resp.BizMail != "u1@biz.example.com" {
		t.Errorf("contact: %+v", resp)
	}
	if resp.Gender != "1" || resp.Avatar != "http://img/a.png" || resp.Address != "Beijing" {
		t.Errorf("profile: %+v", resp)
	}
}
