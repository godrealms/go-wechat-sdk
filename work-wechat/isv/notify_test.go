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

// ── Contact change: user tests ──────────────────────────────────────

func TestParseNotify_ContactCreateUser(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml>
<SuiteId><![CDATA[suite1]]></SuiteId>
<InfoType><![CDATA[change_contact]]></InfoType>
<AuthCorpId><![CDATA[wxcorp1]]></AuthCorpId>
<ChangeType><![CDATA[create_user]]></ChangeType>
<UserID><![CDATA[u1]]></UserID>
<Name><![CDATA[Alice]]></Name>
<Department><![CDATA[1,2]]></Department>
<Mobile><![CDATA[13800138000]]></Mobile>
<Email><![CDATA[alice@example.com]]></Email>
<Position><![CDATA[Engineer]]></Position>
<Gender>1</Gender>
<Avatar><![CDATA[https://img/a.png]]></Avatar>
<Status>1</Status>
<IsLeaderInDept><![CDATA[0,1]]></IsLeaderInDept>
</xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	cev, ok := ev.(*ContactCreateUserEvent)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if cev.ChangeType != "create_user" {
		t.Errorf("ChangeType = %q", cev.ChangeType)
	}
	if cev.AuthCorpID != "wxcorp1" {
		t.Errorf("AuthCorpID = %q", cev.AuthCorpID)
	}
	if cev.UserID != "u1" {
		t.Errorf("UserID = %q", cev.UserID)
	}
	if cev.Name != "Alice" {
		t.Errorf("Name = %q", cev.Name)
	}
	if cev.Department != "1,2" {
		t.Errorf("Department = %q", cev.Department)
	}
	if cev.Mobile != "13800138000" {
		t.Errorf("Mobile = %q", cev.Mobile)
	}
	if cev.Email != "alice@example.com" {
		t.Errorf("Email = %q", cev.Email)
	}
	if cev.Position != "Engineer" {
		t.Errorf("Position = %q", cev.Position)
	}
	if cev.Gender != 1 {
		t.Errorf("Gender = %d", cev.Gender)
	}
	if cev.Avatar != "https://img/a.png" {
		t.Errorf("Avatar = %q", cev.Avatar)
	}
	if cev.Status != 1 {
		t.Errorf("Status = %d", cev.Status)
	}
	if cev.IsLeaderInDept != "0,1" {
		t.Errorf("IsLeaderInDept = %q", cev.IsLeaderInDept)
	}
}

func TestParseNotify_ContactUpdateUser(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml>
<SuiteId><![CDATA[suite1]]></SuiteId>
<InfoType><![CDATA[change_contact]]></InfoType>
<AuthCorpId><![CDATA[wxcorp1]]></AuthCorpId>
<ChangeType><![CDATA[update_user]]></ChangeType>
<UserID><![CDATA[u1]]></UserID>
<NewUserID><![CDATA[u1new]]></NewUserID>
<Name><![CDATA[Alice Updated]]></Name>
<Department><![CDATA[3]]></Department>
<Mobile><![CDATA[13900139000]]></Mobile>
<Email><![CDATA[alice2@example.com]]></Email>
<Position><![CDATA[Senior Engineer]]></Position>
<Gender>2</Gender>
<Avatar><![CDATA[https://img/b.png]]></Avatar>
<Status>4</Status>
<IsLeaderInDept><![CDATA[1]]></IsLeaderInDept>
</xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	cev, ok := ev.(*ContactUpdateUserEvent)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if cev.ChangeType != "update_user" {
		t.Errorf("ChangeType = %q", cev.ChangeType)
	}
	if cev.NewUserID != "u1new" {
		t.Errorf("NewUserID = %q", cev.NewUserID)
	}
	if cev.UserID != "u1" {
		t.Errorf("UserID = %q", cev.UserID)
	}
	if cev.Name != "Alice Updated" {
		t.Errorf("Name = %q", cev.Name)
	}
}

func TestParseNotify_ContactDeleteUser(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml>
<SuiteId><![CDATA[suite1]]></SuiteId>
<InfoType><![CDATA[change_contact]]></InfoType>
<AuthCorpId><![CDATA[wxcorp1]]></AuthCorpId>
<ChangeType><![CDATA[delete_user]]></ChangeType>
<UserID><![CDATA[u_gone]]></UserID>
</xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	cev, ok := ev.(*ContactDeleteUserEvent)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if cev.ChangeType != "delete_user" {
		t.Errorf("ChangeType = %q", cev.ChangeType)
	}
	if cev.UserID != "u_gone" {
		t.Errorf("UserID = %q", cev.UserID)
	}
	if cev.AuthCorpID != "wxcorp1" {
		t.Errorf("AuthCorpID = %q", cev.AuthCorpID)
	}
}

// ── Contact change: department tests ─────────────────────────────────

func TestParseNotify_ContactCreateParty(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml>
<SuiteId><![CDATA[suite1]]></SuiteId>
<InfoType><![CDATA[change_contact]]></InfoType>
<AuthCorpId><![CDATA[wxcorp1]]></AuthCorpId>
<ChangeType><![CDATA[create_party]]></ChangeType>
<Id>10</Id>
<Name><![CDATA[Engineering]]></Name>
<ParentId>1</ParentId>
<Order>5</Order>
</xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	cev, ok := ev.(*ContactCreatePartyEvent)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if cev.ChangeType != "create_party" {
		t.Errorf("ChangeType = %q", cev.ChangeType)
	}
	if cev.ID != 10 {
		t.Errorf("ID = %d", cev.ID)
	}
	if cev.Name != "Engineering" {
		t.Errorf("Name = %q", cev.Name)
	}
	if cev.ParentID != 1 {
		t.Errorf("ParentID = %d", cev.ParentID)
	}
	if cev.Order != 5 {
		t.Errorf("Order = %d", cev.Order)
	}
}

func TestParseNotify_ContactUpdateParty(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml>
<SuiteId><![CDATA[suite1]]></SuiteId>
<InfoType><![CDATA[change_contact]]></InfoType>
<AuthCorpId><![CDATA[wxcorp1]]></AuthCorpId>
<ChangeType><![CDATA[update_party]]></ChangeType>
<Id>10</Id>
<Name><![CDATA[Engineering v2]]></Name>
<ParentId>2</ParentId>
</xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	cev, ok := ev.(*ContactUpdatePartyEvent)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if cev.ChangeType != "update_party" {
		t.Errorf("ChangeType = %q", cev.ChangeType)
	}
	if cev.ID != 10 {
		t.Errorf("ID = %d", cev.ID)
	}
	if cev.Name != "Engineering v2" {
		t.Errorf("Name = %q", cev.Name)
	}
	if cev.ParentID != 2 {
		t.Errorf("ParentID = %d", cev.ParentID)
	}
}

func TestParseNotify_ContactDeleteParty(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml>
<SuiteId><![CDATA[suite1]]></SuiteId>
<InfoType><![CDATA[change_contact]]></InfoType>
<AuthCorpId><![CDATA[wxcorp1]]></AuthCorpId>
<ChangeType><![CDATA[delete_party]]></ChangeType>
<Id>10</Id>
</xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	cev, ok := ev.(*ContactDeletePartyEvent)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if cev.ChangeType != "delete_party" {
		t.Errorf("ChangeType = %q", cev.ChangeType)
	}
	if cev.ID != 10 {
		t.Errorf("ID = %d", cev.ID)
	}
	if cev.AuthCorpID != "wxcorp1" {
		t.Errorf("AuthCorpID = %q", cev.AuthCorpID)
	}
}

// ── Contact change: tag test ─────────────────────────────────────────

func TestParseNotify_ContactUpdateTag(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml>
<SuiteId><![CDATA[suite1]]></SuiteId>
<InfoType><![CDATA[change_contact]]></InfoType>
<AuthCorpId><![CDATA[wxcorp1]]></AuthCorpId>
<ChangeType><![CDATA[update_tag]]></ChangeType>
<TagId>100</TagId>
<AddUserItems><![CDATA[u1,u2]]></AddUserItems>
<DelUserItems><![CDATA[u3]]></DelUserItems>
<AddPartyItems><![CDATA[10,20]]></AddPartyItems>
<DelPartyItems><![CDATA[30]]></DelPartyItems>
</xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	cev, ok := ev.(*ContactUpdateTagEvent)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if cev.ChangeType != "update_tag" {
		t.Errorf("ChangeType = %q", cev.ChangeType)
	}
	if cev.TagID != 100 {
		t.Errorf("TagID = %d", cev.TagID)
	}
	if cev.AddUserItems != "u1,u2" {
		t.Errorf("AddUserItems = %q", cev.AddUserItems)
	}
	if cev.DelUserItems != "u3" {
		t.Errorf("DelUserItems = %q", cev.DelUserItems)
	}
	if cev.AddPartyItems != "10,20" {
		t.Errorf("AddPartyItems = %q", cev.AddPartyItems)
	}
	if cev.DelPartyItems != "30" {
		t.Errorf("DelPartyItems = %q", cev.DelPartyItems)
	}
}

// ── External contact change tests ────────────────────────────────────

func TestParseNotify_ExtContactAdd(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml>
<SuiteId><![CDATA[suite1]]></SuiteId>
<InfoType><![CDATA[change_external_contact]]></InfoType>
<AuthCorpId><![CDATA[wxcorp1]]></AuthCorpId>
<ChangeType><![CDATA[add_external_contact]]></ChangeType>
<UserID><![CDATA[u1]]></UserID>
<ExternalUserID><![CDATA[ex1]]></ExternalUserID>
<State><![CDATA[mystate]]></State>
<WelcomeCode><![CDATA[wc123]]></WelcomeCode>
</xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	cev, ok := ev.(*ExtContactAddEvent)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if cev.ChangeType != "add_external_contact" {
		t.Errorf("ChangeType = %q", cev.ChangeType)
	}
	if cev.UserID != "u1" {
		t.Errorf("UserID = %q", cev.UserID)
	}
	if cev.ExternalUserID != "ex1" {
		t.Errorf("ExternalUserID = %q", cev.ExternalUserID)
	}
	if cev.State != "mystate" {
		t.Errorf("State = %q", cev.State)
	}
	if cev.WelcomeCode != "wc123" {
		t.Errorf("WelcomeCode = %q", cev.WelcomeCode)
	}
}

func TestParseNotify_ExtContactEdit(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml>
<SuiteId><![CDATA[suite1]]></SuiteId>
<InfoType><![CDATA[change_external_contact]]></InfoType>
<AuthCorpId><![CDATA[wxcorp1]]></AuthCorpId>
<ChangeType><![CDATA[edit_external_contact]]></ChangeType>
<UserID><![CDATA[u1]]></UserID>
<ExternalUserID><![CDATA[ex1]]></ExternalUserID>
</xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	cev, ok := ev.(*ExtContactEditEvent)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if cev.ChangeType != "edit_external_contact" {
		t.Errorf("ChangeType = %q", cev.ChangeType)
	}
	if cev.UserID != "u1" {
		t.Errorf("UserID = %q", cev.UserID)
	}
	if cev.ExternalUserID != "ex1" {
		t.Errorf("ExternalUserID = %q", cev.ExternalUserID)
	}
}

func TestParseNotify_ExtContactDel(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml>
<SuiteId><![CDATA[suite1]]></SuiteId>
<InfoType><![CDATA[change_external_contact]]></InfoType>
<AuthCorpId><![CDATA[wxcorp1]]></AuthCorpId>
<ChangeType><![CDATA[del_external_contact]]></ChangeType>
<UserID><![CDATA[u1]]></UserID>
<ExternalUserID><![CDATA[ex_gone]]></ExternalUserID>
</xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	cev, ok := ev.(*ExtContactDelEvent)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if cev.ChangeType != "del_external_contact" {
		t.Errorf("ChangeType = %q", cev.ChangeType)
	}
	if cev.ExternalUserID != "ex_gone" {
		t.Errorf("ExternalUserID = %q", cev.ExternalUserID)
	}
}

func TestParseNotify_ExtContactDelFollow(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml>
<SuiteId><![CDATA[suite1]]></SuiteId>
<InfoType><![CDATA[change_external_contact]]></InfoType>
<AuthCorpId><![CDATA[wxcorp1]]></AuthCorpId>
<ChangeType><![CDATA[del_follow_user]]></ChangeType>
<UserID><![CDATA[u1]]></UserID>
<ExternalUserID><![CDATA[ex1]]></ExternalUserID>
</xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	cev, ok := ev.(*ExtContactDelFollowEvent)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if cev.ChangeType != "del_follow_user" {
		t.Errorf("ChangeType = %q", cev.ChangeType)
	}
	if cev.UserID != "u1" {
		t.Errorf("UserID = %q", cev.UserID)
	}
	if cev.ExternalUserID != "ex1" {
		t.Errorf("ExternalUserID = %q", cev.ExternalUserID)
	}
}

func TestParseNotify_ExtContactAddHalf(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml>
<SuiteId><![CDATA[suite1]]></SuiteId>
<InfoType><![CDATA[change_external_contact]]></InfoType>
<AuthCorpId><![CDATA[wxcorp1]]></AuthCorpId>
<ChangeType><![CDATA[add_half_external_contact]]></ChangeType>
<UserID><![CDATA[u1]]></UserID>
<ExternalUserID><![CDATA[ex_half]]></ExternalUserID>
<State><![CDATA[halfstate]]></State>
<WelcomeCode><![CDATA[wc_half]]></WelcomeCode>
</xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	cev, ok := ev.(*ExtContactAddHalfEvent)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if cev.ChangeType != "add_half_external_contact" {
		t.Errorf("ChangeType = %q", cev.ChangeType)
	}
	if cev.ExternalUserID != "ex_half" {
		t.Errorf("ExternalUserID = %q", cev.ExternalUserID)
	}
	if cev.State != "halfstate" {
		t.Errorf("State = %q", cev.State)
	}
	if cev.WelcomeCode != "wc_half" {
		t.Errorf("WelcomeCode = %q", cev.WelcomeCode)
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
