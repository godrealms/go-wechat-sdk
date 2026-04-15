package offiaccount

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"
)

// computeOffiaccountSig mirrors wxcrypto.VerifyServerToken's signature
// algorithm for use in tests.
func computeOffiaccountSig(token, timestamp, nonce string) string {
	parts := []string{token, timestamp, nonce}
	sort.Strings(parts)
	h := sha1.New()
	_, _ = h.Write([]byte(strings.Join(parts, "")))
	return hex.EncodeToString(h.Sum(nil))
}

// These tests ensure the offiaccount shim still exposes the same
// behavior through type aliasing. The full Biz Msg Crypt test suite
// lives in utils/wxcrypto.

func shimKey() string {
	raw := make([]byte, 32)
	for i := range raw {
		raw[i] = byte(i)
	}
	return strings.TrimSuffix(base64.StdEncoding.EncodeToString(raw), "=")
}

func TestOffiaccountMsgCryptoShim_Roundtrip(t *testing.T) {
	mc, err := NewMsgCrypto("tk", shimKey(), "wxappid")
	if err != nil {
		t.Fatal(err)
	}
	plain := []byte(`<xml><Content><![CDATA[hi]]></Content></xml>`)
	enc, err := mc.Encrypt(plain)
	if err != nil {
		t.Fatal(err)
	}
	got, _, err := mc.Decrypt(enc)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(plain) {
		t.Errorf("roundtrip mismatch")
	}
}

func TestParseNotify_EncryptedMode(t *testing.T) {
	mc, _ := NewMsgCrypto("tk", shimKey(), "wxappid")
	plain := []byte(`<xml><MsgType><![CDATA[text]]></MsgType><Content><![CDATA[hi]]></Content></xml>`)
	encrypted, err := mc.Encrypt(plain)
	if err != nil {
		t.Fatal(err)
	}
	ts := fmt.Sprintf("%d", time.Now().Unix())
	sig := mc.Signature(ts, "n1", encrypted)
	bodyXML := `<xml><ToUserName>gh_x</ToUserName><Encrypt><![CDATA[` + encrypted + `]]></Encrypt></xml>`
	req := httptest.NewRequest(http.MethodPost,
		"/notify?msg_signature="+sig+"&timestamp="+ts+"&nonce=n1",
		strings.NewReader(bodyXML))
	got, err := ParseNotify(req, mc)
	if err != nil {
		t.Fatal(err)
	}
	var env struct {
		MsgType string `xml:"MsgType"`
		Content string `xml:"Content"`
	}
	if err := xml.Unmarshal(got, &env); err != nil {
		t.Fatal(err)
	}
	if env.MsgType != "text" || env.Content != "hi" {
		t.Errorf("unexpected decrypted: %+v", env)
	}
}

func TestParseNotify_BadSignature(t *testing.T) {
	mc, _ := NewMsgCrypto("tk", shimKey(), "wxappid")
	encrypted, _ := mc.Encrypt([]byte(`<xml/>`))
	bodyXML := `<xml><Encrypt><![CDATA[` + encrypted + `]]></Encrypt></xml>`
	ts := fmt.Sprintf("%d", time.Now().Unix())
	req := httptest.NewRequest(http.MethodPost,
		"/notify?msg_signature=deadbeef&timestamp="+ts+"&nonce=n1",
		strings.NewReader(bodyXML))
	if _, err := ParseNotify(req, mc); err == nil {
		t.Error("expected signature error")
	}
}

// TestParseNotify_RejectsStaleTimestamp verifies the ±5min replay window.
func TestParseNotify_RejectsStaleTimestamp(t *testing.T) {
	mc, _ := NewMsgCrypto("tk", shimKey(), "wxappid")
	encrypted, _ := mc.Encrypt([]byte(`<xml/>`))
	// Timestamp from 10 minutes ago — well outside the ±5min window.
	stale := fmt.Sprintf("%d", time.Now().Add(-10*time.Minute).Unix())
	sig := mc.Signature(stale, "n1", encrypted)
	bodyXML := `<xml><Encrypt><![CDATA[` + encrypted + `]]></Encrypt></xml>`
	req := httptest.NewRequest(http.MethodPost,
		"/notify?msg_signature="+sig+"&timestamp="+stale+"&nonce=n1",
		strings.NewReader(bodyXML))
	_, err := ParseNotify(req, mc)
	if err == nil {
		t.Fatal("expected stale-timestamp error")
	}
	if !strings.Contains(err.Error(), "out of ±5min window") {
		t.Errorf("expected replay-window error, got %v", err)
	}
}

