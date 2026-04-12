package isv

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateMenu(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		if r.URL.Path != "/cgi-bin/menu/create" {
			t.Errorf("path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if got := r.URL.Query().Get("agentid"); got != "1000002" {
			t.Errorf("agentid: %q", got)
		}

		var body CreateMenuReq
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if len(body.Button) != 2 {
			t.Fatalf("len(button): %d", len(body.Button))
		}
		// First button: click type with key
		if body.Button[0].Type != "click" {
			t.Errorf("button[0].type: %q", body.Button[0].Type)
		}
		if body.Button[0].Key != "V1001_TODAY_MUSIC" {
			t.Errorf("button[0].key: %q", body.Button[0].Key)
		}
		// Second button: parent with sub_button
		if len(body.Button[1].SubButton) != 2 {
			t.Fatalf("len(button[1].sub_button): %d", len(body.Button[1].SubButton))
		}
		if body.Button[1].SubButton[0].Type != "view" {
			t.Errorf("button[1].sub_button[0].type: %q", body.Button[1].SubButton[0].Type)
		}

		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"errcode": 0,
			"errmsg":  "ok",
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	err := cc.CreateMenu(context.Background(), 1000002, &CreateMenuReq{
		Button: []MenuButton{
			{
				Type: "click",
				Name: "Today Music",
				Key:  "V1001_TODAY_MUSIC",
			},
			{
				Name: "Menu",
				SubButton: []MenuButton{
					{Type: "view", Name: "Search", URL: "https://www.example.com"},
					{Type: "click", Name: "Like", Key: "V1001_LIKE"},
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetMenu(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method: %s", r.Method)
		}
		if r.URL.Path != "/cgi-bin/menu/get" {
			t.Errorf("path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if got := r.URL.Query().Get("agentid"); got != "1000002" {
			t.Errorf("agentid: %q", got)
		}

		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"errcode": 0,
			"errmsg":  "ok",
			"button": []map[string]interface{}{
				{
					"type": "click",
					"name": "Today Music",
					"key":  "V1001_TODAY_MUSIC",
				},
				{
					"name": "Menu",
					"sub_button": []map[string]interface{}{
						{"type": "view", "name": "Search", "url": "https://www.example.com"},
						{"type": "click", "name": "Like", "key": "V1001_LIKE"},
					},
				},
			},
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	resp, err := cc.GetMenu(context.Background(), 1000002)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Button) != 2 {
		t.Fatalf("len(button): %d", len(resp.Button))
	}
	if resp.Button[0].Type != "click" {
		t.Errorf("button[0].type: %q", resp.Button[0].Type)
	}
	if resp.Button[0].Name != "Today Music" {
		t.Errorf("button[0].name: %q", resp.Button[0].Name)
	}
	if resp.Button[0].Key != "V1001_TODAY_MUSIC" {
		t.Errorf("button[0].key: %q", resp.Button[0].Key)
	}
	if len(resp.Button[1].SubButton) != 2 {
		t.Fatalf("len(button[1].sub_button): %d", len(resp.Button[1].SubButton))
	}
	if resp.Button[1].SubButton[0].Type != "view" {
		t.Errorf("sub_button[0].type: %q", resp.Button[1].SubButton[0].Type)
	}
	if resp.Button[1].SubButton[0].URL != "https://www.example.com" {
		t.Errorf("sub_button[0].url: %q", resp.Button[1].SubButton[0].URL)
	}
	if resp.Button[1].SubButton[1].Key != "V1001_LIKE" {
		t.Errorf("sub_button[1].key: %q", resp.Button[1].SubButton[1].Key)
	}
}

func TestDeleteMenu(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method: %s", r.Method)
		}
		if r.URL.Path != "/cgi-bin/menu/delete" {
			t.Errorf("path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if got := r.URL.Query().Get("agentid"); got != "1000002" {
			t.Errorf("agentid: %q", got)
		}

		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"errcode": 0,
			"errmsg":  "ok",
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	err := cc.DeleteMenu(context.Background(), 1000002)
	if err != nil {
		t.Fatal(err)
	}
}
