package aispeech

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestASRLong(t *testing.T) {
	tests := []struct {
		name    string
		handler func(w http.ResponseWriter, r *http.Request)
		wantErr bool
		wantID  string
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/cgi-bin/token":
					_, _ = w.Write([]byte(`{"access_token":"TOK","expires_in":7200}`))
				case "/aispeech/asr/aiasrlong":
					var req ASRLongReq
					_ = json.NewDecoder(r.Body).Decode(&req)
					_, _ = w.Write([]byte(`{"task_id":"task_001","errcode":0,"errmsg":"ok"}`))
				}
			},
			wantErr: false,
			wantID:  "task_001",
		},
		{
			name: "api_error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/cgi-bin/token":
					_, _ = w.Write([]byte(`{"access_token":"TOK","expires_in":7200}`))
				case "/aispeech/asr/aiasrlong":
					_, _ = w.Write([]byte(`{"errcode":40001,"errmsg":"invalid credential"}`))
				}
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(tt.handler))
			defer srv.Close()
			c := newTestClient(t, srv.URL)
			resp, err := c.ASRLong(context.Background(), &ASRLongReq{
				VoiceID:  "v1",
				VoiceURL: "https://example.com/audio.mp3",
				Format:   "mp3",
			})
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
			if !tt.wantErr && resp.TaskID != tt.wantID {
				t.Errorf("got task_id=%s, want %s", resp.TaskID, tt.wantID)
			}
		})
	}
}

func TestASRShort(t *testing.T) {
	tests := []struct {
		name       string
		handler    func(w http.ResponseWriter, r *http.Request)
		wantErr    bool
		wantResult string
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/cgi-bin/token":
					_, _ = w.Write([]byte(`{"access_token":"TOK","expires_in":7200}`))
				case "/aispeech/asr/aiasrshort":
					_, _ = w.Write([]byte(`{"result":"你好世界","errcode":0,"errmsg":"ok"}`))
				}
			},
			wantErr:    false,
			wantResult: "你好世界",
		},
		{
			name: "api_error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/cgi-bin/token":
					_, _ = w.Write([]byte(`{"access_token":"TOK","expires_in":7200}`))
				case "/aispeech/asr/aiasrshort":
					_, _ = w.Write([]byte(`{"errcode":40029,"errmsg":"invalid code"}`))
				}
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(tt.handler))
			defer srv.Close()
			c := newTestClient(t, srv.URL)
			resp, err := c.ASRShort(context.Background(), &ASRShortReq{
				VoiceID:   "v1",
				VoiceData: "base64data==",
				Format:    "wav",
				Rate:      16000,
				Bits:      16,
			})
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
			if !tt.wantErr && resp.Result != tt.wantResult {
				t.Errorf("got result=%q, want %q", resp.Result, tt.wantResult)
			}
		})
	}
}
