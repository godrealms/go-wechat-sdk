package isv

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetPermanentCode_StoresAuthorizer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token":   "CORP_TOK",
			"expires_in":     7200,
			"permanent_code": "PERM",
			"auth_corp_info": map[string]interface{}{
				"corpid":    "wxcorp1",
				"corp_name": "ACME",
			},
			"auth_info": map[string]interface{}{
				"agent": []map[string]interface{}{
					{"agentid": 1000001, "name": "HR"},
				},
			},
			"auth_user_info": map[string]interface{}{
				"userid": "admin",
				"name":   "Admin",
			},
		})
	}))
	defer srv.Close()

	c := newTestISVClient(t, srv.URL)
	_ = c.store.PutSuiteToken(context.Background(), "suite1", "STOK", time.Now().Add(time.Hour))

	resp, err := c.GetPermanentCode(context.Background(), "auth_code_xyz")
	if err != nil {
		t.Fatal(err)
	}
	if resp.PermanentCode != "PERM" || resp.AuthCorpInfo.CorpID != "wxcorp1" {
		t.Fatalf("resp: %+v", resp)
	}

	// Verify AuthorizerTokens written to store
	got, err := c.store.GetAuthorizer(context.Background(), "suite1", "wxcorp1")
	if err != nil {
		t.Fatal(err)
	}
	if got.PermanentCode != "PERM" || got.CorpAccessToken != "CORP_TOK" {
		t.Fatalf("stored: %+v", got)
	}
}

func TestGetAuthInfo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["auth_corpid"] != "wxcorp1" || body["permanent_code"] != "PERM" {
			t.Errorf("body: %+v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"auth_corp_info": map[string]interface{}{
				"corpid":    "wxcorp1",
				"corp_name": "ACME",
			},
			"auth_info": map[string]interface{}{
				"agent": []map[string]interface{}{{"agentid": 1}},
			},
		})
	}))
	defer srv.Close()

	c := newTestISVClient(t, srv.URL)
	_ = c.store.PutSuiteToken(context.Background(), "suite1", "STOK", time.Now().Add(time.Hour))

	resp, err := c.GetAuthInfo(context.Background(), "wxcorp1", "PERM")
	if err != nil || resp.AuthCorpInfo.CorpID != "wxcorp1" {
		t.Fatalf("got %+v err=%v", resp, err)
	}
}

func TestGetAdminList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["auth_corpid"] != "wxcorp1" {
			t.Errorf("body: %+v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"admin": []map[string]interface{}{
				{"userid": "u1", "open_userid": "o1", "auth_type": 1},
			},
		})
	}))
	defer srv.Close()

	c := newTestISVClient(t, srv.URL)
	_ = c.store.PutSuiteToken(context.Background(), "suite1", "STOK", time.Now().Add(time.Hour))

	resp, err := c.GetAdminList(context.Background(), "wxcorp1", "1000001")
	if err != nil || len(resp.Admin) != 1 || resp.Admin[0].AuthType != 1 {
		t.Fatalf("got %+v err=%v", resp, err)
	}
}
