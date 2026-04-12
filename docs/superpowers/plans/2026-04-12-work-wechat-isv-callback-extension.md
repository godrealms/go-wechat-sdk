# Callback Extension Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Enrich existing InfoType callback events with per-ChangeType structs and add a new `ParseDataNotify` method for the data callback URL.

**Architecture:** Split flat `ChangeContactEvent`/`ChangeExternalContactEvent` into 12 typed structs with ChangeType-based secondary routing. Extract shared decryption logic into `decryptNotify`. Add `ParseDataNotify` with `DataEvent` interface for MsgType/Event routing (15 concrete types).

**Tech Stack:** Go 1.23+, `encoding/xml`, `github.com/godrealms/go-wechat-sdk/utils/wxcrypto`

---

### Task 1: Extend componentInner and Replace Event Structs

**Files:**
- Modify: `work-wechat/isv/struct.notify.go`
- Modify: `work-wechat/isv/notify.go` (componentInner struct only)

- [ ] **Step 1: Add new fields to componentInner in notify.go**

In `work-wechat/isv/notify.go`, add these fields to the `componentInner` struct after the existing `IsAdmin` field (line 35):

```go
	// Contact change – user fields
	Mobile         string `xml:"Mobile,omitempty"`
	Email          string `xml:"Email,omitempty"`
	Position       string `xml:"Position,omitempty"`
	Avatar         string `xml:"Avatar,omitempty"`
	Gender         int    `xml:"Gender,omitempty"`
	Status         int    `xml:"Status,omitempty"`
	IsLeaderInDept string `xml:"IsLeaderInDept,omitempty"`
	ExtAttr        string `xml:"ExtAttr,omitempty"`
	// Contact change – department fields
	ID       int `xml:"Id,omitempty"`
	ParentID int `xml:"ParentId,omitempty"`
	Order    int `xml:"Order,omitempty"`
	// Contact change – tag fields
	TagID         int    `xml:"TagId,omitempty"`
	AddUserItems  string `xml:"AddUserItems,omitempty"`
	DelUserItems  string `xml:"DelUserItems,omitempty"`
	AddPartyItems string `xml:"AddPartyItems,omitempty"`
	DelPartyItems string `xml:"DelPartyItems,omitempty"`
	// External contact fields
	State       string `xml:"State,omitempty"`
	WelcomeCode string `xml:"WelcomeCode,omitempty"`
```

- [ ] **Step 2: Replace ChangeContactEvent and ChangeExternalContactEvent in struct.notify.go**

In `work-wechat/isv/struct.notify.go`, delete `ChangeContactEvent` (lines 49-57) and `ChangeExternalContactEvent` (lines 60-66). Replace with these 12 structs:

```go
// ── Contact change: user ──────────────────────────────────────────

// ContactCreateUserEvent is fired when a user is created in the authorized corp.
type ContactCreateUserEvent struct {
	baseEvent
	AuthCorpID     string
	ChangeType     string // "create_user"
	UserID         string
	Name           string
	Department     string
	Mobile         string
	Email          string
	Position       string
	Gender         int
	Avatar         string
	Status         int
	IsLeaderInDept string
	ExtAttr        string
}

// ContactUpdateUserEvent is fired when a user is updated.
type ContactUpdateUserEvent struct {
	baseEvent
	AuthCorpID     string
	ChangeType     string // "update_user"
	UserID         string
	NewUserID      string
	Name           string
	Department     string
	Mobile         string
	Email          string
	Position       string
	Gender         int
	Avatar         string
	Status         int
	IsLeaderInDept string
	ExtAttr        string
}

// ContactDeleteUserEvent is fired when a user is deleted.
type ContactDeleteUserEvent struct {
	baseEvent
	AuthCorpID string
	ChangeType string // "delete_user"
	UserID     string
}

// ── Contact change: department ────────────────────────────────────

// ContactCreatePartyEvent is fired when a department is created.
type ContactCreatePartyEvent struct {
	baseEvent
	AuthCorpID string
	ChangeType string // "create_party"
	ID         int
	Name       string
	ParentID   int
	Order      int
}

// ContactUpdatePartyEvent is fired when a department is updated.
type ContactUpdatePartyEvent struct {
	baseEvent
	AuthCorpID string
	ChangeType string // "update_party"
	ID         int
	Name       string
	ParentID   int
}

// ContactDeletePartyEvent is fired when a department is deleted.
type ContactDeletePartyEvent struct {
	baseEvent
	AuthCorpID string
	ChangeType string // "delete_party"
	ID         int
}

// ── Contact change: tag ───────────────────────────────────────────

// ContactUpdateTagEvent is fired when a tag's membership changes.
type ContactUpdateTagEvent struct {
	baseEvent
	AuthCorpID    string
	ChangeType    string // "update_tag"
	TagID         int
	AddUserItems  string
	DelUserItems  string
	AddPartyItems string
	DelPartyItems string
}

// ── External contact changes ──────────────────────────────────────

// ExtContactAddEvent is fired when an external contact is added.
type ExtContactAddEvent struct {
	baseEvent
	AuthCorpID     string
	ChangeType     string // "add_external_contact"
	UserID         string
	ExternalUserID string
	State          string
	WelcomeCode    string
}

// ExtContactEditEvent is fired when an external contact is edited.
type ExtContactEditEvent struct {
	baseEvent
	AuthCorpID     string
	ChangeType     string // "edit_external_contact"
	UserID         string
	ExternalUserID string
}

// ExtContactDelEvent is fired when an external contact is deleted.
type ExtContactDelEvent struct {
	baseEvent
	AuthCorpID     string
	ChangeType     string // "del_external_contact"
	UserID         string
	ExternalUserID string
}

// ExtContactDelFollowEvent is fired when a follow user is removed.
type ExtContactDelFollowEvent struct {
	baseEvent
	AuthCorpID     string
	ChangeType     string // "del_follow_user"
	UserID         string
	ExternalUserID string
}

// ExtContactAddHalfEvent is fired for a half-added external contact.
type ExtContactAddHalfEvent struct {
	baseEvent
	AuthCorpID     string
	ChangeType     string // "add_half_external_contact"
	UserID         string
	ExternalUserID string
	State          string
	WelcomeCode    string
}
```

