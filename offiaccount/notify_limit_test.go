package offiaccount

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

// oversizeNotifyBody is 1 byte over the shared 1 MiB notify cap.
func oversizeNotifyBody() string { return strings.Repeat("A", (1<<20)+1) }

// TestParseNotify_RejectsOversizeBody proves the size cap fires before the XML
// parse / signature verify on the encrypted POST path: a fresh timestamp and a
// non-nil crypto get us past the pre-read guards, and the (garbage) signature is
// never reached because the read is rejected first.
func TestParseNotify_RejectsOversizeBody(t *testing.T) {
	mc, err := NewMsgCrypto("tk", shimKey(), "wxappid")
	if err != nil {
		t.Fatal(err)
	}
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	req := httptest.NewRequest(http.MethodPost,
		"/wx/notify?msg_signature=x&timestamp="+ts+"&nonce=n1",
		strings.NewReader(oversizeNotifyBody()))
	if _, err := ParseNotify(req, mc); err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("expected oversize-body rejection with 'exceeds', got %v", err)
	}
}

// TestParseNotifyPlaintext_RejectsOversizeBody covers the plaintext path, whose
// signature is verified before the body read — so a *valid* signature is
// required to reach (and trip) the cap. This guards the case a captured valid
// signature is replayed with a huge body.
func TestParseNotifyPlaintext_RejectsOversizeBody(t *testing.T) {
	const token = "tk"
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	const nonce = "n1"
	sig := computeOffiaccountSig(token, ts, nonce)
	req := httptest.NewRequest(http.MethodPost,
		"/wx/notify?signature="+sig+"&timestamp="+ts+"&nonce="+nonce,
		strings.NewReader(oversizeNotifyBody()))
	if _, err := ParseNotifyPlaintext(req, token); err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("expected oversize-body rejection with 'exceeds', got %v", err)
	}
}
