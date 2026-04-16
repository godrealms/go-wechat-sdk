package isv

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestGetPermanentCode_SerializesWithCorpRefresh is a concurrency regression
// guard for audit Batch 3. GetPermanentCode must acquire the same per-corpID
// mutex that CorpClient.refreshLocked holds, so that an in-flight corp_token
// refresh cannot clobber a freshly-obtained permanent_code record.
//
// The test holds the per-corpID mutex before calling GetPermanentCode and
// asserts that the call blocks waiting on the mutex — proving the store
// write is inside the critical section. Pre-fix, the PutAuthorizer happened
// unconditionally outside any lock and this assertion would fail.
func TestGetPermanentCode_SerializesWithCorpRefresh(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":   "CORP_TOK",
			"expires_in":     7200,
			"permanent_code": "PERM",
			"auth_corp_info": map[string]any{
				"corpid":    "wxLockedCorp",
				"corp_name": "ACME",
			},
		})
	}))
	defer srv.Close()

	c := newTestISVClient(t, srv.URL)
	_ = c.store.PutSuiteToken(context.Background(), "suite1", "STOK", time.Now().Add(time.Hour))

	// Pre-hold the per-corpID mutex for the corp the server will return.
	lock := c.lockFor("wxLockedCorp")
	lock.Lock()

	done := make(chan error, 1)
	go func() {
		_, err := c.GetPermanentCode(context.Background(), "auth_code_xyz")
		done <- err
	}()

	select {
	case err := <-done:
		lock.Unlock()
		t.Fatalf("GetPermanentCode returned while corp mutex was held (err=%v) — the store write is NOT inside the per-corpID critical section", err)
	case <-time.After(200 * time.Millisecond):
		// expected: blocked on mutex
	}

	lock.Unlock()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("GetPermanentCode failed after mutex released: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("GetPermanentCode did not complete after mutex was released")
	}

	// Verify fresh tokens landed in the store.
	got, err := c.store.GetAuthorizer(context.Background(), "suite1", "wxLockedCorp")
	if err != nil {
		t.Fatal(err)
	}
	if got.PermanentCode != "PERM" || got.CorpAccessToken != "CORP_TOK" {
		t.Fatalf("stored: %+v", got)
	}
}

func TestGetPermanentCode_StoresAuthorizer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":   "CORP_TOK",
			"expires_in":     7200,
			"permanent_code": "PERM",
			"auth_corp_info": map[string]any{
				"corpid":    "wxcorp1",
				"corp_name": "ACME",
			},
			"auth_info": map[string]any{
				"agent": []map[string]any{
					{"agentid": 1000001, "name": "HR"},
				},
			},
			"auth_user_info": map[string]any{
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
		_ = json.NewEncoder(w).Encode(map[string]any{
			"auth_corp_info": map[string]any{
				"corpid":    "wxcorp1",
				"corp_name": "ACME",
			},
			"auth_info": map[string]any{
				"agent": []map[string]any{{"agentid": 1}},
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
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["auth_corpid"] != "wxcorp1" {
			t.Errorf("body: %+v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"admin": []map[string]any{
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
