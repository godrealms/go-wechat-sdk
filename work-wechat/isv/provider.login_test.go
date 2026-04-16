package isv

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetLoginInfo_Admin(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/service/get_provider_token":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"provider_access_token": "PTOK",
				"expires_in":            7200,
			})
		case "/cgi-bin/service/get_login_info":
			if got := r.URL.Query().Get("provider_access_token"); got != "PTOK" {
				t.Errorf("token query: %q", got)
			}
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["auth_code"] != "AUTH1" {
				t.Errorf("auth_code body: %+v", body)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"usertype": 1,
				"user_info": map[string]any{
					"userid":      "admin1",
					"open_userid": "oadmin1",
					"name":        "Admin",
					"avatar":      "http://img/a.png",
				},
				"corp_info": map[string]any{"corpid": "wxcorp1"},
				"agent": []map[string]any{
					{"agentid": 1000001, "auth_type": 1},
				},
				"auth_info": map[string]any{
					"department": []map[string]any{
						{"id": 1, "writable": true},
						{"id": 2, "writable": false},
					},
				},
			})
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := newTestISVClientWithProvider(t, srv.URL)
	resp, err := c.GetLoginInfo(context.Background(), "AUTH1")
	if err != nil {
		t.Fatal(err)
	}
	if resp.UserType != 1 {
		t.Errorf("usertype: %d", resp.UserType)
	}
	if resp.UserInfo.UserID != "admin1" || resp.UserInfo.OpenUserID != "oadmin1" {
		t.Errorf("user_info: %+v", resp.UserInfo)
	}
	if resp.CorpInfo.CorpID != "wxcorp1" {
		t.Errorf("corp_info: %+v", resp.CorpInfo)
	}
	if len(resp.Agent) != 1 || resp.Agent[0].AgentID != 1000001 || resp.Agent[0].AuthType != 1 {
		t.Errorf("agent: %+v", resp.Agent)
	}
	if len(resp.AuthInfo.Department) != 2 || !resp.AuthInfo.Department[0].Writable {
		t.Errorf("auth_info: %+v", resp.AuthInfo)
	}
}

func TestGetLoginInfo_Member(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/service/get_provider_token":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"provider_access_token": "PTOK",
				"expires_in":            7200,
			})
		case "/cgi-bin/service/get_login_info":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"usertype": 2,
				"user_info": map[string]any{
					"userid":      "u1",
					"open_userid": "ou1",
					"name":        "Alice",
				},
				"corp_info": map[string]any{"corpid": "wxcorp1"},
				"agent": []map[string]any{
					{"agentid": 1000001, "auth_type": 0},
				},
			})
		}
	}))
	defer srv.Close()

	c := newTestISVClientWithProvider(t, srv.URL)
	resp, err := c.GetLoginInfo(context.Background(), "AUTH2")
	if err != nil {
		t.Fatal(err)
	}
	if resp.UserType != 2 {
		t.Errorf("usertype: %d", resp.UserType)
	}
	if len(resp.AuthInfo.Department) != 0 {
		t.Errorf("department should be empty for member, got: %+v", resp.AuthInfo.Department)
	}
}

func TestGetLoginInfo_WeixinError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/service/get_provider_token":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"provider_access_token": "PTOK",
				"expires_in":            7200,
			})
		case "/cgi-bin/service/get_login_info":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"errcode": 40029,
				"errmsg":  "invalid code",
			})
		}
	}))
	defer srv.Close()

	c := newTestISVClientWithProvider(t, srv.URL)
	_, err := c.GetLoginInfo(context.Background(), "BAD")
	if err == nil {
		t.Fatal("want error, got nil")
	}
	var we *WeixinError
	if !errors.As(err, &we) || we.ErrCode != 40029 {
		t.Errorf("want *WeixinError errcode=40029, got %v", err)
	}
}

func TestGetRegisterCode_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/service/get_provider_token":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"provider_access_token": "PTOK",
				"expires_in":            7200,
			})
		case "/cgi-bin/service/get_register_code":
			if got := r.URL.Query().Get("provider_access_token"); got != "PTOK" {
				t.Errorf("token query: %q", got)
			}
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["template_id"] != "tmpl1" || body["corp_name"] != "ACME" || body["state"] != "s1" {
				t.Errorf("body: %+v", body)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"register_code": "REG123",
				"expires_in":    604800,
			})
		}
	}))
	defer srv.Close()

	c := newTestISVClientWithProvider(t, srv.URL)
	resp, err := c.GetRegisterCode(context.Background(), &GetRegisterCodeReq{
		TemplateID: "tmpl1",
		CorpName:   "ACME",
		State:      "s1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.RegisterCode != "REG123" || resp.ExpiresIn != 604800 {
		t.Errorf("resp: %+v", resp)
	}
}

func TestGetRegisterCode_NilRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/service/get_provider_token":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"provider_access_token": "PTOK",
				"expires_in":            7200,
			})
		case "/cgi-bin/service/get_register_code":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"register_code": "REG_EMPTY",
				"expires_in":    3600,
			})
		}
	}))
	defer srv.Close()

	c := newTestISVClientWithProvider(t, srv.URL)
	resp, err := c.GetRegisterCode(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.RegisterCode != "REG_EMPTY" {
		t.Errorf("resp: %+v", resp)
	}
}

func TestGetRegistrationInfo_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/service/get_provider_token":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"provider_access_token": "PTOK",
				"expires_in":            7200,
			})
		case "/cgi-bin/service/get_registration_info":
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["register_code"] != "REG123" {
				t.Errorf("body: %+v", body)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"corp_info": map[string]any{
					"corpid":        "wxcorp1",
					"corp_name":     "ACME",
					"corp_user_max": 200,
					"subject_type":  1,
					"corp_industry": "Tech",
				},
				"auth_user_info": map[string]any{
					"userid": "admin1",
					"name":   "Root",
				},
				"contact_sync": map[string]any{
					"access_token": "CTOK",
					"expires_in":   7200,
				},
				"auth_info": map[string]any{
					"agent": []map[string]any{
						{"agentid": 1000001, "name": "HR"},
					},
				},
				"permanent_code": "PERM_REG",
			})
		}
	}))
	defer srv.Close()

	c := newTestISVClientWithProvider(t, srv.URL)
	resp, err := c.GetRegistrationInfo(context.Background(), "REG123")
	if err != nil {
		t.Fatal(err)
	}
	if resp.CorpInfo.CorpID != "wxcorp1" || resp.CorpInfo.CorpName != "ACME" {
		t.Errorf("corp_info: %+v", resp.CorpInfo)
	}
	if resp.AuthUserInfo.UserID != "admin1" {
		t.Errorf("auth_user_info: %+v", resp.AuthUserInfo)
	}
	if resp.ContactSync.AccessToken != "CTOK" {
		t.Errorf("contact_sync: %+v", resp.ContactSync)
	}
	if len(resp.AuthInfo.Agent) != 1 || resp.AuthInfo.Agent[0].AgentID != 1000001 {
		t.Errorf("auth_info: %+v", resp.AuthInfo)
	}
	if resp.PermanentCode != "PERM_REG" {
		t.Errorf("permanent_code: %q", resp.PermanentCode)
	}
}
