package aispeech

import (
	"context"
	"net/http/httptest"
	"testing"
)

func TestDialogQuery(t *testing.T) {
	tests := []struct {
		name       string
		respBody   string
		wantErr    bool
		wantAnswer string
	}{
		{
			name:       "success",
			respBody:   `{"answer":"今天北京天气晴","session_id":"s1","end_flag":false,"errcode":0,"errmsg":"ok"}`,
			wantErr:    false,
			wantAnswer: "今天北京天气晴",
		},
		{
			name:     "api_error",
			respBody: `{"errcode":40001,"errmsg":"invalid credential"}`,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(handler200("/aispeech/dialog/airequ", tt.respBody))
			defer srv.Close()
			c := newTestClient(t, srv.URL)
			resp, err := c.DialogQuery(context.Background(), &DialogQueryReq{
				Query:     "北京天气",
				SessionID: "s1",
			})
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
			if !tt.wantErr && resp.Answer != tt.wantAnswer {
				t.Errorf("got answer=%q, want %q", resp.Answer, tt.wantAnswer)
			}
		})
	}
}

func TestDialogReset(t *testing.T) {
	tests := []struct {
		name     string
		respBody string
		wantErr  bool
	}{
		{
			name:     "success",
			respBody: `{"errcode":0,"errmsg":"ok"}`,
			wantErr:  false,
		},
		{
			name:     "api_error",
			respBody: `{"errcode":40001,"errmsg":"invalid credential"}`,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(handler200("/aispeech/dialog/aireset", tt.respBody))
			defer srv.Close()
			c := newTestClient(t, srv.URL)
			err := c.DialogReset(context.Background(), &DialogResetReq{SessionID: "s1"})
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}
