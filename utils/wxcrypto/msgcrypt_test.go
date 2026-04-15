package wxcrypto

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"io"
	"sort"
	"strings"
	"testing"
)

func genEncodingAESKey(t *testing.T) string {
	t.Helper()
	raw := make([]byte, 32)
	for i := range raw {
		raw[i] = byte(i)
	}
	k := base64.StdEncoding.EncodeToString(raw)
	return strings.TrimSuffix(k, "=")
}

func TestMsgCrypto_EncryptDecryptRoundtrip(t *testing.T) {
	key := genEncodingAESKey(t)
	mc, err := New("tk", key, "wxappid")
	if err != nil {
		t.Fatal(err)
	}
	plain := []byte(`<xml><Content><![CDATA[hello]]></Content></xml>`)
	cipher, err := mc.Encrypt(plain)
	if err != nil {
		t.Fatal(err)
	}
	got, fromAppid, err := mc.Decrypt(cipher)
	if err != nil {
		t.Fatal(err)
	}
	if fromAppid != "wxappid" {
		t.Errorf("fromAppid: got %q", fromAppid)
	}
	if string(got) != string(plain) {
		t.Errorf("roundtrip mismatch: %q vs %q", got, plain)
	}
}

func TestMsgCrypto_SignatureDeterministic(t *testing.T) {
	mc, _ := New("tk", genEncodingAESKey(t), "wxappid")
	sig1 := mc.Signature("1700000000", "nonceA", "encX")
	sig2 := mc.Signature("1700000000", "nonceA", "encX")
	if sig1 != sig2 || sig1 == "" {
		t.Errorf("signature not deterministic: %s vs %s", sig1, sig2)
	}
	if !mc.VerifySignature(sig1, "1700000000", "nonceA", "encX") {
		t.Error("expected VerifySignature to succeed")
	}
	if mc.VerifySignature("bogus", "1700000000", "nonceA", "encX") {
		t.Error("expected VerifySignature to fail for bad sig")
	}
}

func TestMsgCrypto_BuildEncryptedReply(t *testing.T) {
	mc, _ := New("tk", genEncodingAESKey(t), "wxappid")
	reply, err := mc.BuildEncryptedReply([]byte(`<xml><MsgType>text</MsgType></xml>`), "1700000001", "nonce1")
	if err != nil {
		t.Fatal(err)
	}
	var env struct {
		Encrypt      string `xml:"Encrypt"`
		MsgSignature string `xml:"MsgSignature"`
		Nonce        string `xml:"Nonce"`
	}
	if err := xml.Unmarshal(reply, &env); err != nil {
		t.Fatal(err)
	}
	if !mc.VerifySignature(env.MsgSignature, "1700000001", "nonce1", env.Encrypt) {
		t.Error("reply signature should verify")
	}
	plain, _, err := mc.Decrypt(env.Encrypt)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(plain), "text") {
		t.Errorf("unexpected decrypted: %s", plain)
	}
}

func TestVerifyServerToken(t *testing.T) {
	sig := simpleSHA1("tk", "1700", "n1")
	if err := VerifyServerToken("tk", sig, "1700", "n1"); err != nil {
		t.Errorf("expected ok, got %v", err)
	}
	if err := VerifyServerToken("tk", "wrong", "1700", "n1"); err == nil {
		t.Error("expected error")
	}
}

func TestMsgCrypto_Decrypt_AppidMismatch(t *testing.T) {
	mcA, _ := New("tk", genEncodingAESKey(t), "appA")
	mcB, _ := New("tk", genEncodingAESKey(t), "appB")
	cipher, err := mcA.Encrypt([]byte(`<xml/>`))
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := mcB.Decrypt(cipher); err == nil {
		t.Error("expected appid mismatch error")
	}
}

func simpleSHA1(token, ts, nonce string) string {
	parts := []string{token, ts, nonce}
	sort.Strings(parts)
	h := sha1.New()
	_, _ = io.WriteString(h, strings.Join(parts, ""))
	return hex.EncodeToString(h.Sum(nil))
}

// Audit C4: after AES decryption, every failure mode must return the SAME opaque
// error so we don't leak distinguishable signals (padding oracle defense).
func TestDecrypt_OpaqueErrorOnAnyTampering(t *testing.T) {
	key := genEncodingAESKey(t)
	mc, err := New("tk", key, "wxappid")
	if err != nil {
		t.Fatal(err)
	}

	// Build a valid ciphertext we can mutate.
	good, err := mc.Encrypt([]byte(`<xml><Content><![CDATA[hello]]></Content></xml>`))
	if err != nil {
		t.Fatal(err)
	}
	rawGood, err := base64.StdEncoding.DecodeString(good)
	if err != nil {
		t.Fatal(err)
	}

	// Variant 1: flip the last byte (which lives in the padding region most of the time)
	flipped := make([]byte, len(rawGood))
	copy(flipped, rawGood)
	flipped[len(flipped)-1] ^= 0xFF

	// Variant 2: encrypted under a different appid — passes pad check, fails appid check
	mcWrong, err := New("tk", key, "different_appid")
	if err != nil {
		t.Fatal(err)
	}
	wrongAppid, err := mcWrong.Encrypt([]byte(`<xml/>`))
	if err != nil {
		t.Fatal(err)
	}

	// Variant 3: truncated by one byte (length not multiple of block size — pre-AES check)
	truncated := rawGood[:len(rawGood)-1]

	cases := []struct {
		name   string
		cipher string
	}{
		{"flipped padding", base64.StdEncoding.EncodeToString(flipped)},
		{"wrong appid", wrongAppid},
		{"truncated", base64.StdEncoding.EncodeToString(truncated)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := mc.Decrypt(tc.cipher)
			if err == nil {
				t.Fatalf("expected error")
			}
			if err.Error() != "wxcrypto: decrypt failed" {
				t.Errorf("expected opaque error, got %q", err.Error())
			}
		})
	}
}

// Audit: Decrypt must also reject malformed inputs (invalid base64 / empty)
// with the same opaque error as cryptographic failures. Anything else risks
// leaking a distinguishable signal back to an attacker.
func TestDecrypt_OpaqueErrorOnMalformedInput(t *testing.T) {
	mc, err := New("tk", genEncodingAESKey(t), "wxappid")
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name   string
		cipher string
	}{
		{"empty", ""},
		{"not base64", "@@@not-base64@@@"},
		{"valid base64 but too short", base64.StdEncoding.EncodeToString([]byte("short"))},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := mc.Decrypt(tc.cipher)
			if err == nil {
				t.Fatalf("expected error")
			}
			if err.Error() != "wxcrypto: decrypt failed" {
				t.Errorf("expected opaque error, got %q", err.Error())
			}
		})
	}
}
