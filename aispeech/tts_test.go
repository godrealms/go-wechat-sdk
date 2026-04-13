package aispeech

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTextToSpeech(t *testing.T) {
	tests := []struct {
		name      string
		handler   func(w http.ResponseWriter, r *http.Request)
		wantErr   bool
		wantAudio string
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/cgi-bin/token":
					_, _ = w.Write([]byte(`{"access_token":"TOK","expires_in":7200}`))
				case "/aispeech/tts/aitts":
					_, _ = w.Write([]byte(`{"audio_data":"AAEC","audio_size":3,"session_id":"s1","errcode":0,"errmsg":"ok"}`))
				}
			},
			wantErr:   false,
			wantAudio: "AAEC",
		},
		{
			name: "api_error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/cgi-bin/token":
					_, _ = w.Write([]byte(`{"access_token":"TOK","expires_in":7200}`))
				case "/aispeech/tts/aitts":
					_, _ = w.Write([]byte(`{"errcode":45009,"errmsg":"reach max api daily quota limit"}`))
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
			resp, err := c.TextToSpeech(context.Background(), &TextToSpeechReq{Text: "你好"})
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
			if !tt.wantErr && resp.AudioData != tt.wantAudio {
				t.Errorf("got audio_data=%q, want %q", resp.AudioData, tt.wantAudio)
			}
		})
	}
}