- [ ] **Step 3: Verify the code compiles**

Run: `cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk && go vet ./work-wechat/isv/`

Expected: Compilation errors because `notify.go` still references `ChangeContactEvent` and `ChangeExternalContactEvent`. This is expected — we will fix the routing in Task 2.

- [ ] **Step 4: Commit struct changes**

```bash
git add work-wechat/isv/struct.notify.go work-wechat/isv/notify.go
git commit -m "feat(work-wechat/isv): add per-ChangeType event structs and extend componentInner

Replace flat ChangeContactEvent/ChangeExternalContactEvent with 12
typed structs. Extend componentInner with user/department/tag/external
contact fields for secondary routing.

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

Note: This commit may not compile yet — the routing update in Task 2 will fix that.

---

### Task 2: Extract decryptNotify and Add ChangeType Secondary Routing

**Files:**
- Modify: `work-wechat/isv/notify.go`

- [ ] **Step 1: Extract decryptNotify helper**

In `work-wechat/isv/notify.go`, add this private method before `ParseNotify`:

```go
// decryptNotify reads, verifies, and decrypts a callback request body.
func (c *Client) decryptNotify(r *http.Request) ([]byte, error) {
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("isv: read body: %w", err)
	}
	var env componentEnvelope
	if err := xml.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("isv: parse envelope: %w", err)
	}

	q := r.URL.Query()
	msgSig := q.Get("msg_signature")
	timestamp := q.Get("timestamp")
	nonce := q.Get("nonce")
	if !c.crypto.VerifySignature(msgSig, timestamp, nonce, env.Encrypt) {
		return nil, errors.New("isv: invalid msg_signature")
	}

	plain, _, err := c.crypto.Decrypt(env.Encrypt)
	if err != nil {
		return nil, fmt.Errorf("isv: decrypt: %w", err)
	}
	return plain, nil
}
```

- [ ] **Step 2: Refactor ParseNotify to use decryptNotify and add secondary routing**

Replace the entire `ParseNotify` method body with:

```go
func (c *Client) ParseNotify(r *http.Request) (Event, error) {
	ctx := r.Context()

	plain, err := c.decryptNotify(r)
	if err != nil {
		return nil, err
	}

	var inner componentInner
	if err := xml.Unmarshal(plain, &inner); err != nil {
		return nil, fmt.Errorf("isv: parse inner: %w", err)
	}

	now := time.Now()
	base := baseEvent{SuiteID: inner.SuiteID, ReceiveAt: now}

	switch inner.InfoType {
	case "suite_ticket":
		if err := c.store.PutSuiteTicket(ctx, inner.SuiteID, inner.SuiteTicket); err != nil {
			return nil, fmt.Errorf("isv: persist suite_ticket: %w", err)
		}
		return &SuiteTicketEvent{baseEvent: base, SuiteTicket: inner.SuiteTicket}, nil
	case "create_auth":
		return &CreateAuthEvent{baseEvent: base, AuthCode: inner.AuthCode}, nil
	case "change_auth":
		return &ChangeAuthEvent{baseEvent: base, AuthCorpID: inner.AuthCorpID}, nil
	case "cancel_auth":
		return &CancelAuthEvent{baseEvent: base, AuthCorpID: inner.AuthCorpID}, nil
	case "reset_permanent_code":
		return &ResetPermanentCodeEvent{baseEvent: base, AuthCorpID: inner.AuthCorpID}, nil

	case "change_contact":
		switch inner.ChangeType {
		case "create_user":
			return &ContactCreateUserEvent{
				baseEvent: base, AuthCorpID: inner.AuthCorpID,
				ChangeType: inner.ChangeType, UserID: inner.UserID,
				Name: inner.Name, Department: inner.Department,
				Mobile: inner.Mobile, Email: inner.Email,
				Position: inner.Position, Gender: inner.Gender,
				Avatar: inner.Avatar, Status: inner.Status,
				IsLeaderInDept: inner.IsLeaderInDept, ExtAttr: inner.ExtAttr,
			}, nil
		case "update_user":
			return &ContactUpdateUserEvent{
				baseEvent: base, AuthCorpID: inner.AuthCorpID,
				ChangeType: inner.ChangeType, UserID: inner.UserID,
				NewUserID: inner.NewUserID, Name: inner.Name,
				Department: inner.Department, Mobile: inner.Mobile,
				Email: inner.Email, Position: inner.Position,
				Gender: inner.Gender, Avatar: inner.Avatar,
				Status: inner.Status, IsLeaderInDept: inner.IsLeaderInDept,
				ExtAttr: inner.ExtAttr,
			}, nil
		case "delete_user":
			return &ContactDeleteUserEvent{
				baseEvent: base, AuthCorpID: inner.AuthCorpID,
				ChangeType: inner.ChangeType, UserID: inner.UserID,
			}, nil
		case "create_party":
			return &ContactCreatePartyEvent{
				baseEvent: base, AuthCorpID: inner.AuthCorpID,
				ChangeType: inner.ChangeType, ID: inner.ID,
				Name: inner.Name, ParentID: inner.ParentID, Order: inner.Order,
			}, nil
		case "update_party":
			return &ContactUpdatePartyEvent{
				baseEvent: base, AuthCorpID: inner.AuthCorpID,
				ChangeType: inner.ChangeType, ID: inner.ID,
				Name: inner.Name, ParentID: inner.ParentID,
			}, nil
		case "delete_party":
			return &ContactDeletePartyEvent{
				baseEvent: base, AuthCorpID: inner.AuthCorpID,
				ChangeType: inner.ChangeType, ID: inner.ID,
			}, nil
		case "update_tag":
			return &ContactUpdateTagEvent{
				baseEvent: base, AuthCorpID: inner.AuthCorpID,
				ChangeType: inner.ChangeType, TagID: inner.TagID,
				AddUserItems: inner.AddUserItems, DelUserItems: inner.DelUserItems,
				AddPartyItems: inner.AddPartyItems, DelPartyItems: inner.DelPartyItems,
			}, nil
		default:
			return &RawEvent{baseEvent: base, InfoType: inner.InfoType, RawXML: string(plain)}, nil
		}

	case "change_external_contact":
		switch inner.ChangeType {
		case "add_external_contact":
			return &ExtContactAddEvent{
				baseEvent: base, AuthCorpID: inner.AuthCorpID,
				ChangeType: inner.ChangeType, UserID: inner.UserID,
				ExternalUserID: inner.ExternalUserID,
				State: inner.State, WelcomeCode: inner.WelcomeCode,
			}, nil
		case "edit_external_contact":
			return &ExtContactEditEvent{
				baseEvent: base, AuthCorpID: inner.AuthCorpID,
				ChangeType: inner.ChangeType, UserID: inner.UserID,
				ExternalUserID: inner.ExternalUserID,
			}, nil
		case "del_external_contact":
			return &ExtContactDelEvent{
				baseEvent: base, AuthCorpID: inner.AuthCorpID,
				ChangeType: inner.ChangeType, UserID: inner.UserID,
				ExternalUserID: inner.ExternalUserID,
			}, nil
		case "del_follow_user":
			return &ExtContactDelFollowEvent{
				baseEvent: base, AuthCorpID: inner.AuthCorpID,
				ChangeType: inner.ChangeType, UserID: inner.UserID,
				ExternalUserID: inner.ExternalUserID,
			}, nil
		case "add_half_external_contact":
			return &ExtContactAddHalfEvent{
				baseEvent: base, AuthCorpID: inner.AuthCorpID,
				ChangeType: inner.ChangeType, UserID: inner.UserID,
				ExternalUserID: inner.ExternalUserID,
				State: inner.State, WelcomeCode: inner.WelcomeCode,
			}, nil
		default:
			return &RawEvent{baseEvent: base, InfoType: inner.InfoType, RawXML: string(plain)}, nil
		}

	case "share_agent_change":
		return &ShareAgentChangeEvent{
			baseEvent: base, AuthCorpID: inner.AuthCorpID, AgentID: inner.AgentID,
		}, nil
	case "change_app_admin":
		return &ChangeAppAdminEvent{
			baseEvent: base, AuthCorpID: inner.AuthCorpID,
			UserID: inner.UserID, IsAdmin: inner.IsAdmin == 1,
		}, nil
	default:
		return &RawEvent{baseEvent: base, InfoType: inner.InfoType, RawXML: string(plain)}, nil
	}
}
```

- [ ] **Step 3: Verify the code compiles and existing tests pass**

Run: `cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk && go test ./work-wechat/isv/ -run "TestParseNotify_SuiteTicket|TestParseNotify_CreateAuth|TestParseNotify_ChangeAuth|TestParseNotify_CancelAuth|TestParseNotify_ResetPermanentCode|TestParseNotify_ShareAgentChange|TestParseNotify_ChangeAppAdmin|TestParseNotify_UnknownInfoType|TestParseNotify_BadSignature" -v`

Expected: All 9 existing tests PASS. The two old ChangeContact/ChangeExternalContact tests will fail because the return type changed — that is expected and will be fixed in Task 3.

- [ ] **Step 4: Commit**

```bash
git add work-wechat/isv/notify.go
git commit -m "feat(work-wechat/isv): extract decryptNotify and add ChangeType secondary routing

