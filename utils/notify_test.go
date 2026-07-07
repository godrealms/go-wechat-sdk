package utils

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

// errReader always fails, to exercise the read-error branch.
type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("boom") }

func newBodyReq(body string) *http.Request {
	return &http.Request{Body: io.NopCloser(strings.NewReader(body))}
}

func TestReadNotifyBody_UnderLimit(t *testing.T) {
	req := newBodyReq("hello")
	got, err := ReadNotifyBody(req, 1<<20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}
}

func TestReadNotifyBody_AtLimit(t *testing.T) {
	body := strings.Repeat("A", 8)
	got, err := ReadNotifyBody(newBodyReq(body), 8)
	if err != nil {
		t.Fatalf("body exactly at limit must be accepted, got %v", err)
	}
	if len(got) != 8 {
		t.Errorf("got %d bytes, want 8", len(got))
	}
}

func TestReadNotifyBody_OverLimit(t *testing.T) {
	body := strings.Repeat("A", 9) // limit+1
	_, err := ReadNotifyBody(newBodyReq(body), 8)
	if err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("expected 'exceeds' error for over-limit body, got %v", err)
	}
	if !strings.Contains(err.Error(), "8 bytes") {
		t.Errorf("error should name the limit, got %v", err)
	}
}

func TestReadNotifyBody_DefaultLimit(t *testing.T) {
	// limit<=0 falls back to DefaultMaxNotifyBodySize.
	over := strings.Repeat("A", int(DefaultMaxNotifyBodySize)+1)
	if _, err := ReadNotifyBody(newBodyReq(over), 0); err == nil ||
		!strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("expected default-limit rejection, got %v", err)
	}
	// A modest body under the default limit is fine.
	if _, err := ReadNotifyBody(newBodyReq("ok"), 0); err != nil {
		t.Errorf("under-default body should pass, got %v", err)
	}
	if _, err := ReadNotifyBody(newBodyReq("ok"), -1); err != nil {
		t.Errorf("negative limit should fall back to default, got %v", err)
	}
}

func TestReadNotifyBody_NilRequestAndBody(t *testing.T) {
	if _, err := ReadNotifyBody(nil, 0); err == nil {
		t.Error("expected error for nil request")
	}
	if _, err := ReadNotifyBody(&http.Request{}, 0); err == nil {
		t.Error("expected error for nil body")
	}
}

func TestReadNotifyBody_ReadError(t *testing.T) {
	req := &http.Request{Body: io.NopCloser(errReader{})}
	if _, err := ReadNotifyBody(req, 16); err == nil ||
		!strings.Contains(err.Error(), "read notify body") {
		t.Fatalf("expected read error, got %v", err)
	}
}
