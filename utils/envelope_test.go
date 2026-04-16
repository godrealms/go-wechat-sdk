package utils

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// testAPIError is a fake package-specific error for testing DecodeEnvelope.
type testAPIError struct {
	Code int
	Msg  string
	Path string
}

func (e *testAPIError) Error() string {
	return fmt.Sprintf("test: %s errcode=%d errmsg=%s", e.Path, e.Code, e.Msg)
}

func testErrFactory(code int, msg, path string) error {
	return &testAPIError{Code: code, Msg: msg, Path: path}
}

func TestDecodeEnvelope_Success(t *testing.T) {
	type result struct {
		Name string `json:"name"`
	}
	var out result
	body := []byte(`{"errcode":0,"errmsg":"ok","name":"Alice"}`)
	err := DecodeEnvelope("test", "/api/foo", body, &out, testErrFactory)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Name != "Alice" {
		t.Errorf("expected Name=Alice, got %q", out.Name)
	}
}

func TestDecodeEnvelope_APIError(t *testing.T) {
	body := []byte(`{"errcode":40001,"errmsg":"invalid credential"}`)
	err := DecodeEnvelope("test", "/api/foo", body, nil, testErrFactory)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *testAPIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *testAPIError, got %T: %v", err, err)
	}
	if apiErr.Code != 40001 {
		t.Errorf("expected code 40001, got %d", apiErr.Code)
	}
	if apiErr.Msg != "invalid credential" {
		t.Errorf("expected msg 'invalid credential', got %q", apiErr.Msg)
	}
	if apiErr.Path != "/api/foo" {
		t.Errorf("expected path /api/foo, got %q", apiErr.Path)
	}
}

func TestDecodeEnvelope_MalformedJSON(t *testing.T) {
	body := []byte(`<html>503 Bad Gateway</html>`)
	err := DecodeEnvelope("mypkg", "/api/bar", body, nil, testErrFactory)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "mypkg") {
		t.Errorf("expected package prefix in error, got: %s", msg)
	}
	if !strings.Contains(msg, "decode envelope") {
		t.Errorf("expected 'decode envelope' in error, got: %s", msg)
	}
	if !strings.Contains(msg, "503 Bad Gateway") {
		t.Errorf("expected body snippet in error, got: %s", msg)
	}
}

func TestDecodeEnvelope_NilOut(t *testing.T) {
	body := []byte(`{"errcode":0,"errmsg":"ok"}`)
	err := DecodeEnvelope("test", "/api/foo", body, nil, testErrFactory)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDecodeEnvelope_DecodeResultError(t *testing.T) {
	// Body is valid JSON but cannot be decoded into the target struct
	// because the field types don't match.
	type strict struct {
		Count int `json:"count"`
	}
	var out strict
	body := []byte(`{"errcode":0,"errmsg":"ok","count":"not_a_number"}`)
	err := DecodeEnvelope("test", "/api/foo", body, &out, testErrFactory)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "decode result") {
		t.Errorf("expected 'decode result' in error, got: %s", err.Error())
	}
}

func TestSnippet_Short(t *testing.T) {
	data := []byte("hello")
	if got := Snippet(data); got != "hello" {
		t.Errorf("expected 'hello', got %q", got)
	}
}

func TestSnippet_Long(t *testing.T) {
	data := make([]byte, 300)
	for i := range data {
		data[i] = 'x'
	}
	got := Snippet(data)
	if len(got) != 200+len("...(truncated)") {
		t.Errorf("unexpected length: %d", len(got))
	}
	if !strings.HasSuffix(got, "...(truncated)") {
		t.Errorf("expected truncation suffix, got: %s", got[190:])
	}
}

func TestSnippet_ExactBoundary(t *testing.T) {
	data := make([]byte, 200)
	for i := range data {
		data[i] = 'y'
	}
	got := Snippet(data)
	if strings.Contains(got, "truncated") {
		t.Errorf("200 bytes should not be truncated")
	}
	if len(got) != 200 {
		t.Errorf("expected 200, got %d", len(got))
	}
}