Refactor ParseNotify to use shared decryptNotify helper. Replace flat
change_contact/change_external_contact cases with per-ChangeType
secondary switch returning 12 typed event structs.

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

### Task 3: Update Existing Tests and Add ChangeType Tests

**Files:**
- Modify: `work-wechat/isv/notify_test.go`

- [ ] **Step 1: Replace existing ChangeContact and ChangeExternalContact tests**

In `work-wechat/isv/notify_test.go`, replace `TestParseNotify_ChangeContact` (lines 134-158) with these 7 tests:

```go
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
	if cev.ChangeType != "create_user" || cev.UserID != "u1" || cev.Name != "Alice" {
		t.Errorf("basic: %+v", cev)
	}
	if cev.Mobile != "13800138000" || cev.Email != "alice@example.com" || cev.Position != "Engineer" {
		t.Errorf("contact: %+v", cev)
	}
	if cev.Gender != 1 || cev.Status != 1 || cev.IsLeaderInDept != "0,1" {
		t.Errorf("flags: %+v", cev)
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
	if cev.ChangeType != "update_user" || cev.NewUserID != "u1new" || cev.Name != "Alice Updated" {
		t.Errorf("event: %+v", cev)
	}
}

func TestParseNotify_ContactDeleteUser(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml>
<SuiteId><![CDATA[suite1]]></SuiteId>
<InfoType><![CDATA[change_contact]]></InfoType>
<AuthCorpId><![CDATA[wxcorp1]]></AuthCorpId>
<ChangeType><![CDATA[delete_user]]></ChangeType>
<UserID><![CDATA[u1]]></UserID>
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
	if cev.UserID != "u1" {
		t.Errorf("event: %+v", cev)
	}
}

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
<Order>100</Order>
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
	if cev.ID != 10 || cev.Name != "Engineering" || cev.ParentID != 1 || cev.Order != 100 {
		t.Errorf("event: %+v", cev)
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
<Name><![CDATA[Eng Updated]]></Name>
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
	if cev.ID != 10 || cev.Name != "Eng Updated" || cev.ParentID != 2 {
		t.Errorf("event: %+v", cev)
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
	if cev.ID != 10 {
		t.Errorf("event: %+v", cev)
	}
}

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
<AddPartyItems><![CDATA[1,2]]></AddPartyItems>
<DelPartyItems><![CDATA[3]]></DelPartyItems>
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
	if cev.TagID != 100 || cev.AddUserItems != "u1,u2" || cev.DelUserItems != "u3" {
		t.Errorf("event: %+v", cev)
	}
	if cev.AddPartyItems != "1,2" || cev.DelPartyItems != "3" {
		t.Errorf("party items: %+v", cev)
	}
}
```

