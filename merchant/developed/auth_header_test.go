package pay

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"testing"
)

// TestAuthorizationHeader_GoldenFormat verifies the WECHATPAY2-SHA256-RSA2048
// header has the exact field order and quoting WeChat documents:
//
//	WECHATPAY2-SHA256-RSA2048 mchid="...",nonce_str="...",signature="...",timestamp="...",serial_no="..."
//
// Field ORDER matters because some downstream parsers (and the canonical form
// in our own platform_cert verifyResponseSignature) treat it as a structured
// string. A drift here would break signing in a way that's hard to spot in
// unit tests against our own server fixtures (where we'd verify with the
// same code that signs).
func TestAuthorizationHeader_GoldenFormat(t *testing.T) {
	c, _ := buildClient(t)

	got, err := c.authorizationHeader("POST", "/v3/pay/transactions/jsapi", `{"x":1}`, "NONCE_ABC", 1700000000)
	if err != nil {
		t.Fatal(err)
	}

	// Prefix must be exactly the WeChat-defined scheme.
	if !strings.HasPrefix(got, "WECHATPAY2-SHA256-RSA2048 ") {
		t.Errorf("missing prefix, got %q", got)
	}

	// All 5 fields, in WeChat-documented order, no extras.
	pattern := regexp.MustCompile(`^WECHATPAY2-SHA256-RSA2048 mchid="([^"]*)",nonce_str="([^"]*)",signature="([^"]*)",timestamp="([^"]*)",serial_no="([^"]*)"$`)
	m := pattern.FindStringSubmatch(got)
	if m == nil {
		t.Fatalf("header does not match documented field order/quoting:\n%s", got)
	}

	mchid, nonce, sig, ts, serial := m[1], m[2], m[3], m[4], m[5]
	if mchid != "1900000001" {
		t.Errorf("mchid = %q, want 1900000001", mchid)
	}
	if nonce != "NONCE_ABC" {
		t.Errorf("nonce_str = %q, want NONCE_ABC", nonce)
	}
	if ts != "1700000000" {
		t.Errorf("timestamp = %q, want 1700000000", ts)
	}
	if serial != "TESTSERIAL" {
		t.Errorf("serial_no = %q, want TESTSERIAL", serial)
	}
	if sig == "" {
		t.Errorf("signature is empty")
	}
}

// TestAuthorizationHeader_SignatureRoundTrip rebuilds the canonical request
// (METHOD\nPATH\nTS\nNONCE\nBODY\n) outside the SDK and verifies the
// header's `signature` field decodes to a valid RSA-PKCS1v15 SHA-256 signature
// of that canonical string. This is the contract WeChat verifies on their
// side; if our header diverges from this, all live API calls would 401.
func TestAuthorizationHeader_SignatureRoundTrip(t *testing.T) {
	c, key := buildClient(t)

	method, path, body, nonce := "POST", "/v3/refund/domestic/refunds", `{"out_refund_no":"R1"}`, "N42"
	var ts int64 = 1700001234

	got, err := c.authorizationHeader(method, path, body, nonce, ts)
	if err != nil {
		t.Fatal(err)
	}

	// Extract just the base64 signature.
	pattern := regexp.MustCompile(`signature="([^"]+)"`)
	m := pattern.FindStringSubmatch(got)
	if m == nil {
		t.Fatalf("could not extract signature from %q", got)
	}
	sigBytes, err := base64.StdEncoding.DecodeString(m[1])
	if err != nil {
		t.Fatalf("signature is not base64: %v", err)
	}

	// Reconstruct the canonical request the way WeChat does.
	canonical := fmt.Sprintf("%s\n%s\n%d\n%s\n%s\n", method, path, ts, nonce, body)
	hash := sha256.Sum256([]byte(canonical))

	if err := rsa.VerifyPKCS1v15(&key.PublicKey, crypto.SHA256, hash[:], sigBytes); err != nil {
		t.Errorf("signature did not verify against canonical request: %v", err)
	}
}

// TestAuthorizationHeader_DifferentBodiesProduceDifferentSignatures guards
// against an accidental signing of a fixed string (e.g. if the canonical
// request were not rebuilt per call).
func TestAuthorizationHeader_DifferentBodiesProduceDifferentSignatures(t *testing.T) {
	c, _ := buildClient(t)

	h1, err1 := c.authorizationHeader("POST", "/v3/pay/transactions/jsapi", `{"a":1}`, "N1", 1700000000)
	h2, err2 := c.authorizationHeader("POST", "/v3/pay/transactions/jsapi", `{"a":2}`, "N1", 1700000000)
	if err1 != nil || err2 != nil {
		t.Fatal(err1, err2)
	}
	if h1 == h2 {
		t.Error("signatures should differ when body differs")
	}
}
