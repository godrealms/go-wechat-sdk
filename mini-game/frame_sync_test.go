package mini_game

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateGameRoom_ErrcodeError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/token":
			_, _ = w.Write([]byte(`{"access_token":"TEST_TOKEN","expires_in":7200}`))
		default:
			_, _ = w.Write([]byte(`{"errcode":40001,"errmsg":"invalid token"}`))
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.CreateGameRoom(context.Background(), &CreateGameRoomReq{MaxNum: 4})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.Code() != 40001 {
		t.Errorf("expected Code() == 40001, got %d", apiErr.Code())
	}
}

func TestCreateGameRoom(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/token":
			_, _ = w.Write([]byte(`{"access_token":"TEST_TOKEN","expires_in":7200}`))
		case "/wxa/game/createroom":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Query().Get("access_token") != "TEST_TOKEN" {
				t.Errorf("missing access_token")
			}
			var req CreateGameRoomReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatal(err)
			}
			if req.MaxNum != 4 {
				t.Errorf("expected max_num 4, got %d", req.MaxNum)
			}
			_, _ = w.Write([]byte(`{"room_id":"room_001"}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	resp, err := c.CreateGameRoom(context.Background(), &CreateGameRoomReq{
		MaxNum:     4,
		AccessInfo: "test_access",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.RoomID != "room_001" {
		t.Errorf("got room_id=%s, want room_001", resp.RoomID)
	}
}

func TestGetRoomInfo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/token":
			_, _ = w.Write([]byte(`{"access_token":"TEST_TOKEN","expires_in":7200}`))
		case "/wxa/game/getroominfo":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Query().Get("access_token") != "TEST_TOKEN" {
				t.Errorf("missing access_token")
			}
			var req GetRoomInfoReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatal(err)
			}
			if req.RoomID != "room_001" {
				t.Errorf("expected room_id room_001, got %s", req.RoomID)
			}
			_, _ = w.Write([]byte(`{"room_id":"room_001","status":1,"members":[{"openid":"oUSER1","role":1},{"openid":"oUSER2","role":0}]}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	resp, err := c.GetRoomInfo(context.Background(), &GetRoomInfoReq{RoomID: "room_001"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.RoomID != "room_001" {
		t.Errorf("got room_id=%s, want room_001", resp.RoomID)
	}
	if resp.Status != 1 {
		t.Errorf("got status=%d, want 1", resp.Status)
	}
	if len(resp.Members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(resp.Members))
	}
	if resp.Members[0].OpenID != "oUSER1" {
		t.Errorf("unexpected first member openid: %s", resp.Members[0].OpenID)
	}
	if resp.Members[0].Role != 1 {
		t.Errorf("unexpected first member role: %d", resp.Members[0].Role)
	}
}