- [ ] **Step 2: Replace TestParseNotify_ChangeExternalContact with 5 tests**

Replace `TestParseNotify_ChangeExternalContact` (lines 160-173) with:

```go
func TestParseNotify_ExtContactAdd(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml>
<SuiteId><![CDATA[suite1]]></SuiteId>
<InfoType><![CDATA[change_external_contact]]></InfoType>
<AuthCorpId><![CDATA[wxcorp1]]></AuthCorpId>
<ChangeType><![CDATA[add_external_contact]]></ChangeType>
<UserID><![CDATA[u1]]></UserID>
<ExternalUserID><![CDATA[ex1]]></ExternalUserID>
<State><![CDATA[state123]]></State>
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
	if cev.ExternalUserID != "ex1" || cev.State != "state123" || cev.WelcomeCode != "wc123" {
		t.Errorf("event: %+v", cev)
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
	if cev.UserID != "u1" || cev.ExternalUserID != "ex1" {
		t.Errorf("event: %+v", cev)
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
<ExternalUserID><![CDATA[ex1]]></ExternalUserID>
</xml>`
	req := buildNotifyRequest(t, c, inner)
	ev, err := c.ParseNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := ev.(*ExtContactDelEvent); !ok {
		t.Fatalf("type: %T", ev)
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
	if _, ok := ev.(*ExtContactDelFollowEvent); !ok {
		t.Fatalf("type: %T", ev)
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
<ExternalUserID><![CDATA[ex1]]></ExternalUserID>
<State><![CDATA[state456]]></State>
<WelcomeCode><![CDATA[wc456]]></WelcomeCode>
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
	if cev.State != "state456" || cev.WelcomeCode != "wc456" {
		t.Errorf("event: %+v", cev)
	}
}
```

- [ ] **Step 3: Run all tests**

Run: `cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk && go test ./work-wechat/isv/ -race -count=1 -v 2>&1 | grep -E "^(ok|FAIL|---)"` 

Expected: All tests PASS including the 12 new ChangeType tests.

- [ ] **Step 4: Commit**

```bash
git add work-wechat/isv/notify_test.go
git commit -m "test(work-wechat/isv): add per-ChangeType tests for ParseNotify

Replace flat ChangeContact/ChangeExternalContact tests with 12
individual tests covering all ChangeType secondary routing paths.

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

### Task 4: Create DataEvent DTOs

**Files:**
- Create: `work-wechat/isv/struct.data_notify.go`

- [ ] **Step 1: Create struct.data_notify.go with all DataEvent types**

Create `work-wechat/isv/struct.data_notify.go`:

```go
package isv

import "time"

// DataEvent represents an event from the data callback URL (数据回调).
type DataEvent interface {
	isDataEvent()
}

type dataEventBase struct {
	ToUserName string    // corp ID receiving the event
	AgentID    int       // application agent ID
	ReceiveAt  time.Time // time.Now() at parse time
}

func (dataEventBase) isDataEvent() {}

// ── Message types ─────────────────────────────────────────────────

// DataTextMsg represents a text message from a user.
type DataTextMsg struct {
	dataEventBase
	MsgID   int64
	Content string
}

// DataImageMsg represents an image message.
type DataImageMsg struct {
	dataEventBase
	MsgID   int64
	PicURL  string
	MediaID string
}

// DataVoiceMsg represents a voice message.
type DataVoiceMsg struct {
	dataEventBase
	MsgID   int64
	MediaID string
	Format  string
}

// DataVideoMsg represents a video message.
type DataVideoMsg struct {
	dataEventBase
	MsgID        int64
	MediaID      string
	ThumbMediaID string
}

// DataLocationMsg represents a location message.
type DataLocationMsg struct {
	dataEventBase
	MsgID int64
	Lat   float64
	Lng   float64
	Scale int
	Label string
}

// DataLinkMsg represents a link message.
type DataLinkMsg struct {
	dataEventBase
	MsgID       int64
	Title       string
	Description string
	URL         string
}

// DataRawMsg is the fallback for unknown MsgType values.
type DataRawMsg struct {
	dataEventBase
	MsgType string
	RawXML  string
}

// ── Event types (MsgType=event) ───────────────────────────────────

// DataEnterAgentEvent is fired when a user enters the application.
type DataEnterAgentEvent struct {
	dataEventBase
	EventKey string
}

// DataMenuClickEvent is fired when a user clicks a menu button.
type DataMenuClickEvent struct {
	dataEventBase
	EventKey string
}

// DataMenuViewEvent is fired when a user clicks a menu link.
type DataMenuViewEvent struct {
	dataEventBase
	EventKey string
	URL      string
}

// DataApprovalChangeEvent is fired when an approval status changes.
type DataApprovalChangeEvent struct {
	dataEventBase
	ApprovalInfo ApprovalInfo
}

// ApprovalInfo contains the nested approval details.
type ApprovalInfo struct {
	ThirdNo       string         `xml:"ThirdNo"`
	OpenSpName    string         `xml:"OpenSpName"`
	OpenSpStatus  int            `xml:"OpenSpStatus"`
	ApplyTime     int64          `xml:"ApplyTime"`
	ApplyUserID   string         `xml:"ApplyUserid"`
	ApplyUserName string         `xml:"ApplyUserName"`
	ApprovalNodes []ApprovalNode `xml:"ApprovalNodes>ApprovalNode"`
	NotifyNodes   []NotifyNode   `xml:"NotifyNodes>NotifyNode"`
}

// ApprovalNode represents one node in the approval chain.
type ApprovalNode struct {
	NodeStatus int            `xml:"NodeStatus"`
	NodeAttr   int            `xml:"NodeAttr"`
	Items      []ApprovalItem `xml:"Items>Item"`
}

// ApprovalItem represents one approver within a node.
type ApprovalItem struct {
	ItemName   string `xml:"ItemName"`
	ItemUserID string `xml:"ItemUserid"`
	ItemStatus int    `xml:"ItemStatus"`
	ItemSpeech string `xml:"ItemSpeech"`
	ItemOpTime int64  `xml:"ItemOpTime"`
}

// NotifyNode represents a notification recipient.
type NotifyNode struct {
	ItemName   string `xml:"ItemName"`
	ItemUserID string `xml:"ItemUserid"`
}

// DataBatchJobResultEvent is fired when an async batch job completes.
type DataBatchJobResultEvent struct {
	dataEventBase
	BatchJob BatchJobResult
}

// BatchJobResult contains the async job completion details.
type BatchJobResult struct {
	JobID   string `xml:"JobId"`
	JobType string `xml:"JobType"`
	ErrCode int    `xml:"ErrCode"`
	ErrMsg  string `xml:"ErrMsg"`
}

// DataExtContactChangeEvent is fired for external contact changes via data callback.
type DataExtContactChangeEvent struct {
	dataEventBase
	ChangeType     string
	UserID         string
	ExternalUserID string
	State          string
	WelcomeCode    string
}

// DataRawEvent is the fallback for unknown Event values under MsgType=event.
type DataRawEvent struct {
	dataEventBase
	Event  string
	RawXML string
}
```

- [ ] **Step 2: Verify compilation**

Run: `cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk && go vet ./work-wechat/isv/`

Expected: PASS (no references to these types yet beyond compilation)

- [ ] **Step 3: Commit**

```bash
git add work-wechat/isv/struct.data_notify.go
git commit -m "feat(work-wechat/isv): add DataEvent interface and DTOs for data callback

Add 15 DataEvent types (7 message + 6 event + 2 fallback) plus
ApprovalInfo/BatchJobResult nested structs for ParseDataNotify.

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

### Task 5: Implement ParseDataNotify

**Files:**
- Create: `work-wechat/isv/data_notify.go`

- [ ] **Step 1: Create data_notify.go with dataInner and ParseDataNotify**

Create `work-wechat/isv/data_notify.go`:

```go
package isv

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"time"
)

