package aispeech

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func handler200(path, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/token":
			_, _ = w.Write([]byte(`{"access_token":"TOK","expires_in":7200}`))
		case path:
			_, _ = w.Write([]byte(body))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func TestNLUUnderstand(t *testing.T) {
	tests := []struct {
		name       string
		respBody   string
		wantErr    bool
		wantIntent string
	}{
		{
			name:       "success",
			respBody:   `{"intent":"weather_query","slots":[{"type":"LOCATION","value":"北京","begin":2,"end":4}],"session_id":"s1","errcode":0,"errmsg":"ok"}`,
			wantErr:    false,
			wantIntent: "weather_query",
		},
		{
			name:     "api_error",
			respBody: `{"errcode":40001,"errmsg":"invalid credential"}`,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(handler200("/aispeech/nlu/airequ", tt.respBody))
			defer srv.Close()
			c := newTestClient(t, srv.URL)
			resp, err := c.NLUUnderstand(context.Background(), &NLUUnderstandReq{Query: "北京天气"})
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
			if !tt.wantErr && resp.Intent != tt.wantIntent {
				t.Errorf("got intent=%q, want %q", resp.Intent, tt.wantIntent)
			}
		})
	}
}

func TestNLUIntentRecognize(t *testing.T) {
	tests := []struct {
		name       string
		respBody   string
		wantErr    bool
		wantIntent string
	}{
		{
			name:       "success",
			respBody:   `{"intent_id":"i1","intent_name":"查天气","confidence":0.95,"session_id":"s1","errcode":0,"errmsg":"ok"}`,
			wantErr:    false,
			wantIntent: "i1",
		},
		{
			name:     "api_error",
			respBody: `{"errcode":40003,"errmsg":"invalid openid"}`,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(handler200("/aispeech/nlu/aiintentrequ", tt.respBody))
			defer srv.Close()
			c := newTestClient(t, srv.URL)
			resp, err := c.NLUIntentRecognize(context.Background(), &NLUIntentRecognizeReq{Query: "查天气"})
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
			if !tt.wantErr && resp.IntentID != tt.wantIntent {
				t.Errorf("got intent_id=%q, want %q", resp.IntentID, tt.wantIntent)
			}
		})
	}
}
