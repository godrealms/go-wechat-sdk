package isv

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func seedSuiteToken(t *testing.T, c *Client) {
	t.Helper()
	_ = c.store.PutSuiteToken(context.Background(), "suite1", "STOK", time.Now().Add(time.Hour))
}

func TestGetPreAuthCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/service/get_pre_auth_code" {
			t.Errorf("path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("suite_access_token"); got != "STOK" {
			t.Errorf("token query: %q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"pre_auth_code": "PCODE",
			"expires_in":    1200,
		})
	}))
	defer srv.Close()

	c := newTestISVClient(t, srv.URL)
	seedSuiteToken(t, c)

	resp, err := c.GetPreAuthCode(context.Background())
	if err != nil || resp.PreAuthCode != "PCODE" {
		t.Fatalf("got %+v err=%v", resp, err)
	}
}

func TestSetSessionInfo(t *testing.T) {
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()

	c := newTestISVClient(t, srv.URL)
	seedSuiteToken(t, c)

	info := &SessionInfo{AppID: []int{1, 2}, AuthType: 1}
	if err := c.SetSessionInfo(context.Background(), "PCODE", info); err != nil {
		t.Fatal(err)
	}
	if gotBody["pre_auth_code"] != "PCODE" {
		t.Errorf("pre_auth_code missing: %+v", gotBody)
	}
	sess, ok := gotBody["session_info"].(map[string]interface{})
	if !ok {
		t.Fatalf("session_info missing: %+v", gotBody)
	}
	if sess["auth_type"].(float64) != 1 {
		t.Errorf("auth_type wrong: %+v", sess)
	}
}

func TestAuthorizeURL(t *testing.T) {
	c, err := NewClient(testConfig())
	if err != nil {
		t.Fatal(err)
	}
	got := c.AuthorizeURL("PCODE", "https://cb.example/ret", "state1")
	u, err := url.Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(got, "https://open.work.weixin.qq.com/3rdapp/install?") {
		t.Errorf("url prefix: %q", got)
	}
	q := u.Query()
	if q.Get("suite_id") != "suite1" || q.Get("pre_auth_code") != "PCODE" ||
		q.Get("redirect_uri") != "https://cb.example/ret" || q.Get("state") != "state1" {
		t.Errorf("query: %+v", q)
	}
}