// dataInner is the decrypted inner XML for data callback events.
type dataInner struct {
	XMLName    xml.Name `xml:"xml"`
	ToUserName string   `xml:"ToUserName"`
	AgentID    int      `xml:"AgentID"`
	MsgType    string   `xml:"MsgType"`
	Event      string   `xml:"Event,omitempty"`
	EventKey   string   `xml:"EventKey,omitempty"`
	// Message fields
	MsgID        int64   `xml:"MsgId,omitempty"`
	Content      string  `xml:"Content,omitempty"`
	PicURL       string  `xml:"PicUrl,omitempty"`
	MediaID      string  `xml:"MediaId,omitempty"`
	Format       string  `xml:"Format,omitempty"`
	ThumbMediaID string  `xml:"ThumbMediaId,omitempty"`
	LocationX    float64 `xml:"Location_X,omitempty"`
	LocationY    float64 `xml:"Location_Y,omitempty"`
	Scale        int     `xml:"Scale,omitempty"`
	Label        string  `xml:"Label,omitempty"`
	Title        string  `xml:"Title,omitempty"`
	Description  string  `xml:"Description,omitempty"`
	URL          string  `xml:"Url,omitempty"`
	// Event-specific
	ChangeType     string       `xml:"ChangeType,omitempty"`
	UserID         string       `xml:"UserID,omitempty"`
	ExternalUserID string       `xml:"ExternalUserID,omitempty"`
	State          string       `xml:"State,omitempty"`
	WelcomeCode    string       `xml:"WelcomeCode,omitempty"`
	ApprovalInfo   ApprovalInfo `xml:"ApprovalInfo"`
	BatchJob       BatchJobResult `xml:"BatchJob"`
}

