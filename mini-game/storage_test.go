package mini_game

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetUserStorage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/token":
			_, _ = w.Write([]byte(`{"access_token":"TEST_TOKEN","expires_in":7200}`))
		case "/wxa/set_user_storage":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Query().Get("access_token") != "TEST_TOKEN" {
				t.Errorf("missing access_token")
			}
			var req SetUserStorageReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatal(err)
			}
			if req.OpenID != "oUSER1" {
				t.Errorf("expected openid oUSER1, got %s", req.OpenID)
			}
			if len(req.KVList) != 1 {
				t.Fatalf("expected 1 kv item, got %d", len(req.KVList))
			}
			if req.KVList[0].Key != "score" {
				t.Errorf("expected key score, got %s", req.KVList[0].Key)
			}
			_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	err := c.SetUserStorage(context.Background(), &SetUserStorageReq{
		OpenID:    "oUSER1",
		KVList:    []KVData{{Key: "score", Value: "100"}},
		SigMethod: "hmac_sha256",
		Signature: "sig_abc",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetUserStorage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/token":
			_, _ = w.Write([]byte(`{"access_token":"TEST_TOKEN","expires_in":7200}`))
		case "/wxa/get_user_storage":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Query().Get("access_token") != "TEST_TOKEN" {
				t.Errorf("missing access_token")
			}
			var req GetUserStorageReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatal(err)
			}
			if req.OpenID != "oUSER1" {
				t.Errorf("expected openid oUSER1, got %s", req.OpenID)
			}
			if len(req.KeyList) != 1 || req.KeyList[0] != "score" {
				t.Errorf("expected key_list [score], got %v", req.KeyList)
			}
			_, _ = w.Write([]byte(`{"kv_list":[{"key":"score","value":"100"}]}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	resp, err := c.GetUserStorage(context.Background(), &GetUserStorageReq{
		OpenID:    "oUSER1",
		KeyList:   []string{"score"},
		SigMethod: "hmac_sha256",
		Signature: "sig_abc",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.KVList) != 1 {
		t.Fatalf("expected 1 kv item, got %d", len(resp.KVList))
	}
	if resp.KVList[0].Key != "score" {
		t.Errorf("unexpected key: %s", resp.KVList[0].Key)
	}
	if resp.KVList[0].Value != "100" {
		t.Errorf("unexpected value: %s", resp.KVList[0].Value)
	}
}
