package isv

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateDepartment(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if r.URL.Path != "/cgi-bin/department/create" {
			t.Errorf("path: %s", r.URL.Path)
		}
		var body CreateDeptReq
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Name != "Tech" {
			t.Errorf("body.Name: %q", body.Name)
		}
		if body.ParentID != 1 {
			t.Errorf("body.ParentID: %d", body.ParentID)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"errcode": 0,
			"errmsg":  "ok",
			"id":      42,
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	resp, err := cc.CreateDepartment(context.Background(), &CreateDeptReq{
		Name:     "Tech",
		ParentID: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.ID != 42 {
		t.Errorf("resp.ID: %d", resp.ID)
	}
}

func TestListDepartment(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if got := r.URL.Query().Get("id"); got != "1" {
			t.Errorf("query id: %q", got)
		}
		if r.URL.Path != "/cgi-bin/department/list" {
			t.Errorf("path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"errcode": 0,
			"errmsg":  "ok",
			"department": []map[string]interface{}{
				{"id": 1, "name": "Root", "parentid": 0},
				{"id": 2, "name": "Tech", "parentid": 1},
			},
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	depts, err := cc.ListDepartment(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(depts) != 2 {
		t.Fatalf("len(depts): %d", len(depts))
	}
	if depts[0].Name != "Root" {
		t.Errorf("depts[0].Name: %q", depts[0].Name)
	}
	if depts[1].Name != "Tech" {
		t.Errorf("depts[1].Name: %q", depts[1].Name)
	}
}

func TestDeleteDepartment(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if got := r.URL.Query().Get("id"); got != "99" {
			t.Errorf("query id: %q", got)
		}
		if r.URL.Path != "/cgi-bin/department/delete" {
			t.Errorf("path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"errcode": 0,
			"errmsg":  "ok",
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	err := cc.DeleteDepartment(context.Background(), 99)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUpdateDepartment(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if r.URL.Path != "/cgi-bin/department/update" {
			t.Errorf("path: %s", r.URL.Path)
		}
		var body UpdateDeptReq
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.ID != 2 {
			t.Errorf("body.ID: %d", body.ID)
		}
		if body.Name != "Engineering" {
			t.Errorf("body.Name: %q", body.Name)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"errcode": 0,
			"errmsg":  "ok",
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	err := cc.UpdateDepartment(context.Background(), &UpdateDeptReq{
		ID:   2,
		Name: "Engineering",
	})
	if err != nil {
		t.Fatal(err)
	}
}