// ParseDataNotify verifies, decrypts, and parses a data callback request,
// returning a strongly typed DataEvent.
func (c *Client) ParseDataNotify(r *http.Request) (DataEvent, error) {
	plain, err := c.decryptNotify(r)
	if err != nil {
		return nil, err
	}

	var inner dataInner
	if err := xml.Unmarshal(plain, &inner); err != nil {
		return nil, fmt.Errorf("isv: unmarshal data notify: %w", err)
	}

	now := time.Now()
	base := dataEventBase{
		ToUserName: inner.ToUserName,
		AgentID:    inner.AgentID,
		ReceiveAt:  now,
	}

	switch inner.MsgType {
	case "text":
		return &DataTextMsg{dataEventBase: base, MsgID: inner.MsgID, Content: inner.Content}, nil
	case "image":
		return &DataImageMsg{dataEventBase: base, MsgID: inner.MsgID, PicURL: inner.PicURL, MediaID: inner.MediaID}, nil
	case "voice":
		return &DataVoiceMsg{dataEventBase: base, MsgID: inner.MsgID, MediaID: inner.MediaID, Format: inner.Format}, nil
	case "video":
		return &DataVideoMsg{dataEventBase: base, MsgID: inner.MsgID, MediaID: inner.MediaID, ThumbMediaID: inner.ThumbMediaID}, nil
	case "location":
		return &DataLocationMsg{dataEventBase: base, MsgID: inner.MsgID, Lat: inner.LocationX, Lng: inner.LocationY, Scale: inner.Scale, Label: inner.Label}, nil
	case "link":
		return &DataLinkMsg{dataEventBase: base, MsgID: inner.MsgID, Title: inner.Title, Description: inner.Description, URL: inner.URL}, nil
	case "event":
		switch inner.Event {
		case "enter_agent":
			return &DataEnterAgentEvent{dataEventBase: base, EventKey: inner.EventKey}, nil
		case "click":
			return &DataMenuClickEvent{dataEventBase: base, EventKey: inner.EventKey}, nil
		case "view":
			return &DataMenuViewEvent{dataEventBase: base, EventKey: inner.EventKey, URL: inner.URL}, nil
		case "open_approval_change":
			return &DataApprovalChangeEvent{dataEventBase: base, ApprovalInfo: inner.ApprovalInfo}, nil
		case "batch_job_result":
			return &DataBatchJobResultEvent{dataEventBase: base, BatchJob: inner.BatchJob}, nil
		case "change_external_contact":
			return &DataExtContactChangeEvent{
				dataEventBase: base, ChangeType: inner.ChangeType,
				UserID: inner.UserID, ExternalUserID: inner.ExternalUserID,
				State: inner.State, WelcomeCode: inner.WelcomeCode,
			}, nil
		default:
			return &DataRawEvent{dataEventBase: base, Event: inner.Event, RawXML: string(plain)}, nil
		}
	default:
		return &DataRawMsg{dataEventBase: base, MsgType: inner.MsgType, RawXML: string(plain)}, nil
	}
}
```

Note: Add `"net/http"` to the import list — the file needs it for `*http.Request`.

- [ ] **Step 2: Verify compilation**

Run: `cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk && go vet ./work-wechat/isv/`

Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add work-wechat/isv/data_notify.go
git commit -m "feat(work-wechat/isv): implement ParseDataNotify for data callback URL

Add MsgType/Event routing supporting 7 message types, 6 event types,
and 2 fallbacks (DataRawMsg, DataRawEvent). Reuses decryptNotify
for shared signature verification and decryption.

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

### Task 6: Add ParseDataNotify Tests

**Files:**
- Create: `work-wechat/isv/data_notify_test.go`

- [ ] **Step 1: Create data_notify_test.go with helper and all tests**

Create `work-wechat/isv/data_notify_test.go`:

```go
package isv

import (
	"strings"
	"testing"
)

// buildDataNotifyRequest constructs a signed/encrypted HTTP POST for data callback.
// innerXML is the plaintext inner body (no Encrypt envelope).
// Reuses the same Client crypto as buildNotifyRequest.
func buildDataNotifyRequest(t *testing.T, c *Client, innerXML string) *http.Request {
	t.Helper()
	return buildNotifyRequest(t, c, innerXML)
}

func TestParseDataNotify_TextMsg(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml>
<ToUserName><![CDATA[wxcorp1]]></ToUserName>
<AgentID>1000001</AgentID>
<MsgType><![CDATA[text]]></MsgType>
<MsgId>12345</MsgId>
<Content><![CDATA[Hello World]]></Content>
</xml>`
	req := buildDataNotifyRequest(t, c, inner)
	ev, err := c.ParseDataNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	msg, ok := ev.(*DataTextMsg)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if msg.ToUserName != "wxcorp1" || msg.AgentID != 1000001 {
		t.Errorf("base: %+v", msg)
	}
	if msg.MsgID != 12345 || msg.Content != "Hello World" {
		t.Errorf("text: %+v", msg)
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
<MediaId><![CDATA[M001]]></MediaId>
</xml>`
	req := buildDataNotifyRequest(t, c, inner)
	ev, err := c.ParseDataNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	msg, ok := ev.(*DataImageMsg)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if msg.MsgID != 12346 || msg.PicURL != "https://img/pic.jpg" || msg.MediaID != "M001" {
		t.Errorf("image: %+v", msg)
	}
}

