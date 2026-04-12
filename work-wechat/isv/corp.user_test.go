package isv

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateUser(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if r.URL.Path != "/cgi-bin/user/create" {
			t.Errorf("path: %s", r.URL.Path)
		}
		var body CreateUserReq
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.UserID != "zhangsan" {
			t.Errorf("body.UserID: %q", body.UserID)
		}
		if body.Name != "Zhang San" {
			t.Errorf("body.Name: %q", body.Name)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"errcode": 0,
			"errmsg":  "ok",
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	err := cc.CreateUser(context.Background(), &CreateUserReq{
		UserID:     "zhangsan",
		Name:       "Zhang San",
		Department: []int{1},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetUser(t *testing.T) {
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
		if r.URL.Path != "/cgi-bin/user/get" {
			t.Errorf("path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"errcode":    0,
			"errmsg":     "ok",
			"userid":     "zhangsan",
			"name":       "Zhang San",
			"department": []int{1, 2},
			"status":     1,
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	u, err := cc.GetUser(context.Background(), "zhangsan")
	if err != nil {
		t.Fatal(err)
	}
	if u.UserID != "zhangsan" {
		t.Errorf("UserID: %q", u.UserID)
	}
	if u.Name != "Zhang San" {
		t.Errorf("Name: %q", u.Name)
	}
	if u.Status != 1 {
		t.Errorf("Status: %d", u.Status)
	}
}

func TestListUserSimple(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if got := r.URL.Query().Get("department_id"); got != "1" {
			t.Errorf("query department_id: %q", got)
		}
		if got := r.URL.Query().Get("fetch_child"); got != "1" {
			t.Errorf("query fetch_child: %q", got)
		}
		if r.URL.Path != "/cgi-bin/user/simplelist" {
			t.Errorf("path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"errcode": 0,
			"errmsg":  "ok",
			"userlist": []map[string]interface{}{
				{"userid": "zhangsan", "name": "Zhang San", "department": []int{1}},
			},
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	resp, err := cc.ListUserSimple(context.Background(), 1, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.UserList) != 1 {
		t.Fatalf("len(UserList): %d", len(resp.UserList))
	}
	if resp.UserList[0].UserID != "zhangsan" {
		t.Errorf("UserList[0].UserID: %q", resp.UserList[0].UserID)
	}
}

func TestListUserDetail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if got := r.URL.Query().Get("department_id"); got != "2" {
			t.Errorf("query department_id: %q", got)
		}
		if got := r.URL.Query().Get("fetch_child"); got != "0" {
			t.Errorf("query fetch_child: %q", got)
		}
		if r.URL.Path != "/cgi-bin/user/list" {
			t.Errorf("path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"errcode": 0,
			"errmsg":  "ok",
			"userlist": []map[string]interface{}{
				{"userid": "lisi", "name": "Li Si", "department": []int{2}, "status": 1},
			},
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	resp, err := cc.ListUserDetail(context.Background(), 2, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.UserList) != 1 {
		t.Fatalf("len(UserList): %d", len(resp.UserList))
	}
	if resp.UserList[0].UserID != "lisi" {
		t.Errorf("UserList[0].UserID: %q", resp.UserList[0].UserID)
	}
	if resp.UserList[0].Status != 1 {
		t.Errorf("UserList[0].Status: %d", resp.UserList[0].Status)
	}
}
