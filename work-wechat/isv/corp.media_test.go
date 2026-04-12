package isv

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUploadMedia(t *testing.T) {
	const fileContent = "fake image data"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: got %s, want POST", r.Method)
		}
		if r.URL.Path != "/cgi-bin/media/upload" {
			t.Errorf("path: got %s, want /cgi-bin/media/upload", r.URL.Path)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: got %q, want CTOK", got)
		}
		if got := r.URL.Query().Get("type"); got != "image" {
			t.Errorf("type: got %q, want image", got)
		}
		ct := r.Header.Get("Content-Type")
		if !strings.Contains(ct, "multipart/form-data") {
			t.Fatalf("Content-Type: got %q, want multipart/form-data", ct)
		}

		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("parse multipart: %v", err)
		}
		f, fh, err := r.FormFile("media")
		if err != nil {
			t.Fatalf("form file 'media': %v", err)
		}
		defer f.Close()

		if fh.Filename != "test.png" {
			t.Errorf("filename: got %q, want test.png", fh.Filename)
		}
		data, _ := io.ReadAll(f)
		if string(data) != fileContent {
			t.Errorf("file data: got %q, want %q", string(data), fileContent)
		}

		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"type":       "image",
			"media_id":   "MID001",
			"created_at": "1712900000",
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	resp, err := cc.UploadMedia(context.Background(), "image", "test.png", strings.NewReader(fileContent))
	if err != nil {
		t.Fatal(err)
	}
	if resp.Type != "image" {
		t.Errorf("Type: got %q, want image", resp.Type)
	}
	if resp.MediaID != "MID001" {
		t.Errorf("MediaID: got %q, want MID001", resp.MediaID)
	}
	if resp.CreatedAt != "1712900000" {
		t.Errorf("CreatedAt: got %q, want 1712900000", resp.CreatedAt)
	}
}

func TestUploadMedia_WeixinError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"errcode": 40004,
			"errmsg":  "invalid media type",
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	_, err := cc.UploadMedia(context.Background(), "bad", "test.png", strings.NewReader("data"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var we *WeixinError
	if !errors.As(err, &we) {
		t.Fatalf("expected WeixinError, got %T: %v", err, err)
	}
	if we.ErrCode != 40004 {
		t.Errorf("ErrCode: got %d, want 40004", we.ErrCode)
	}
}
