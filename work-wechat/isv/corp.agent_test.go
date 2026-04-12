package isv

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetAgent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if got := r.URL.Query().Get("agentid"); got != "1000002" {
			t.Errorf("query agentid: %q", got)
		}
		if r.URL.Path != "/cgi-bin/agent/get" {
			t.Errorf("path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"errcode":              0,
			"errmsg":               "ok",
			"agentid":              1000002,
			"name":                 "TestApp",
			"description":          "A test application",
			"square_logo_url":      "https://example.com/logo.png",
			"home_url":             "https://example.com/home",
			"redirect_domain":      "example.com",
			"isreportenter":        1,
			"report_location_flag": 0,
			"allow_userinfos": map[string]interface{}{
				"user": []map[string]interface{}{
					{"userid": "zhangsan"},
				},
			},
			"allow_partys": map[string]interface{}{
				"partyid": []int{1, 2},
			},
			"allow_tags": map[string]interface{}{
				"tagid": []int{10},
			},
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	agent, err := cc.GetAgent(context.Background(), 1000002)
	if err != nil {
		t.Fatal(err)
	}
	if agent.AgentID != 1000002 {
		t.Errorf("AgentID: %d", agent.AgentID)
	}
	if agent.Name != "TestApp" {
		t.Errorf("Name: %q", agent.Name)
	}
	if agent.Description != "A test application" {
		t.Errorf("Description: %q", agent.Description)
	}
	if agent.HomeURL != "https://example.com/home" {
		t.Errorf("HomeURL: %q", agent.HomeURL)
	}
	if agent.IsReportEnter != 1 {
		t.Errorf("IsReportEnter: %d", agent.IsReportEnter)
	}
	if len(agent.AllowUserInfos.User) != 1 {
		t.Fatalf("len(AllowUserInfos.User): %d", len(agent.AllowUserInfos.User))
	}
	if agent.AllowUserInfos.User[0].UserID != "zhangsan" {
		t.Errorf("AllowUserInfos.User[0].UserID: %q", agent.AllowUserInfos.User[0].UserID)
	}
	if len(agent.AllowParties.PartyID) != 2 {
		t.Fatalf("len(AllowParties.PartyID): %d", len(agent.AllowParties.PartyID))
	}
	if agent.AllowParties.PartyID[0] != 1 || agent.AllowParties.PartyID[1] != 2 {
		t.Errorf("AllowParties.PartyID: %v", agent.AllowParties.PartyID)
	}
	if len(agent.AllowTags.TagID) != 1 || agent.AllowTags.TagID[0] != 10 {
		t.Errorf("AllowTags.TagID: %v", agent.AllowTags.TagID)
	}
}

func TestSetAgent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if r.URL.Path != "/cgi-bin/agent/set" {
			t.Errorf("path: %s", r.URL.Path)
		}
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if int(body["agentid"].(float64)) != 1000002 {
			t.Errorf("body.agentid: %v", body["agentid"])
		}
		if body["name"] != "UpdatedApp" {
			t.Errorf("body.name: %v", body["name"])
		}
		if body["home_url"] != "https://example.com/new" {
			t.Errorf("body.home_url: %v", body["home_url"])
		}
		// Verify *int pointer field: isreportenter should be present with value 0.
		val, ok := body["isreportenter"]
		if !ok {
			t.Error("body.isreportenter not present")
		} else if int(val.(float64)) != 0 {
			t.Errorf("body.isreportenter: %v", val)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"errcode": 0,
			"errmsg":  "ok",
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	zero := 0
	err := cc.SetAgent(context.Background(), &SetAgentReq{
		AgentID:       1000002,
		Name:          "UpdatedApp",
		HomeURL:       "https://example.com/new",
		IsReportEnter: &zero,
	})
	if err != nil {
		t.Fatal(err)
	}
}
