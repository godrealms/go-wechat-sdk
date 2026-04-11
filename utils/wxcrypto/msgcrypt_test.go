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