func TestParseDataNotify_VoiceMsg(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml>
<ToUserName><![CDATA[wxcorp1]]></ToUserName>
<AgentID>1000001</AgentID>
<MsgType><![CDATA[voice]]></MsgType>
<MsgId>12347</MsgId>
<MediaId><![CDATA[M002]]></MediaId>
<Format><![CDATA[amr]]></Format>
</xml>`
	req := buildDataNotifyRequest(t, c, inner)
	ev, err := c.ParseDataNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	msg, ok := ev.(*DataVoiceMsg)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if msg.MediaID != "M002" || msg.Format != "amr" {
		t.Errorf("voice: %+v", msg)
	}
}

func TestParseDataNotify_VideoMsg(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml>
<ToUserName><![CDATA[wxcorp1]]></ToUserName>
<AgentID>1000001</AgentID>
<MsgType><![CDATA[video]]></MsgType>
<MsgId>12348</MsgId>
<MediaId><![CDATA[M003]]></MediaId>
<ThumbMediaId><![CDATA[TM003]]></ThumbMediaId>
</xml>`
	req := buildDataNotifyRequest(t, c, inner)
	ev, err := c.ParseDataNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	msg, ok := ev.(*DataVideoMsg)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if msg.MediaID != "M003" || msg.ThumbMediaID != "TM003" {
		t.Errorf("video: %+v", msg)
	}
}

func TestParseDataNotify_LocationMsg(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml>
<ToUserName><![CDATA[wxcorp1]]></ToUserName>
<AgentID>1000001</AgentID>
<MsgType><![CDATA[location]]></MsgType>
<MsgId>12349</MsgId>
<Location_X>39.9042</Location_X>
<Location_Y>116.4074</Location_Y>
<Scale>15</Scale>
<Label><![CDATA[Beijing]]></Label>
</xml>`
	req := buildDataNotifyRequest(t, c, inner)
	ev, err := c.ParseDataNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	msg, ok := ev.(*DataLocationMsg)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if msg.Lat != 39.9042 || msg.Lng != 116.4074 || msg.Scale != 15 || msg.Label != "Beijing" {
		t.Errorf("location: %+v", msg)
	}
}

func TestParseDataNotify_LinkMsg(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml>
<ToUserName><![CDATA[wxcorp1]]></ToUserName>
<AgentID>1000001</AgentID>
<MsgType><![CDATA[link]]></MsgType>
<MsgId>12350</MsgId>
<Title><![CDATA[My Link]]></Title>
<Description><![CDATA[A description]]></Description>
<Url><![CDATA[https://example.com]]></Url>
</xml>`
	req := buildDataNotifyRequest(t, c, inner)
	ev, err := c.ParseDataNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	msg, ok := ev.(*DataLinkMsg)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if msg.Title != "My Link" || msg.Description != "A description" || msg.URL != "https://example.com" {
		t.Errorf("link: %+v", msg)
	}
}

