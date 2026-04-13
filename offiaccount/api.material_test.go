package offiaccount

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetTempMedia_BinaryResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		w.WriteHeader(200)
		_, _ = w.Write([]byte{0xFF, 0xD8, 0xFF})
	}))
	defer srv.Close()

	c := newMenuTestClient(t, srv)
	data, videoResult, err := c.GetTempMedia(context.Background(), "media123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if videoResult != nil {
		t.Error("expected nil videoResult for binary response")
	}
	if len(data) == 0 {
		t.Error("expected non-empty binary data")
	}
}

func TestGetTempMedia_JSONResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"video_url":"https://video.weixin.qq.com/xxx","down_url":"https://video.weixin.qq.com/yyy"}`))
	}))
	defer srv.Close()

	c := newMenuTestClient(t, srv)
	data, videoResult, err := c.GetTempMedia(context.Background(), "media123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data != nil {
		t.Error("expected nil data for JSON response")
	}
	if videoResult == nil {
		t.Error("expected non-nil videoResult")
	}
}

func TestGetTempMedia_NetworkError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close()
	c := newMenuTestClient(t, srv)
	_, _, err := c.GetTempMedia(context.Background(), "media123")
	if err == nil {
		t.Error("expected network error")
	}
}

func TestGetMaterial_News(t *testing.T) {
	body := `{"news_item":[{"title":"Article 1","author":"Author","content":"Hello"}]}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	c := newMenuTestClient(t, srv)
	newsResult, videoResult, rawData, err := c.GetMaterial(context.Background(), "media_news_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = videoResult
	_ = rawData
	if newsResult == nil {
		t.Error("expected non-nil newsResult")
	}
}

func TestGetMaterial_NetworkError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close()
	c := newMenuTestClient(t, srv)
	_, _, _, err := c.GetMaterial(context.Background(), "media123")
	if err == nil {
		t.Error("expected network error")
	}
}

func TestAddDraft_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"media_id":"draft123"}`))
	}))
	defer srv.Close()

	c := newMenuTestClient(t, srv)
	result, err := c.AddDraft(context.Background(), []*DraftArticle{{Title: "Hello", Content: "World"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestAddDraft_NetworkError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close()
	c := newMenuTestClient(t, srv)
	_, err := c.AddDraft(context.Background(), []*DraftArticle{})
	if err == nil {
		t.Error("expected network error")
	}
}

func TestGetDraft_Success(t *testing.T) {
	body := `{"news_item":[{"title":"Draft 1","author":"Writer"}],"update_time":1700000000}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	c := newMenuTestClient(t, srv)
	result, err := c.GetDraft(context.Background(), "draft123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}
