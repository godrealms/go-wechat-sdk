package isv

import (
	"bytes"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
)

// ── Message type tests ──────────────────────────────────────────────

func TestParseDataNotify_TextMsg(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml>
<ToUserName><![CDATA[wxcorp1]]></ToUserName>
<AgentID>1000001</AgentID>
<MsgType><![CDATA[text]]></MsgType>
<MsgId>12345</MsgId>
<Content><![CDATA[Hello World]]></Content>
</xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseDataNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	msg, ok := ev.(*DataTextMsg)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if msg.ToUserName != "wxcorp1" {
		t.Errorf("ToUserName = %q", msg.ToUserName)
	}
	if msg.AgentID != 1000001 {
		t.Errorf("AgentID = %d", msg.AgentID)
	}
	if msg.MsgID != 12345 {
		t.Errorf("MsgID = %d", msg.MsgID)
	}
	if msg.Content != "Hello World" {
		t.Errorf("Content = %q", msg.Content)
	}
}

func TestParseDataNotify_ImageMsg(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml>
<ToUserName><![CDATA[wxcorp1]]></ToUserName>
<AgentID>1000001</AgentID>
<MsgType><![CDATA[image]]></MsgType>
<MsgId>12346</MsgId>
<PicUrl><![CDATA[https://img/pic.jpg]]></PicUrl>
<MediaId><![CDATA[media_img_001]]></MediaId>
</xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseDataNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	msg, ok := ev.(*DataImageMsg)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if msg.MsgID != 12346 {
		t.Errorf("MsgID = %d", msg.MsgID)
	}
	if msg.PicURL != "https://img/pic.jpg" {
		t.Errorf("PicURL = %q", msg.PicURL)
	}
	if msg.MediaID != "media_img_001" {
		t.Errorf("MediaID = %q", msg.MediaID)
	}
}

func TestParseDataNotify_VoiceMsg(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml>
<ToUserName><![CDATA[wxcorp1]]></ToUserName>
<AgentID>1000001</AgentID>
<MsgType><![CDATA[voice]]></MsgType>
<MsgId>12347</MsgId>
<MediaId><![CDATA[media_voice_001]]></MediaId>
<Format><![CDATA[amr]]></Format>
</xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseDataNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	msg, ok := ev.(*DataVoiceMsg)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if msg.MediaID != "media_voice_001" {
		t.Errorf("MediaID = %q", msg.MediaID)
	}
	if msg.Format != "amr" {
		t.Errorf("Format = %q", msg.Format)
	}
}

func TestParseDataNotify_VideoMsg(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml>
<ToUserName><![CDATA[wxcorp1]]></ToUserName>
<AgentID>1000001</AgentID>
<MsgType><![CDATA[video]]></MsgType>
<MsgId>12348</MsgId>
<MediaId><![CDATA[media_video_001]]></MediaId>
<ThumbMediaId><![CDATA[thumb_001]]></ThumbMediaId>
</xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseDataNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	msg, ok := ev.(*DataVideoMsg)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if msg.MediaID != "media_video_001" {
		t.Errorf("MediaID = %q", msg.MediaID)
	}
	if msg.ThumbMediaID != "thumb_001" {
		t.Errorf("ThumbMediaID = %q", msg.ThumbMediaID)
	}
}

func TestParseDataNotify_LocationMsg(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml>
<ToUserName><![CDATA[wxcorp1]]></ToUserName>
<AgentID>1000001</AgentID>
<MsgType><![CDATA[location]]></MsgType>
<MsgId>12349</MsgId>
<Location_X>39.916527</Location_X>
<Location_Y>116.397128</Location_Y>
<Scale>15</Scale>
<Label><![CDATA[Tiananmen Square]]></Label>
</xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseDataNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	msg, ok := ev.(*DataLocationMsg)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if msg.Lat != 39.916527 {
		t.Errorf("Lat = %f", msg.Lat)
	}
	if msg.Lng != 116.397128 {
		t.Errorf("Lng = %f", msg.Lng)
	}
	if msg.Scale != 15 {
		t.Errorf("Scale = %d", msg.Scale)
	}
	if msg.Label != "Tiananmen Square" {
		t.Errorf("Label = %q", msg.Label)
	}
}

func TestParseDataNotify_LinkMsg(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml>
<ToUserName><![CDATA[wxcorp1]]></ToUserName>
<AgentID>1000001</AgentID>
<MsgType><![CDATA[link]]></MsgType>
<MsgId>12350</MsgId>
<Title><![CDATA[Go SDK]]></Title>
<Description><![CDATA[WeChat SDK for Go]]></Description>
<Url><![CDATA[https://example.com/sdk]]></Url>
</xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseDataNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	msg, ok := ev.(*DataLinkMsg)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if msg.Title != "Go SDK" {
		t.Errorf("Title = %q", msg.Title)
	}
	if msg.Description != "WeChat SDK for Go" {
		t.Errorf("Description = %q", msg.Description)
	}
	if msg.URL != "https://example.com/sdk" {
		t.Errorf("URL = %q", msg.URL)
	}
}

// ── Event type tests ────────────────────────────────────────────────

func TestParseDataNotify_EnterAgent(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml>
<ToUserName><![CDATA[wxcorp1]]></ToUserName>
<AgentID>1000001</AgentID>
<MsgType><![CDATA[event]]></MsgType>
<Event><![CDATA[enter_agent]]></Event>
<EventKey><![CDATA[]]></EventKey>
</xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseDataNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	_, ok := ev.(*DataEnterAgentEvent)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
}

func TestParseDataNotify_MenuClick(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml>
<ToUserName><![CDATA[wxcorp1]]></ToUserName>
<AgentID>1000001</AgentID>
<MsgType><![CDATA[event]]></MsgType>
<Event><![CDATA[click]]></Event>
<EventKey><![CDATA[btn_report]]></EventKey>
</xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseDataNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	cev, ok := ev.(*DataMenuClickEvent)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if cev.EventKey != "btn_report" {
		t.Errorf("EventKey = %q", cev.EventKey)
	}
}

func TestParseDataNotify_MenuView(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml>
<ToUserName><![CDATA[wxcorp1]]></ToUserName>
<AgentID>1000001</AgentID>
<MsgType><![CDATA[event]]></MsgType>
<Event><![CDATA[view]]></Event>
<EventKey><![CDATA[menu_link]]></EventKey>
<Url><![CDATA[https://example.com/page]]></Url>
</xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseDataNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	cev, ok := ev.(*DataMenuViewEvent)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if cev.EventKey != "menu_link" {
		t.Errorf("EventKey = %q", cev.EventKey)
	}
	if cev.URL != "https://example.com/page" {
		t.Errorf("URL = %q", cev.URL)
	}
}

func TestParseDataNotify_ApprovalChange(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml>
<ToUserName><![CDATA[wxcorp1]]></ToUserName>
<AgentID>1000001</AgentID>
<MsgType><![CDATA[event]]></MsgType>
<Event><![CDATA[open_approval_change]]></Event>
<ApprovalInfo>
<ThirdNo><![CDATA[T001]]></ThirdNo>
<OpenSpName><![CDATA[Leave]]></OpenSpName>
<OpenSpStatus>1</OpenSpStatus>
<ApplyTime>1712900000</ApplyTime>
<ApplyUserid><![CDATA[u_apply]]></ApplyUserid>
<ApplyUserName><![CDATA[Bob]]></ApplyUserName>
<ApprovalNodes>
<ApprovalNode>
<NodeStatus>1</NodeStatus>
<NodeAttr>1</NodeAttr>
<Items>
<Item>
<ItemName><![CDATA[Manager]]></ItemName>
<ItemUserid><![CDATA[u_mgr]]></ItemUserid>
<ItemStatus>1</ItemStatus>
<ItemSpeech><![CDATA[]]></ItemSpeech>
<ItemOpTime>0</ItemOpTime>
</Item>
</Items>
</ApprovalNode>
</ApprovalNodes>
<NotifyNodes>
<NotifyNode>
<ItemName><![CDATA[HR]]></ItemName>
<ItemUserid><![CDATA[u_hr]]></ItemUserid>
</NotifyNode>
</NotifyNodes>
</ApprovalInfo>
</xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseDataNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	cev, ok := ev.(*DataApprovalChangeEvent)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	ai := cev.ApprovalInfo
	if ai.ThirdNo != "T001" {
		t.Errorf("ThirdNo = %q", ai.ThirdNo)
	}
	if ai.OpenSpName != "Leave" {
		t.Errorf("OpenSpName = %q", ai.OpenSpName)
	}
	if ai.OpenSpStatus != 1 {
		t.Errorf("OpenSpStatus = %d", ai.OpenSpStatus)
	}
	if ai.ApplyUserID != "u_apply" {
		t.Errorf("ApplyUserID = %q", ai.ApplyUserID)
	}
	if len(ai.ApprovalNodes) != 1 {
		t.Fatalf("ApprovalNodes len = %d", len(ai.ApprovalNodes))
	}
	node := ai.ApprovalNodes[0]
	if len(node.Items) != 1 {
		t.Fatalf("Items len = %d", len(node.Items))
	}
	item := node.Items[0]
	if item.ItemName != "Manager" {
		t.Errorf("ItemName = %q", item.ItemName)
	}
	if item.ItemUserID != "u_mgr" {
		t.Errorf("ItemUserID = %q", item.ItemUserID)
	}
	if item.ItemStatus != 1 {
		t.Errorf("ItemStatus = %d", item.ItemStatus)
	}
	if len(ai.NotifyNodes) != 1 {
		t.Fatalf("NotifyNodes len = %d", len(ai.NotifyNodes))
	}
	nn := ai.NotifyNodes[0]
	if nn.ItemName != "HR" {
		t.Errorf("NotifyNode ItemName = %q", nn.ItemName)
	}
	if nn.ItemUserID != "u_hr" {
		t.Errorf("NotifyNode ItemUserID = %q", nn.ItemUserID)
	}
}

func TestParseDataNotify_BatchJobResult(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml>
<ToUserName><![CDATA[wxcorp1]]></ToUserName>
<AgentID>1000001</AgentID>
<MsgType><![CDATA[event]]></MsgType>
<Event><![CDATA[batch_job_result]]></Event>
<BatchJob>
<JobId><![CDATA[job_123]]></JobId>
<JobType><![CDATA[sync_user]]></JobType>
<ErrCode>0</ErrCode>
<ErrMsg><![CDATA[ok]]></ErrMsg>
</BatchJob>
</xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseDataNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	cev, ok := ev.(*DataBatchJobResultEvent)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if cev.BatchJob.JobID != "job_123" {
		t.Errorf("JobID = %q", cev.BatchJob.JobID)
	}
	if cev.BatchJob.JobType != "sync_user" {
		t.Errorf("JobType = %q", cev.BatchJob.JobType)
	}
	if cev.BatchJob.ErrCode != 0 {
		t.Errorf("ErrCode = %d", cev.BatchJob.ErrCode)
	}
}

func TestParseDataNotify_ExtContactChange(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml>
<ToUserName><![CDATA[wxcorp1]]></ToUserName>
<AgentID>1000001</AgentID>
<MsgType><![CDATA[event]]></MsgType>
<Event><![CDATA[change_external_contact]]></Event>
<ChangeType><![CDATA[add_external_contact]]></ChangeType>
<UserID><![CDATA[u_sales]]></UserID>
<ExternalUserID><![CDATA[ext_cust_001]]></ExternalUserID>
<State><![CDATA[channel_a]]></State>
<WelcomeCode><![CDATA[wc_abc]]></WelcomeCode>
</xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseDataNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	cev, ok := ev.(*DataExtContactChangeEvent)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if cev.ChangeType != "add_external_contact" {
		t.Errorf("ChangeType = %q", cev.ChangeType)
	}
	if cev.UserID != "u_sales" {
		t.Errorf("UserID = %q", cev.UserID)
	}
	if cev.ExternalUserID != "ext_cust_001" {
		t.Errorf("ExternalUserID = %q", cev.ExternalUserID)
	}
	if cev.State != "channel_a" {
		t.Errorf("State = %q", cev.State)
	}
	if cev.WelcomeCode != "wc_abc" {
		t.Errorf("WelcomeCode = %q", cev.WelcomeCode)
	}
}

// ── Fallback tests ──────────────────────────────────────────────────

func TestParseDataNotify_RawMsg(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml>
<ToUserName><![CDATA[wxcorp1]]></ToUserName>
<AgentID>1000001</AgentID>
<MsgType><![CDATA[shortvideo]]></MsgType>
<MsgId>99999</MsgId>
</xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseDataNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	raw, ok := ev.(*DataRawMsg)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if raw.MsgType != "shortvideo" {
		t.Errorf("MsgType = %q", raw.MsgType)
	}
	if !strings.Contains(raw.RawXML, "shortvideo") {
		t.Errorf("RawXML missing MsgType: %q", raw.RawXML)
	}
}

func TestParseDataNotify_RawEvent(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml>
<ToUserName><![CDATA[wxcorp1]]></ToUserName>
<AgentID>1000001</AgentID>
<MsgType><![CDATA[event]]></MsgType>
<Event><![CDATA[future_event_type]]></Event>
</xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseDataNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	raw, ok := ev.(*DataRawEvent)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if raw.Event != "future_event_type" {
		t.Errorf("Event = %q", raw.Event)
	}
	if !strings.Contains(raw.RawXML, "future_event_type") {
		t.Errorf("RawXML missing Event: %q", raw.RawXML)
	}
}

// ── Error tests ─────────────────────────────────────────────────────

func TestParseDataNotify_InvalidSignature(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml>
<ToUserName><![CDATA[wxcorp1]]></ToUserName>
<AgentID>1000001</AgentID>
<MsgType><![CDATA[text]]></MsgType>
<MsgId>12345</MsgId>
<Content><![CDATA[Hello]]></Content>
</xml>`
	req := buildNotifyRequest(t, c, inner)

	// Tamper with the signature
	q := req.URL.Query()
	q.Set("msg_signature", "deadbeef")
	req.URL.RawQuery = q.Encode()

	_, err := c.ParseDataNotify(req)
	if err == nil || !strings.Contains(err.Error(), "signature") {
		t.Fatalf("want signature error, got %v", err)
	}
}

func TestParseDataNotify_InvalidXML(t *testing.T) {
	c := newTestISVClient(t, "http://unused")

	// Build a valid encrypted request but with non-XML content.
	// We encrypt garbage that is not valid XML.
	badContent := "this is not xml at all {{{}}}"
	timestamp := "1712900000"
	nonce := "nonce123"
	payload, err := c.crypto.BuildEncryptedReply([]byte(badContent), timestamp, nonce)
	if err != nil {
		t.Fatal(err)
	}

	u := fmt.Sprintf("/cb?msg_signature=%s&timestamp=%s&nonce=%s",
		extractSignature(t, c.crypto, payload, timestamp, nonce), timestamp, nonce)
	req := httptest.NewRequest("POST", u, bytes.NewReader(payload))

	_, err = c.ParseDataNotify(req)
	if err == nil {
		t.Fatal("expected error for invalid XML, got nil")
	}
}
