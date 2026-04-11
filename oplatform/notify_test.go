package oplatform

import (
	"context"
	"encoding/xml"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// helper — build an encrypted POST request for a given plaintext XML payload
func buildEncryptedReq(t *testing.T, c *Client, plain string, ts, nonce string) *http.Request {
	t.Helper()
	encrypted, err := c.crypto.Encrypt([]byte(plain))
	if err != nil {
		t.Fatal(err)
	}
	sig := c.crypto.Signature(ts, nonce, encrypted)
	body := `<xml><ToUserName><![CDATA[wxcomp]]></ToUserName><Encrypt><![CDATA[` + encrypted + `]]></Encrypt></xml>`
	req := httptest.NewRequest(http.MethodPost,
		"/oplatform/notify?msg_signature="+sig+"&timestamp="+ts+"&nonce="+nonce,
		strings.NewReader(body))
	return req
}

func TestParseNotify_VerifyTicket_AutoWritesStore(t *testing.T) {
	c, _ := NewClient(testConfig())
	plain := `<xml>
<AppId><![CDATA[wxcomp]]></AppId>
<CreateTime>1700000000</CreateTime>
<InfoType><![CDATA[component_verify_ticket]]></InfoType>
<ComponentVerifyTicket><![CDATA[TICKET_ABC]]></ComponentVerifyTicket>
</xml>`
	req := buildEncryptedReq(t, c, plain, "1700000000", "nonceA")

	notify, err := c.ParseNotify(req, nil)
	if err != nil {
		t.Fatal(err)
	}
	if notify.InfoType != "component_verify_ticket" {
		t.Errorf("info_type mismatch: %q", notify.InfoType)
	}
	if notify.ComponentVerifyTicket != "TICKET_ABC" {
		t.Errorf("ticket mismatch: %q", notify.ComponentVerifyTicket)
	}
	got, err := c.store.GetVerifyTicket(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got != "TICKET_ABC" {
		t.Errorf("store ticket mismatch: %q", got)
	}
}

func TestParseNotify_Authorized(t *testing.T) {
	c, _ := NewClient(testConfig())
	plain := `<xml>
<AppId><![CDATA[wxcomp]]></AppId>
<CreateTime>1700000001</CreateTime>
<InfoType><![CDATA[authorized]]></InfoType>
<AuthorizerAppid><![CDATA[wxAuthed]]></AuthorizerAppid>
<AuthorizationCode><![CDATA[AC_CODE]]></AuthorizationCode>
<AuthorizationCodeExpiredTime>1700003601</AuthorizationCodeExpiredTime>
<PreAuthCode><![CDATA[PRE]]></PreAuthCode>
</xml>`
	req := buildEncryptedReq(t, c, plain, "1700000001", "nonceB")

	notify, err := c.ParseNotify(req, nil)
	if err != nil {
		t.Fatal(err)
	}
	if notify.InfoType != "authorized" {
		t.Errorf("info_type mismatch: %q", notify.InfoType)
	}
	if notify.AuthorizerAppID != "wxAuthed" || notify.AuthorizationCode != "AC_CODE" {
		t.Errorf("unexpected: %+v", notify)
	}
	if notify.AuthorizationCodeExpiredTime != 1700003601 {
		t.Errorf("expire time mismatch: %d", notify.AuthorizationCodeExpiredTime)
	}
}

func TestParseNotify_BadSignature(t *testing.T) {
	c, _ := NewClient(testConfig())
	encrypted, _ := c.crypto.Encrypt([]byte(`<xml><InfoType>x</InfoType></xml>`))
	body := `<xml><Encrypt><![CDATA[` + encrypted + `]]></Encrypt></xml>`
	req := httptest.NewRequest(http.MethodPost,
		"/oplatform/notify?msg_signature=deadbeef&timestamp=1700&nonce=n1",
		strings.NewReader(body))
	if _, err := c.ParseNotify(req, nil); err == nil {
		t.Error("expected signature error")
	}
}

func TestParseNotify_RawBodyOverride(t *testing.T) {
	c, _ := NewClient(testConfig())
	plain := `<xml><AppId><![CDATA[wxcomp]]></AppId><InfoType><![CDATA[unauthorized]]></InfoType><AuthorizerAppid><![CDATA[wxAuthed]]></AuthorizerAppid></xml>`
	encrypted, _ := c.crypto.Encrypt([]byte(plain))
	sig := c.crypto.Signature("1700", "n1", encrypted)
	body := []byte(`<xml><Encrypt><![CDATA[` + encrypted + `]]></Encrypt></xml>`)
	req := httptest.NewRequest(http.MethodPost,
		"/oplatform/notify?msg_signature="+sig+"&timestamp=1700&nonce=n1",
		strings.NewReader("")) // empty body; we pass via rawBody
	notify, err := c.ParseNotify(req, body)
	if err != nil {
		t.Fatal(err)
	}
	if notify.InfoType != "unauthorized" || notify.AuthorizerAppID != "wxAuthed" {
		t.Errorf("unexpected: %+v", notify)
	}
}

func TestParseNotify_NotFoundSentinel(t *testing.T) {
	store := NewMemoryStore()
	_, err := store.GetVerifyTicket(context.Background())
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
	var env struct {
		XMLName xml.Name `xml:"xml"`
	}
	if err := xml.Unmarshal([]byte(`<xml/>`), &env); err != nil {
		t.Fatal(err)
	}
}
