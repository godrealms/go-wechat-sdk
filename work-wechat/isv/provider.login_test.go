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
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
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
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"usertype": 1,
				"user_info": map[string]interface{}{
					"userid":      "admin1",
					"open_userid": "oadmin1",
					"name":        "Admin",
					"avatar":      "http://img/a.png",
				},
				"corp_info": map[string]interface{}{"corpid": "wxcorp1"},
				"agent": []map[string]interface{}{
					{"agentid": 1000001, "auth_type": 1},
				},
				"auth_info": map[string]interface{}{
					"department": []map[string]interface{}{
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
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"provider_access_token": "PTOK",
				"expires_in":            7200,
			})
		case "/cgi-bin/service/get_login_info":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"usertype": 2,
				"user_info": map[string]interface{}{
					"userid":      "u1",
					"open_userid": "ou1",
					"name":        "Alice",
				},
				"corp_info": map[string]interface{}{"corpid": "wxcorp1"},
				"agent": []map[string]interface{}{
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
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"provider_access_token": "PTOK",
				"expires_in":            7200,
			})
		case "/cgi-bin/service/get_login_info":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
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
