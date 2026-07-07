package oplatform

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestParseNotify_RejectsOversizeBody proves the 1 MiB cap fires during the
// body read, before xml.Unmarshal and signature verification. A fresh timestamp
// passes the pre-read check; the read is rejected before the (garbage)
// signature is ever examined.
func TestParseNotify_RejectsOversizeBody(t *testing.T) {
	c, _ := NewClient(testConfig())
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	huge := strings.Repeat("A", (1<<20)+1)
	req := httptest.NewRequest(http.MethodPost,
		"/oplatform/notify?msg_signature=x&timestamp="+ts+"&nonce=n1",
		strings.NewReader(huge))
	if _, err := c.ParseNotify(req, nil); err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("expected oversize-body rejection with 'exceeds', got %v", err)
	}
}
