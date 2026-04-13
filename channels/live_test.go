package channels

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateRoom(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/token":
			_, _ = w.Write([]byte(`{"access_token":"TEST_TOKEN","expires_in":7200}`))
		case "/channels/ec/basics/live/createroom":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Query().Get("access_token") != "TEST_TOKEN" {
				t.Errorf("missing access_token")
			}
			var req CreateRoomReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatal(err)
			}
			if req.Name != "test_room" {
				t.Errorf("expected name test_room, got %s", req.Name)
			}
			_, _ = w.Write([]byte(`{"room_id":"room_001"}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	resp, err := c.CreateRoom(context.Background(), &CreateRoomReq{
		Name:      "test_room",
		StartTime: 1000,
		EndTime:   2000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.RoomID != "room_001" {
		t.Errorf("got room_id=%s, want room_001", resp.RoomID)
	}
}

func TestDeleteRoom(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/token":
			_, _ = w.Write([]byte(`{"access_token":"TEST_TOKEN","expires_in":7200}`))
		case "/channels/ec/basics/live/deleteroom":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			var req DeleteRoomReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatal(err)
			}
			if req.RoomID != "room_001" {
				t.Errorf("expected room_id room_001, got %s", req.RoomID)
			}
			_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	if err := c.DeleteRoom(context.Background(), &DeleteRoomReq{RoomID: "room_001"}); err != nil {
		t.Fatal(err)
	}
}

func TestGetLiveInfo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/token":
			_, _ = w.Write([]byte(`{"access_token":"TEST_TOKEN","expires_in":7200}`))
		case "/channels/ec/basics/live/getliveinfo":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			var req GetLiveInfoReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatal(err)
			}
			if req.RoomID != "room_001" {
				t.Errorf("expected room_id room_001, got %s", req.RoomID)
			}
			_, _ = w.Write([]byte(`{"live_info":{"room_id":"room_001","name":"live","status":1,"start_time":1000,"end_time":2000}}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	resp, err := c.GetLiveInfo(context.Background(), &GetLiveInfoReq{RoomID: "room_001"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.LiveInfo.RoomID != "room_001" {
		t.Errorf("got room_id=%s, want room_001", resp.LiveInfo.RoomID)
	}
	if resp.LiveInfo.Status != 1 {
		t.Errorf("got status=%d, want 1", resp.LiveInfo.Status)
	}
}

func TestGetLiveReplayList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/token":
			_, _ = w.Write([]byte(`{"access_token":"TEST_TOKEN","expires_in":7200}`))
		case "/channels/ec/basics/live/getlivereplaylist":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			var req GetLiveReplayListReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatal(err)
			}
			if req.RoomID != "room_001" {
				t.Errorf("expected room_id room_001, got %s", req.RoomID)
			}
			_, _ = w.Write([]byte(`{"live_replay_list":[{"media_url":"https://example.com/replay.mp4","expire_time":9999,"create_time":1000}],"total":1}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	resp, err := c.GetLiveReplayList(context.Background(), &GetLiveReplayListReq{RoomID: "room_001"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Total != 1 {
		t.Errorf("got total=%d, want 1", resp.Total)
	}
	if len(resp.LiveReplayList) != 1 {
		t.Fatalf("got %d replays, want 1", len(resp.LiveReplayList))
	}
	if resp.LiveReplayList[0].MediaURL != "https://example.com/replay.mp4" {
		t.Errorf("unexpected media_url: %s", resp.LiveReplayList[0].MediaURL)
	}
}
