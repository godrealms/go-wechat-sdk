package isv

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetWorkbenchTemplate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/agent/set_workbench_template" {
			return
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if int(body["agentid"].(float64)) != 1000001 {
			t.Errorf("agentid: %v", body["agentid"])
		}
		if body["type"] != "key_data" {
			t.Errorf("type: %v", body["type"])
		}
		kd := body["key_data"].(map[string]any)
		items := kd["items"].([]any)
		if len(items) != 2 {
			t.Errorf("items count: %d", len(items))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"errcode": 0, "errmsg": "ok"})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	err := cc.SetWorkbenchTemplate(context.Background(), &WorkbenchTemplateReq{
		AgentID: 1000001,
		Type:    "key_data",
		KeyData: &WBKeyData{
			Items: []WBKeyDataItem{
				{Key: "待审批", Data: "2", JumpURL: "https://example.com/1"},
				{Key: "已通过", Data: "100"},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetWorkbenchTemplate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/agent/get_workbench_template" {
			return
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if int(body["agentid"].(float64)) != 1000001 {
			t.Errorf("agentid: %v", body["agentid"])
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"agentid": 1000001,
			"type":    "key_data",
			"key_data": map[string]any{
				"items": []map[string]any{
					{"key": "待审批", "data": "2", "jump_url": "https://example.com/1"},
				},
			},
			"replace_user_data": true,
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	resp, err := cc.GetWorkbenchTemplate(context.Background(), 1000001)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Type != "key_data" || resp.AgentID != 1000001 {
		t.Errorf("resp: %+v", resp)
	}
	if resp.KeyData == nil || len(resp.KeyData.Items) != 1 || resp.KeyData.Items[0].Key != "待审批" {
		t.Errorf("key_data: %+v", resp.KeyData)
	}
	if !resp.ReplaceUserData {
		t.Errorf("replace_user_data: %v", resp.ReplaceUserData)
	}
}

func TestSetWorkbenchData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/agent/set_workbench_data" {
			return
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["userid"] != "u1" {
			t.Errorf("userid: %v", body["userid"])
		}
		if body["type"] != "key_data" {
			t.Errorf("type: %v", body["type"])
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"errcode": 0, "errmsg": "ok"})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	err := cc.SetWorkbenchData(context.Background(), &WorkbenchDataReq{
		AgentID: 1000001,
		UserID:  "u1",
		Type:    "key_data",
		KeyData: &WBKeyData{
			Items: []WBKeyDataItem{
				{Key: "待审批", Data: "5"},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}
