package isv

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestParseNotify_RejectsOversizeBody proves the 1 MiB cap fires during the
// body read in decryptNotify, before xml.Unmarshal and signature verification.
// A fresh timestamp passes the pre-read check; the read is rejected before the
// (garbage) signature is examined.
func TestParseNotify_RejectsOversizeBody(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	huge := strings.Repeat("A", (1<<20)+1)
	req := httptest.NewRequest(http.MethodPost,
		"/cb?msg_signature=x&timestamp="+ts+"&nonce=n1",
		strings.NewReader(huge))
	if _, err := c.ParseNotify(req); err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("expected oversize-body rejection with 'exceeds', got %v", err)
	}
}
