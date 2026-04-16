package isv

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInviteUser(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if r.URL.Path != "/cgi-bin/batch/invite" {
			t.Errorf("path: %s", r.URL.Path)
		}
		var body InviteReq
		_ = json.NewDecoder(r.Body).Decode(&body)
		if len(body.User) != 2 {
			t.Errorf("len(body.User): %d", len(body.User))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errcode":      0,
			"errmsg":       "ok",
			"invaliduser":  []string{"lisi"},
			"invalidparty": []int{},
			"invalidtag":   []int{},
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	resp, err := cc.InviteUser(context.Background(), &InviteReq{
		User: []string{"zhangsan", "lisi"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.InvalidUser) != 1 || resp.InvalidUser[0] != "lisi" {
		t.Errorf("InvalidUser: %v", resp.InvalidUser)
	}
}
