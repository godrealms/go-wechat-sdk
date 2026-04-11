package offiaccount

import (
	"encoding/base64"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
	sig := mc.Signature("1700", "n1", encrypted)
	bodyXML := `<xml><ToUserName>gh_x</ToUserName><Encrypt><![CDATA[` + encrypted + `]]></Encrypt></xml>`
	req := httptest.NewRequest(http.MethodPost,
		"/notify?msg_signature="+sig+"&timestamp=1700&nonce=n1",
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
	req := httptest.NewRequest(http.MethodPost,
		"/notify?msg_signature=deadbeef&timestamp=1700&nonce=n1",
		strings.NewReader(bodyXML))
	if _, err := ParseNotify(req, mc); err == nil {
		t.Error("expected signature error")
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
