package isv

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetExternalContact(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if got := r.URL.Query().Get("external_userid"); got != "wmXXX" {
			t.Errorf("query external_userid: %q", got)
		}
		if r.URL.Path != "/cgi-bin/externalcontact/get" {
			t.Errorf("path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errcode": 0,
			"errmsg":  "ok",
			"external_contact": map[string]any{
				"external_userid": "wmXXX",
				"name":            "张三",
				"type":            1,
			},
			"follow_user": []map[string]any{
				{
					"userid":     "zhangsan",
					"remark":     "备注",
					"createtime": 1600000000,
				},
			},
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	resp, err := cc.GetExternalContact(context.Background(), "wmXXX")
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExternalContact.Name != "张三" {
		t.Errorf("Name: %q", resp.ExternalContact.Name)
	}
	if resp.ExternalContact.Type != 1 {
		t.Errorf("Type: %d", resp.ExternalContact.Type)
	}
	if len(resp.FollowUser) != 1 {
		t.Fatalf("len(FollowUser): %d", len(resp.FollowUser))
	}
	if resp.FollowUser[0].UserID != "zhangsan" {
		t.Errorf("FollowUser[0].UserID: %q", resp.FollowUser[0].UserID)
	}
}

func TestListExternalContact(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if got := r.URL.Query().Get("userid"); got != "zhangsan" {
			t.Errorf("query userid: %q", got)
		}
		if r.URL.Path != "/cgi-bin/externalcontact/list" {
			t.Errorf("path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errcode":         0,
			"errmsg":          "ok",
			"external_userid": []string{"wmXXX", "wmYYY"},
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	resp, err := cc.ListExternalContact(context.Background(), "zhangsan")
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.ExternalUserID) != 2 {
		t.Fatalf("len(ExternalUserID): %d", len(resp.ExternalUserID))
	}
	if resp.ExternalUserID[0] != "wmXXX" {
		t.Errorf("ExternalUserID[0]: %q", resp.ExternalUserID[0])
	}
	if resp.ExternalUserID[1] != "wmYYY" {
		t.Errorf("ExternalUserID[1]: %q", resp.ExternalUserID[1])
	}
}

func TestBatchGetExternalContactByUser(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if r.URL.Path != "/cgi-bin/externalcontact/batch/get_by_user" {
			t.Errorf("path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		userIDList, ok := body["userid_list"].([]any)
		if !ok || len(userIDList) == 0 {
			t.Error("body.userid_list missing or empty")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errcode": 0,
			"errmsg":  "ok",
			"external_contact_list": []map[string]any{
				{
					"external_contact": map[string]any{
						"external_userid": "wmXXX",
						"name":            "张三",
						"type":            1,
					},
					"follow_user": []map[string]any{},
				},
			},
			"next_cursor": "CURSOR123",
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	resp, err := cc.BatchGetExternalContactByUser(context.Background(), &BatchGetExternalContactReq{
		UserIDList: []string{"zhangsan"},
		Limit:      100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.ExternalContactList) != 1 {
		t.Fatalf("len(ExternalContactList): %d", len(resp.ExternalContactList))
	}
	if resp.ExternalContactList[0].ExternalContact.Name != "张三" {
		t.Errorf("ExternalContact.Name: %q", resp.ExternalContactList[0].ExternalContact.Name)
	}
	if resp.NextCursor != "CURSOR123" {
		t.Errorf("NextCursor: %q", resp.NextCursor)
	}
}

func TestRemarkExternalContact(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if r.URL.Path != "/cgi-bin/externalcontact/remark" {
			t.Errorf("path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["userid"] != "zhangsan" {
			t.Errorf("body.userid: %v", body["userid"])
		}
		if body["external_userid"] != "wmXXX" {
			t.Errorf("body.external_userid: %v", body["external_userid"])
		}
		if body["remark"] != "新备注" {
			t.Errorf("body.remark: %v", body["remark"])
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errcode": 0,
			"errmsg":  "ok",
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	err := cc.RemarkExternalContact(context.Background(), &RemarkExternalContactReq{
		UserID:         "zhangsan",
		ExternalUserID: "wmXXX",
		Remark:         "新备注",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetCorpTagList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if r.URL.Path != "/cgi-bin/externalcontact/get_corp_tag_list" {
			t.Errorf("path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errcode": 0,
			"errmsg":  "ok",
			"tag_group": []map[string]any{
				{
					"group_id":   "GROUP001",
					"group_name": "重要客户",
					"tag": []map[string]any{
						{"id": "TAG001", "name": "VIP", "order": 1},
						{"id": "TAG002", "name": "潜力", "order": 2},
					},
				},
			},
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	resp, err := cc.GetCorpTagList(context.Background(), &GetCorpTagListReq{})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.TagGroup) != 1 {
		t.Fatalf("len(TagGroup): %d", len(resp.TagGroup))
	}
	if resp.TagGroup[0].GroupID != "GROUP001" {
		t.Errorf("GroupID: %q", resp.TagGroup[0].GroupID)
	}
	if resp.TagGroup[0].GroupName != "重要客户" {
		t.Errorf("GroupName: %q", resp.TagGroup[0].GroupName)
	}
	if len(resp.TagGroup[0].Tag) != 2 {
		t.Fatalf("len(Tag): %d", len(resp.TagGroup[0].Tag))
	}
	if resp.TagGroup[0].Tag[0].Name != "VIP" {
		t.Errorf("Tag[0].Name: %q", resp.TagGroup[0].Tag[0].Name)
	}
}

func TestAddCorpTag(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if r.URL.Path != "/cgi-bin/externalcontact/add_corp_tag" {
			t.Errorf("path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["group_name"] != "新标签组" {
			t.Errorf("body.group_name: %v", body["group_name"])
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errcode": 0,
			"errmsg":  "ok",
			"tag_group": map[string]any{
				"group_id":   "NEWGROUP001",
				"group_name": "新标签组",
				"tag": []map[string]any{
					{"id": "NEWTAG001", "name": "新标签", "order": 1},
				},
			},
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	resp, err := cc.AddCorpTag(context.Background(), &AddCorpTagReq{
		GroupName: "新标签组",
		Tag: []CorpTagInput{
			{Name: "新标签", Order: 1},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.TagGroup.GroupID != "NEWGROUP001" {
		t.Errorf("GroupID: %q", resp.TagGroup.GroupID)
	}
	if len(resp.TagGroup.Tag) != 1 {
		t.Fatalf("len(Tag): %d", len(resp.TagGroup.Tag))
	}
	if resp.TagGroup.Tag[0].Name != "新标签" {
		t.Errorf("Tag[0].Name: %q", resp.TagGroup.Tag[0].Name)
	}
}

func TestEditCorpTag(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if r.URL.Path != "/cgi-bin/externalcontact/edit_corp_tag" {
			t.Errorf("path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["id"] != "TAG001" {
			t.Errorf("body.id: %v", body["id"])
		}
		// Verify *int pointer field: order should be present with value 0.
		val, ok := body["order"]
		if !ok {
			t.Error("body.order not present")
		} else if int(val.(float64)) != 0 {
			t.Errorf("body.order: %v", val)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errcode": 0,
			"errmsg":  "ok",
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	zero := 0
	err := cc.EditCorpTag(context.Background(), &EditCorpTagReq{
		ID:    "TAG001",
		Name:  "更新标签",
		Order: &zero,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDelCorpTag(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if r.URL.Path != "/cgi-bin/externalcontact/del_corp_tag" {
			t.Errorf("path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		tagIDs, ok := body["tag_id"].([]any)
		if !ok || len(tagIDs) == 0 {
			t.Error("body.tag_id missing or empty")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errcode": 0,
			"errmsg":  "ok",
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	err := cc.DelCorpTag(context.Background(), &DelCorpTagReq{
		TagID: []string{"TAG001", "TAG002"},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestMarkTag(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if r.URL.Path != "/cgi-bin/externalcontact/mark_tag" {
			t.Errorf("path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["userid"] != "zhangsan" {
			t.Errorf("body.userid: %v", body["userid"])
		}
		if body["external_userid"] != "wmXXX" {
			t.Errorf("body.external_userid: %v", body["external_userid"])
		}
		addTags, ok := body["add_tag"].([]any)
		if !ok || len(addTags) == 0 {
			t.Error("body.add_tag missing or empty")
		}
		removeTags, ok := body["remove_tag"].([]any)
		if !ok || len(removeTags) == 0 {
			t.Error("body.remove_tag missing or empty")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errcode": 0,
			"errmsg":  "ok",
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	err := cc.MarkTag(context.Background(), &MarkTagReq{
		UserID:         "zhangsan",
		ExternalUserID: "wmXXX",
		AddTag:         []string{"TAG001"},
		RemoveTag:      []string{"TAG002"},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetFollowUserList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if r.URL.Path != "/cgi-bin/externalcontact/get_follow_user_list" {
			t.Errorf("path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errcode":     0,
			"errmsg":      "ok",
			"follow_user": []string{"zhangsan", "lisi"},
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	resp, err := cc.GetFollowUserList(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.FollowUser) != 2 {
		t.Fatalf("len(FollowUser): %d", len(resp.FollowUser))
	}
	if resp.FollowUser[0] != "zhangsan" {
		t.Errorf("FollowUser[0]: %q", resp.FollowUser[0])
	}
	if resp.FollowUser[1] != "lisi" {
		t.Errorf("FollowUser[1]: %q", resp.FollowUser[1])
	}
}
