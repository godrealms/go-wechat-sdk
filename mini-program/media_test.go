package mini_program

import (
	"context"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

func TestUploadTempMedia(t *testing.T) {
	srv := httptest.NewServer(tokenHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/media/upload" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Query().Get("access_token") != "TOK" {
			t.Errorf("missing or wrong access_token: %q", r.URL.Query().Get("access_token"))
		}
		if r.URL.Query().Get("type") != "image" {
			t.Errorf("expected type=image, got %q", r.URL.Query().Get("type"))
		}
		ct := r.Header.Get("Content-Type")
		if !strings.HasPrefix(ct, "multipart/form-data") {
			t.Errorf("expected multipart/form-data Content-Type, got %q", ct)
		}
		// Parse multipart body and verify "media" field with filename
		_, params, err := mime.ParseMediaType(ct)
		if err != nil {
			t.Fatalf("failed to parse Content-Type: %v", err)
		}
		mr := multipart.NewReader(r.Body, params["boundary"])
		part, err := mr.NextPart()
		if err != nil {
			t.Fatalf("failed to read multipart part: %v", err)
		}
		if part.FormName() != "media" {
			t.Errorf("expected form name 'media', got %q", part.FormName())
		}
		if part.FileName() == "" {
			t.Error("expected a non-empty filename in multipart part")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"type":"image","media_id":"MEDIA_ID_123","created_at":1234567890}`))
	}))
	defer srv.Close()

	c, err := NewClient(Config{AppId: "wx", AppSecret: "sec"},
		WithHTTP(utils.NewHTTP(srv.URL, utils.WithTimeout(time.Second*3))))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := c.UploadTempMedia(context.Background(), "image", "test.jpg", strings.NewReader("fake-file-content"))
	if err != nil {
		t.Fatal(err)
	}
	if resp.Type != "image" {
		t.Errorf("expected Type=image, got %q", resp.Type)
	}
	if resp.MediaID != "MEDIA_ID_123" {
		t.Errorf("expected MediaID=MEDIA_ID_123, got %q", resp.MediaID)
	}
	if resp.CreatedAt != 1234567890 {
		t.Errorf("expected CreatedAt=1234567890, got %d", resp.CreatedAt)
	}
}

func TestGetTempMedia(t *testing.T) {
	srv := httptest.NewServer(tokenHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/media/get" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Query().Get("access_token") != "TOK" {
			t.Errorf("missing or wrong access_token: %q", r.URL.Query().Get("access_token"))
		}
		if r.URL.Query().Get("media_id") != "MEDIA_123" {
			t.Errorf("expected media_id=MEDIA_123, got %q", r.URL.Query().Get("media_id"))
		}
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = w.Write([]byte("fake-image-data"))
	}))
	defer srv.Close()

	c, err := NewClient(Config{AppId: "wx", AppSecret: "sec"},
		WithHTTP(utils.NewHTTP(srv.URL, utils.WithTimeout(time.Second*3))))
	if err != nil {
		t.Fatal(err)
	}
	got, err := c.GetTempMedia(context.Background(), "MEDIA_123")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "fake-image-data" {
		t.Errorf("unexpected bytes: %q", got)
	}
}
