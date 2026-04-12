package isv

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/godrealms/go-wechat-sdk/utils/wxcrypto"
)

// buildNotifyRequest 构造一个签名/加密正确的 HTTP POST,模拟企业微信的回调推送。
// innerXML 是明文的 innerBody(不含 Encrypt 信封)。
func buildNotifyRequest(t *testing.T, c *Client, innerXML string) *http.Request {
	t.Helper()
	timestamp := "1712900000"
	nonce := "nonce123"
	payload, err := c.crypto.BuildEncryptedReply([]byte(innerXML), timestamp, nonce)
	if err != nil {
		t.Fatal(err)
	}

	u := fmt.Sprintf("/cb?msg_signature=%s&timestamp=%s&nonce=%s",
		extractSignature(t, c.crypto, payload, timestamp, nonce), timestamp, nonce)
	req := httptest.NewRequest(http.MethodPost, u, bytes.NewReader(payload))
	return req
}

// extractSignature recomputes the signature from the envelope so the test URL matches.
// For the helper to work, we parse the envelope and re-sign with crypto.Signature.
func extractSignature(t *testing.T, cry *wxcrypto.MsgCrypto, envelope []byte, timestamp, nonce string) string {
	t.Helper()
	var env struct {
		XMLName xml.Name `xml:"xml"`
		Encrypt string   `xml:"Encrypt"`
	}
	if err := xml.Unmarshal(envelope, &env); err != nil {
		t.Fatal(err)
	}
	return cry.Signature(timestamp, nonce, env.Encrypt)
}

func TestParseNotify_SuiteTicket_Persists(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	// Remove the pre-seeded ticket so we can verify the store write.
	_ = c.store.PutSuiteTicket(context.Background(), "suite1", "")

	inner := `<xml><SuiteId><![CDATA[suite1]]></SuiteId><InfoType><![CDATA[suite_ticket]]></InfoType><SuiteTicket><![CDATA[NEW_TICKET]]></SuiteTicket></xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	stev, ok := ev.(*SuiteTicketEvent)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if stev.SuiteTicket != "NEW_TICKET" || stev.SuiteID != "suite1" {
		t.Errorf("event: %+v", stev)
	}
	// Store should have been updated
	got, _ := c.store.GetSuiteTicket(context.Background(), "suite1")
	if got != "NEW_TICKET" {
		t.Errorf("store not updated: %q", got)
	}
}

func TestParseNotify_CreateAuth(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml><SuiteId><![CDATA[suite1]]></SuiteId><InfoType><![CDATA[create_auth]]></InfoType><AuthCode><![CDATA[AC123]]></AuthCode></xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	cev, ok := ev.(*CreateAuthEvent)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if cev.AuthCode != "AC123" {
		t.Errorf("event: %+v", cev)
	}
}

func TestParseNotify_ChangeAuth(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml><SuiteId><![CDATA[suite1]]></SuiteId><InfoType><![CDATA[change_auth]]></InfoType><AuthCorpId><![CDATA[wxcorp1]]></AuthCorpId></xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	cev, ok := ev.(*ChangeAuthEvent)
	if !ok || cev.AuthCorpID != "wxcorp1" {
		t.Fatalf("event: %T %+v", ev, ev)
	}
}

func TestParseNotify_CancelAuth(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml><SuiteId><![CDATA[suite1]]></SuiteId><InfoType><![CDATA[cancel_auth]]></InfoType><AuthCorpId><![CDATA[wxcorp1]]></AuthCorpId></xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := ev.(*CancelAuthEvent); !ok {
		t.Fatalf("type: %T", ev)
	}
}

func TestParseNotify_ResetPermanentCode(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml><SuiteId><![CDATA[suite1]]></SuiteId><InfoType><![CDATA[reset_permanent_code]]></InfoType><AuthCorpId><![CDATA[wxcorp1]]></AuthCorpId></xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := ev.(*ResetPermanentCodeEvent); !ok {
		t.Fatalf("type: %T", ev)
	}
}

func TestParseNotify_ChangeContact(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml>
<SuiteId><![CDATA[suite1]]></SuiteId>
<InfoType><![CDATA[change_contact]]></InfoType>
<AuthCorpId><![CDATA[wxcorp1]]></AuthCorpId>
<ChangeType><![CDATA[update_user]]></ChangeType>
<UserID><![CDATA[u1]]></UserID>
<NewUserID><![CDATA[u1new]]></NewUserID>
<Name><![CDATA[Alice]]></Name>
</xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	cev, ok := ev.(*ChangeContactEvent)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if cev.ChangeType != "update_user" || cev.NewUserID != "u1new" {
		t.Errorf("event: %+v", cev)
	}
}

func TestParseNotify_ChangeExternalContact(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml><SuiteId><![CDATA[suite1]]></SuiteId><InfoType><![CDATA[change_external_contact]]></InfoType><AuthCorpId><![CDATA[wxcorp1]]></AuthCorpId><ChangeType><![CDATA[add_external_contact]]></ChangeType><UserID><![CDATA[u1]]></UserID><ExternalUserID><![CDATA[ex1]]></ExternalUserID></xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	cev, ok := ev.(*ChangeExternalContactEvent)
	if !ok || cev.ExternalUserID != "ex1" {
		t.Fatalf("event: %T %+v", ev, ev)
	}
}

func TestParseNotify_ShareAgentChange(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml><SuiteId><![CDATA[suite1]]></SuiteId><InfoType><![CDATA[share_agent_change]]></InfoType><AuthCorpId><![CDATA[wxcorp1]]></AuthCorpId><AgentID><![CDATA[1000001]]></AgentID></xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	sev, ok := ev.(*ShareAgentChangeEvent)
	if !ok || sev.AgentID != "1000001" {
		t.Fatalf("event: %T %+v", ev, ev)
	}
}

func TestParseNotify_ChangeAppAdmin(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml><SuiteId><![CDATA[suite1]]></SuiteId><InfoType><![CDATA[change_app_admin]]></InfoType><AuthCorpId><![CDATA[wxcorp1]]></AuthCorpId><UserID><![CDATA[u_admin]]></UserID><IsAdmin>1</IsAdmin></xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	aev, ok := ev.(*ChangeAppAdminEvent)
	if !ok || !aev.IsAdmin || aev.UserID != "u_admin" {
		t.Fatalf("event: %T %+v", ev, ev)
	}
}

func TestParseNotify_UnknownInfoType_ReturnsRawEvent(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml><SuiteId><![CDATA[suite1]]></SuiteId><InfoType><![CDATA[brand_new_event]]></InfoType><Foo><![CDATA[bar]]></Foo></xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	rev, ok := ev.(*RawEvent)
	if !ok || rev.InfoType != "brand_new_event" {
		t.Fatalf("event: %T %+v", ev, ev)
	}
	if !strings.Contains(rev.RawXML, "<Foo>") {
		t.Errorf("RawXML missing Foo: %q", rev.RawXML)
	}
}

func TestParseNotify_BadSignature(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml><SuiteId><![CDATA[suite1]]></SuiteId><InfoType><![CDATA[suite_ticket]]></InfoType><SuiteTicket><![CDATA[X]]></SuiteTicket></xml>`
	req := buildNotifyRequest(t, c, inner)

	// Tamper with the signature
	q := req.URL.Query()
	q.Set("msg_signature", "deadbeef")
	req.URL.RawQuery = q.Encode()

	_, err := c.ParseNotify(req)
	if err == nil || !strings.Contains(err.Error(), "signature") {
		t.Fatalf("want signature error, got %v", err)
	}
}
