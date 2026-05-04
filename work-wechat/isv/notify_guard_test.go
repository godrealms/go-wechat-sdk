package isv

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestDecryptNotify_NilRequest verifies the M14 guard: a nil *http.Request
// must not panic; instead a clear error is returned.
func TestDecryptNotify_NilRequest(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	_, err := c.decryptNotify(nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !errors.Is(err, errors.New("isv: nil *http.Request")) {
		// Use textual contain since we did not export the sentinel.
		if !strings.Contains(err.Error(), "nil *http.Request") {
			t.Errorf("error text mismatch: %v", err)
		}
	}
}

func TestDecryptNotify_NilBody(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	r, _ := http.NewRequest(http.MethodPost, "/cb?msg_signature=s&timestamp=t&nonce=n", nil)
	r.Body = nil
	_, err := c.decryptNotify(r)
	if err == nil {
		t.Fatal("expected error for nil body")
	}
	if !strings.Contains(err.Error(), "nil request body") {
		t.Errorf("error text mismatch: %v", err)
	}
}

// TestDecryptNotify_TimestampCheckRunsBeforeBodyRead verifies the M13 ordering
// fix: a stale-timestamp request is rejected without consuming the body, so a
// hostile sender cannot force XML parsing of attacker-controlled bytes via a
// stale-timestamp gate. We assert ordering by checking that a syntactically
// invalid XML body is NOT what the error reports — instead we get the
// timestamp error.
func TestDecryptNotify_TimestampCheckRunsBeforeBodyRead(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	garbage := []byte("not xml at all <<<<<<<<")
	// Stale timestamp far outside the ±5 min window.
	r := httptest.NewRequest(http.MethodPost,
		"/cb?msg_signature=s&timestamp=1700000000&nonce=n",
		bytes.NewReader(garbage))

	_, err := c.decryptNotify(r)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "timestamp") {
		t.Errorf("expected timestamp error, got: %v (would mean body parse ran first)", err)
	}
}