func TestParseDataNotify_EnterAgent(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml>
<ToUserName><![CDATA[wxcorp1]]></ToUserName>
<AgentID>1000001</AgentID>
<MsgType><![CDATA[event]]></MsgType>
<Event><![CDATA[enter_agent]]></Event>
<EventKey><![CDATA[]]></EventKey>
</xml>`
	req := buildDataNotifyRequest(t, c, inner)
	ev, err := c.ParseDataNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := ev.(*DataEnterAgentEvent); !ok {
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
	req := buildDataNotifyRequest(t, c, inner)
	ev, err := c.ParseDataNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	mev, ok := ev.(*DataMenuClickEvent)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if mev.EventKey != "btn_report" {
		t.Errorf("event: %+v", mev)
	}
}

func TestParseDataNotify_MenuView(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml>
<ToUserName><![CDATA[wxcorp1]]></ToUserName>
<AgentID>1000001</AgentID>
<MsgType><![CDATA[event]]></MsgType>
<Event><![CDATA[view]]></Event>
<EventKey><![CDATA[https://example.com/page]]></EventKey>
<Url><![CDATA[https://example.com/page]]></Url>
</xml>`
	req := buildDataNotifyRequest(t, c, inner)
	ev, err := c.ParseDataNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	mev, ok := ev.(*DataMenuViewEvent)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if mev.EventKey != "https://example.com/page" || mev.URL != "https://example.com/page" {
		t.Errorf("event: %+v", mev)
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
	req := buildDataNotifyRequest(t, c, inner)
	ev, err := c.ParseDataNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	aev, ok := ev.(*DataApprovalChangeEvent)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	ai := aev.ApprovalInfo
	if ai.ThirdNo != "T001" || ai.OpenSpName != "Leave" || ai.OpenSpStatus != 1 {
		t.Errorf("approval base: %+v", ai)
	}
	if ai.ApplyUserID != "u_apply" || ai.ApplyUserName != "Bob" {
		t.Errorf("apply user: %+v", ai)
	}
	if len(ai.ApprovalNodes) != 1 {
		t.Fatalf("nodes: %d", len(ai.ApprovalNodes))
	}
	node := ai.ApprovalNodes[0]
	if node.NodeStatus != 1 || node.NodeAttr != 1 || len(node.Items) != 1 {
		t.Errorf("node: %+v", node)
	}
	if node.Items[0].ItemUserID != "u_mgr" {
		t.Errorf("item: %+v", node.Items[0])
	}
	if len(ai.NotifyNodes) != 1 || ai.NotifyNodes[0].ItemUserID != "u_hr" {
		t.Errorf("notify: %+v", ai.NotifyNodes)
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
<JobId><![CDATA[JOB001]]></JobId>
<JobType><![CDATA[sync_user]]></JobType>
<ErrCode>0</ErrCode>
<ErrMsg><![CDATA[ok]]></ErrMsg>
</BatchJob>
</xml>`
	req := buildDataNotifyRequest(t, c, inner)
	ev, err := c.ParseDataNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	bev, ok := ev.(*DataBatchJobResultEvent)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if bev.BatchJob.JobID != "JOB001" || bev.BatchJob.JobType != "sync_user" || bev.BatchJob.ErrCode != 0 {
		t.Errorf("batch: %+v", bev.BatchJob)
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
<UserID><![CDATA[u1]]></UserID>
<ExternalUserID><![CDATA[ex1]]></ExternalUserID>
<State><![CDATA[st1]]></State>
<WelcomeCode><![CDATA[wc1]]></WelcomeCode>
</xml>`
	req := buildDataNotifyRequest(t, c, inner)
	ev, err := c.ParseDataNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	eev, ok := ev.(*DataExtContactChangeEvent)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if eev.ChangeType != "add_external_contact" || eev.UserID != "u1" || eev.ExternalUserID != "ex1" {
		t.Errorf("ext: %+v", eev)
	}
	if eev.State != "st1" || eev.WelcomeCode != "wc1" {
		t.Errorf("state: %+v", eev)
	}
}

func TestParseDataNotify_RawMsg(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml>
<ToUserName><![CDATA[wxcorp1]]></ToUserName>
<AgentID>1000001</AgentID>
<MsgType><![CDATA[shortvideo]]></MsgType>
<MsgId>99999</MsgId>
</xml>`
	req := buildDataNotifyRequest(t, c, inner)
	ev, err := c.ParseDataNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	raw, ok := ev.(*DataRawMsg)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if raw.MsgType != "shortvideo" {
		t.Errorf("msgtype: %q", raw.MsgType)
	}
	if !strings.Contains(raw.RawXML, "shortvideo") {
		t.Errorf("rawxml missing msgtype: %q", raw.RawXML)
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
	req := buildDataNotifyRequest(t, c, inner)
	ev, err := c.ParseDataNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	raw, ok := ev.(*DataRawEvent)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if raw.Event != "future_event_type" {
		t.Errorf("event: %q", raw.Event)
	}
}

func TestParseDataNotify_InvalidSignature(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml><ToUserName><![CDATA[wxcorp1]]></ToUserName><MsgType><![CDATA[text]]></MsgType></xml>`
	req := buildDataNotifyRequest(t, c, inner)
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
	// The buildDataNotifyRequest encrypts valid XML. We need to test inner XML parse failure.
	// Use a valid envelope but corrupt the inner content by using a body that decrypts to non-XML.
	// Simplest approach: just test that a totally broken request body fails.
	inner := `not xml at all`
	// This won't encrypt properly because buildNotifyRequest does crypto on it,
	// and the decrypted result would be "not xml at all" which isn't valid XML.
	req := buildDataNotifyRequest(t, c, inner)
	_, err := c.ParseDataNotify(req)
	if err == nil {
		t.Fatal("want error, got nil")
	}
}
```

Note: Add `"net/http"` to imports if the `buildDataNotifyRequest` wrapper needs it (the return type references `*http.Request`). However, since `buildNotifyRequest` already returns `*http.Request` and the helper just delegates, you need the import. The full import block should be:

```go
import (
	"strings"
	"testing"
)
```

Since `buildDataNotifyRequest` calls `buildNotifyRequest` which returns `*http.Request` — Go resolves the type through the function call, so no explicit `net/http` import is needed in the test file as long as `buildNotifyRequest` is in the same package. However, `*http.Request` IS used in the function signature. So you DO need `"net/http"` in imports:

```go
import (
	"net/http"
	"strings"
	"testing"
)
```

- [ ] **Step 2: Run all tests**

Run: `cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk && go test ./work-wechat/isv/ -race -count=1 -v 2>&1 | grep -E "^(ok|FAIL|---)"` 

Expected: All tests PASS including the 16 new ParseDataNotify tests.

- [ ] **Step 3: Check coverage**

Run: `cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk && go test ./work-wechat/isv/ -coverprofile=cover.out -count=1 && go tool cover -func=cover.out | tail -1`

Expected: Coverage ≥ 85%. If below, add additional tests for uncovered branches.

- [ ] **Step 4: Commit**

```bash
git add work-wechat/isv/data_notify_test.go
git commit -m "test(work-wechat/isv): add ParseDataNotify tests

16 tests covering all MsgType/Event routing paths: 6 message types,
6 event types (including nested ApprovalInfo/BatchJobResult), 2
fallbacks, and 2 error paths (bad signature, invalid XML).

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

### Task 7: Final Verification and Coverage

**Files:** None (verification only)

- [ ] **Step 1: Run full test suite with race detector**

Run: `cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk && go test ./work-wechat/isv/ -race -count=1 -v`

Expected: All tests PASS, no race conditions.

- [ ] **Step 2: Run go vet**

Run: `cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk && go vet ./work-wechat/isv/`

Expected: No issues.

- [ ] **Step 3: Verify coverage ≥ 85%**

Run: `cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk && go test ./work-wechat/isv/ -coverprofile=cover.out -count=1 && go tool cover -func=cover.out | tail -1`

Expected: Coverage ≥ 85%. If below 85%, identify uncovered lines with `go tool cover -func=cover.out | grep -v "100.0%"` and add targeted tests.

- [ ] **Step 4: If coverage is below 85%, add additional tests and commit**

Only if needed. Common areas that may need extra coverage:
- Unknown ChangeType defaults in change_contact/change_external_contact (falls to RawEvent)
- Edge cases in XML parsing

```bash
git add work-wechat/isv/
git commit -m "test(work-wechat/isv): improve callback extension test coverage

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```
