package isv

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateTag(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if r.URL.Path != "/cgi-bin/tag/create" {
			t.Errorf("path: %s", r.URL.Path)
		}
		var body CreateTagReq
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.TagName != "DevTeam" {
			t.Errorf("body.TagName: %q", body.TagName)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"errcode": 0,
			"errmsg":  "ok",
			"tagid":   7,
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	resp, err := cc.CreateTag(context.Background(), &CreateTagReq{
		TagName: "DevTeam",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.TagID != 7 {
		t.Errorf("resp.TagID: %d", resp.TagID)
	}
}

func TestGetTagUsers(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if got := r.URL.Query().Get("tagid"); got != "7" {
			t.Errorf("query tagid: %q", got)
		}
		if r.URL.Path != "/cgi-bin/tag/get" {
			t.Errorf("path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"errcode": 0,
			"errmsg":  "ok",
			"tagname": "DevTeam",
			"userlist": []map[string]interface{}{
				{"userid": "zhangsan", "name": "Zhang San"},
			},
			"partylist": []int{1, 2},
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	resp, err := cc.GetTagUsers(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	if resp.TagName != "DevTeam" {
		t.Errorf("TagName: %q", resp.TagName)
	}
	if len(resp.UserList) != 1 {
		t.Fatalf("len(UserList): %d", len(resp.UserList))
	}
	if resp.UserList[0].UserID != "zhangsan" {
		t.Errorf("UserList[0].UserID: %q", resp.UserList[0].UserID)
	}
	if len(resp.PartyList) != 2 {
		t.Errorf("len(PartyList): %d", len(resp.PartyList))
	}
}

func TestAddTagUsers(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if r.URL.Path != "/cgi-bin/tag/addtagusers" {
			t.Errorf("path: %s", r.URL.Path)
		}
		var body TagUsersModifyReq
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.TagID != 7 {
			t.Errorf("body.TagID: %d", body.TagID)
		}
		if len(body.UserList) != 2 {
			t.Errorf("len(body.UserList): %d", len(body.UserList))
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"errcode":      0,
			"errmsg":       "ok",
			"invalidlist":  "lisi",
			"invalidparty": []int{3},
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	resp, err := cc.AddTagUsers(context.Background(), &TagUsersModifyReq{
		TagID:    7,
		UserList: []string{"zhangsan", "lisi"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.InvalidList != "lisi" {
		t.Errorf("InvalidList: %q", resp.InvalidList)
	}
	if len(resp.InvalidParty) != 1 || resp.InvalidParty[0] != 3 {
		t.Errorf("InvalidParty: %v", resp.InvalidParty)
	}
}

func TestUpdateTag(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if r.URL.Path != "/cgi-bin/tag/update" {
			t.Errorf("path: %s", r.URL.Path)
		}
		var body UpdateTagReq
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.TagID != 7 {
			t.Errorf("body.TagID: %d", body.TagID)
		}
		if body.TagName != "QA Team" {
			t.Errorf("body.TagName: %q", body.TagName)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"errcode": 0,
			"errmsg":  "ok",
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	err := cc.UpdateTag(context.Background(), &UpdateTagReq{
		TagID:   7,
		TagName: "QA Team",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDeleteTag(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if got := r.URL.Query().Get("tagid"); got != "7" {
			t.Errorf("query tagid: %q", got)
		}
		if r.URL.Path != "/cgi-bin/tag/delete" {
			t.Errorf("path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"errcode": 0,
			"errmsg":  "ok",
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	err := cc.DeleteTag(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
}

func TestListTag(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if r.URL.Path != "/cgi-bin/tag/list" {
			t.Errorf("path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"errcode": 0,
			"errmsg":  "ok",
			"taglist": []map[string]interface{}{
				{"tagid": 1, "tagname": "DevTeam"},
				{"tagid": 2, "tagname": "QA"},
			},
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	tags, err := cc.ListTag(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) != 2 {
		t.Fatalf("len(tags): %d", len(tags))
	}
	if tags[0].TagName != "DevTeam" {
		t.Errorf("tags[0].TagName: %q", tags[0].TagName)
	}
}

func TestDelTagUsers(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if r.URL.Path != "/cgi-bin/tag/deltagusers" {
			t.Errorf("path: %s", r.URL.Path)
		}
		var body TagUsersModifyReq
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.TagID != 7 {
			t.Errorf("body.TagID: %d", body.TagID)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"errcode":      0,
			"errmsg":       "ok",
			"invalidlist":  "",
			"invalidparty": []int{},
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	resp, err := cc.DelTagUsers(context.Background(), &TagUsersModifyReq{
		TagID:    7,
		UserList: []string{"zhangsan"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.InvalidList != "" {
		t.Errorf("InvalidList: %q", resp.InvalidList)
	}
}