// TestParseNotify_RejectsMissingTimestamp makes sure an empty timestamp is a
// hard fail, not a silent accept.
func TestParseNotify_RejectsMissingTimestamp(t *testing.T) {
	mc, _ := NewMsgCrypto("tk", shimKey(), "wxappid")
	req := httptest.NewRequest(http.MethodPost, "/notify?nonce=n1", strings.NewReader(""))
	if _, err := ParseNotify(req, mc); err == nil {
		t.Error("expected missing-timestamp error")
	}
}

// TestParseNotify_RejectsNilCrypto ensures the legacy "crypto == nil accepts
// unsigned body" behavior is gone.
func TestParseNotify_RejectsNilCrypto(t *testing.T) {
	ts := fmt.Sprintf("%d", time.Now().Unix())
	req := httptest.NewRequest(http.MethodPost,
		"/notify?timestamp="+ts+"&nonce=n1", strings.NewReader(`<xml/>`))
	_, err := ParseNotify(req, nil)
	if !errors.Is(err, ErrNotifyNoCrypto) {
		t.Fatalf("expected ErrNotifyNoCrypto, got %v", err)
	}
}

// TestParseNotifyPlaintext_Roundtrip verifies the plaintext-mode signature
// check passes for a correctly signed request and that the body is returned.
func TestParseNotifyPlaintext_Roundtrip(t *testing.T) {
	token := "mytoken"
	body := `<xml><MsgType>text</MsgType><Content>hi</Content></xml>`
	ts := fmt.Sprintf("%d", time.Now().Unix())
	nonce := "n1"
	// compute sha1(sort([token,ts,nonce]))
	sig := computeOffiaccountSig(token, ts, nonce)
	req := httptest.NewRequest(http.MethodPost,
		"/notify?signature="+sig+"&timestamp="+ts+"&nonce="+nonce,
		strings.NewReader(body))
	got, err := ParseNotifyPlaintext(req, token)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != body {
		t.Errorf("body mismatch: %q", got)
	}
}

// TestParseNotifyPlaintext_RejectsBadSig confirms a wrong-token signature
// fails.
func TestParseNotifyPlaintext_RejectsBadSig(t *testing.T) {
	ts := fmt.Sprintf("%d", time.Now().Unix())
	req := httptest.NewRequest(http.MethodPost,
		"/notify?signature=deadbeef&timestamp="+ts+"&nonce=n1",
		strings.NewReader(`<xml/>`))
	if _, err := ParseNotifyPlaintext(req, "mytoken"); err == nil {
		t.Error("expected signature error")
	}
}

// TestParseNotifyPlaintext_RejectsStaleTimestamp for plaintext mode.
func TestParseNotifyPlaintext_RejectsStaleTimestamp(t *testing.T) {
	token := "mytoken"
	stale := fmt.Sprintf("%d", time.Now().Add(-10*time.Minute).Unix())
	sig := computeOffiaccountSig(token, stale, "n1")
	req := httptest.NewRequest(http.MethodPost,
		"/notify?signature="+sig+"&timestamp="+stale+"&nonce=n1",
		strings.NewReader(`<xml/>`))
	if _, err := ParseNotifyPlaintext(req, token); err == nil {
		t.Error("expected stale-timestamp error")
	}
}

func TestVerifyServerTokenShim(t *testing.T) {
	// Validate the shim delegation path (offiaccount.VerifyServerToken
	// forwards to wxcrypto.VerifyServerToken). Algorithm correctness
	// is tested in utils/wxcrypto.
	if err := VerifyServerToken("tk", "deadbeef", "1700", "n1"); err == nil {
		t.Error("expected error for bad signature")
	}
}
