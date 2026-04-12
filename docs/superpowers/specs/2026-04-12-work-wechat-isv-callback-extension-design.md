# Sub-project 6: 事件回调扩展 — Design Spec

**Package:** `github.com/godrealms/go-wechat-sdk/work-wechat/isv`
**Date:** 2026-04-12
**Depends on:** Sub-project 1 (认证底座 — ParseNotify, Event interface, crypto)

---

## 1. Scope

Two workstreams:

1. **Enrich existing InfoType events** — split flat `ChangeContactEvent` and `ChangeExternalContactEvent` into per-ChangeType structs with full field coverage; add ChangeType-based secondary routing in `ParseNotify`.
2. **New data callback parser** — add `ParseDataNotify(r *http.Request) (DataEvent, error)` for the 数据回调 URL, supporting message types and business events routed by `MsgType` + `Event`.

### 1.1 Method Summary

| # | Area | Change |
|---|---|---|
| 1 | ParseNotify | Add ChangeType secondary switch for change_contact (7 types) and change_external_contact (5 types) |
| 2 | ParseDataNotify | New method: MsgType routing (7 message types) + Event routing (6 event types) + 2 fallbacks |

### 1.2 Struct Summary

- 12 new Event structs (replace 2 existing flat structs)
- 1 new DataEvent interface
- 15 new DataEvent structs (7 message + 6 event + 2 fallback)
- 2 nested structs (ApprovalInfo hierarchy, BatchJobResult)

---

## 2. Architecture

### 2.1 Two Callback Entry Points

```
┌─────────────────────┐    ┌──────────────────────┐
│  指令回调 URL        │    │  数据回调 URL          │
│  (suite callback)   │    │  (data callback)      │
│                     │    │                       │
│  ParseNotify()      │    │  ParseDataNotify()    │
│  → Event interface  │    │  → DataEvent interface│
└─────────────────────┘    └──────────────────────┘
        │                           │
   InfoType routing           MsgType routing
   ├─ suite_ticket            ├─ text → DataTextMsg
   ├─ create_auth             ├─ image → DataImageMsg
   ├─ change_auth             ├─ voice → DataVoiceMsg
   ├─ cancel_auth             ├─ video → DataVideoMsg
   ├─ reset_permanent_code    ├─ location → DataLocationMsg
   ├─ change_contact          ├─ link → DataLinkMsg
   │  └─ ChangeType routing   ├─ event → Event routing
   │     ├─ create_user       │  ├─ enter_agent
   │     ├─ update_user       │  ├─ click
   │     ├─ delete_user       │  ├─ view
   │     ├─ create_party      │  ├─ open_approval_change
   │     ├─ update_party      │  ├─ batch_job_result
   │     ├─ delete_party      │  ├─ change_external_contact
   │     └─ update_tag        │  └─ default → DataRawEvent
   ├─ change_external_contact └─ default → DataRawMsg
   │  └─ ChangeType routing
   │     ├─ add_external_contact
   │     ├─ edit_external_contact
   │     ├─ del_external_contact
   │     ├─ del_follow_user
   │     └─ add_half_external_contact
   ├─ share_agent_change
   ├─ change_app_admin
   └─ default → RawEvent
```

### 2.2 Shared Decryption

Both methods share the same flow:
1. Read `r.Body`
2. Unmarshal outer XML envelope (`<xml><Encrypt>...</Encrypt></xml>`)
3. Verify SHA1 signature via `c.crypto.VerifySignature`
4. Decrypt via `c.crypto.Decrypt`
5. Unmarshal inner XML to method-specific struct
6. Route and return typed event

Extract shared steps 1-4 into a private helper:

```go
func (c *Client) decryptNotify(r *http.Request) ([]byte, error)
```

Returns decrypted plaintext XML bytes. Both `ParseNotify` and `ParseDataNotify` call this, then unmarshal and route independently.

### 2.3 File Layout

| File | Content |
|---|---|
| `struct.notify.go` (modify) | Remove ChangeContactEvent, ChangeExternalContactEvent; add 12 per-ChangeType Event structs |
| `notify.go` (modify) | Extract `decryptNotify` helper; add ChangeType secondary switch in change_contact and change_external_contact cases |
| `struct.data_notify.go` (new) | DataEvent interface, dataEventBase, 7 message structs, 6 event structs, 2 fallback structs, ApprovalInfo/BatchJobResult nested types |
| `data_notify.go` (new) | ParseDataNotify method, dataInner struct, MsgType/Event routing |
| `notify_test.go` (modify) | Add 12 tests for ChangeType secondary routing |
| `data_notify_test.go` (new) | ~16 tests for ParseDataNotify |

---

## 3. DTOs

### 3.1 Enriched InfoType Events — Contact Changes

All embed `baseEvent` (SuiteID, ReceiveAt) and add AuthCorpID.

**User changes:**

```go
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
    ExtAttr        string // raw XML of extended attributes
}

// ContactUpdateUserEvent is fired when a user is updated.
type ContactUpdateUserEvent struct {
    baseEvent
    AuthCorpID     string
    ChangeType     string // "update_user"
    UserID         string
    NewUserID      string // only when userid changes
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
```

**Department changes:**

```go
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
```

**Tag changes:**

```go
// ContactUpdateTagEvent is fired when a tag's membership changes.
type ContactUpdateTagEvent struct {
    baseEvent
    AuthCorpID    string
    ChangeType    string // "update_tag"
    TagID         int
    AddUserItems  string // comma-separated userids
    DelUserItems  string
    AddPartyItems string // comma-separated department ids
    DelPartyItems string
}
```

### 3.2 Enriched InfoType Events — External Contact Changes

```go
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

### 3.3 componentInner Extension

Add fields to the existing `componentInner` struct to carry ChangeType-specific data:

```go
// User-related
Mobile         string `xml:"Mobile,omitempty"`
Email          string `xml:"Email,omitempty"`
Position       string `xml:"Position,omitempty"`
Avatar         string `xml:"Avatar,omitempty"`
Gender         int    `xml:"Gender,omitempty"`
Status         int    `xml:"Status,omitempty"`
IsLeaderInDept string `xml:"IsLeaderInDept,omitempty"`
ExtAttr        string `xml:"ExtAttr,omitempty"`
// Department-related
ID             int    `xml:"Id,omitempty"`
ParentID       int    `xml:"ParentId,omitempty"`
Order          int    `xml:"Order,omitempty"`
// Tag-related
TagID          int    `xml:"TagId,omitempty"`
AddUserItems   string `xml:"AddUserItems,omitempty"`
DelUserItems   string `xml:"DelUserItems,omitempty"`
AddPartyItems  string `xml:"AddPartyItems,omitempty"`
DelPartyItems  string `xml:"DelPartyItems,omitempty"`
// External contact
State          string `xml:"State,omitempty"`
WelcomeCode    string `xml:"WelcomeCode,omitempty"`
```

### 3.4 DataEvent Interface and Base

```go
// DataEvent represents an event from the data callback URL.
type DataEvent interface {
    isDataEvent()
}

type dataEventBase struct {
    ToUserName string    // corp ID receiving the event
    AgentID    int       // application agent ID
    ReceiveAt  time.Time // time.Now() at parse time
}
```

### 3.5 DataEvent — Message Types

```go
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

// DataRawMsg is a fallback for unknown MsgType values.
type DataRawMsg struct {
    dataEventBase
    MsgType string
    RawXML  string
}
```

### 3.6 DataEvent — Event Types (MsgType=event)

```go
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
    OpenSpStatus  int            `xml:"OpenSpStatus"` // 1=pending 2=approved 3=rejected 4=revoked
    ApplyTime     int64          `xml:"ApplyTime"`
    ApplyUserID   string         `xml:"ApplyUserid"`
    ApplyUserName string         `xml:"ApplyUserName"`
    ApprovalNodes []ApprovalNode `xml:"ApprovalNodes>ApprovalNode"`
    NotifyNodes   []NotifyNode   `xml:"NotifyNodes>NotifyNode"`
}

// ApprovalNode represents one node in the approval chain.
type ApprovalNode struct {
    NodeStatus int            `xml:"NodeStatus"`
    NodeAttr   int            `xml:"NodeAttr"` // 1=or-sign 2=and-sign
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
    JobType string `xml:"JobType"` // sync_user, replace_user, invite_user, replace_party
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

// DataRawEvent is a fallback for unknown Event values.
type DataRawEvent struct {
    dataEventBase
    Event  string
    RawXML string
}
```

### 3.7 dataInner Struct

```go
type dataInner struct {
    XMLName      xml.Name
    ToUserName   string  `xml:"ToUserName"`
    AgentID      int     `xml:"AgentID"`
    MsgType      string  `xml:"MsgType"`
    Event        string  `xml:"Event,omitempty"`
    EventKey     string  `xml:"EventKey,omitempty"`
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
    ApprovalInfo   ApprovalInfo `xml:"ApprovalInfo,omitempty"`
    BatchJob       BatchJobResult `xml:"BatchJob,omitempty"`
}
```

---

## 4. Routing Logic

### 4.1 ParseNotify — Enhanced change_contact Case

```go
case "change_contact":
    base := baseEvent{SuiteID: inner.SuiteID, ReceiveAt: now}
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
        return &RawEvent{InfoType: inner.InfoType, RawXML: string(plain)}, nil
    }
```

### 4.2 ParseNotify — Enhanced change_external_contact Case

```go
case "change_external_contact":
    base := baseEvent{SuiteID: inner.SuiteID, ReceiveAt: now}
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
        return &RawEvent{InfoType: inner.InfoType, RawXML: string(plain)}, nil
    }
```

### 4.3 ParseDataNotify — Full Routing

```go
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

### 4.4 decryptNotify Helper

Extracted from existing ParseNotify to share with ParseDataNotify:

```go
// decryptNotify reads, verifies, and decrypts a callback request body.
func (c *Client) decryptNotify(r *http.Request) ([]byte, error) {
    body, err := io.ReadAll(r.Body)
    if err != nil {
        return nil, fmt.Errorf("isv: read body: %w", err)
    }

    var env componentEnvelope
    if err := xml.Unmarshal(body, &env); err != nil {
        return nil, fmt.Errorf("isv: unmarshal envelope: %w", err)
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

`ParseNotify` refactored to call `decryptNotify` then unmarshal `componentInner` and route.

---

## 5. Testing

### 5.1 Existing Tests to Modify

`notify_test.go` — replace the existing `change_contact` and `change_external_contact` test cases with per-ChangeType tests.

### 5.2 New Tests for Enhanced ParseNotify

| # | Test Function | Validates |
|---|---|---|
| 1 | TestParseNotify_ContactCreateUser | ChangeType=create_user, all user fields populated |
| 2 | TestParseNotify_ContactUpdateUser | ChangeType=update_user, NewUserID present |
| 3 | TestParseNotify_ContactDeleteUser | ChangeType=delete_user, minimal fields |
| 4 | TestParseNotify_ContactCreateParty | ChangeType=create_party, ID/Name/ParentID/Order |
| 5 | TestParseNotify_ContactUpdateParty | ChangeType=update_party |
| 6 | TestParseNotify_ContactDeleteParty | ChangeType=delete_party |
| 7 | TestParseNotify_ContactUpdateTag | ChangeType=update_tag, AddUserItems/DelUserItems |
| 8 | TestParseNotify_ExtContactAdd | ChangeType=add_external_contact, State/WelcomeCode |
| 9 | TestParseNotify_ExtContactEdit | ChangeType=edit_external_contact |
| 10 | TestParseNotify_ExtContactDel | ChangeType=del_external_contact |
| 11 | TestParseNotify_ExtContactDelFollow | ChangeType=del_follow_user |
| 12 | TestParseNotify_ExtContactAddHalf | ChangeType=add_half_external_contact |

### 5.3 New Tests for ParseDataNotify

| # | Test Function | Validates |
|---|---|---|
| 13 | TestParseDataNotify_TextMsg | MsgType=text, MsgID/Content |
| 14 | TestParseDataNotify_ImageMsg | MsgType=image, PicURL/MediaID |
| 15 | TestParseDataNotify_VoiceMsg | MsgType=voice, MediaID/Format |
| 16 | TestParseDataNotify_VideoMsg | MsgType=video, MediaID/ThumbMediaID |
| 17 | TestParseDataNotify_LocationMsg | MsgType=location, Lat/Lng/Scale/Label |
| 18 | TestParseDataNotify_LinkMsg | MsgType=link, Title/Description/URL |
| 19 | TestParseDataNotify_EnterAgent | Event=enter_agent, EventKey |
| 20 | TestParseDataNotify_MenuClick | Event=click, EventKey |
| 21 | TestParseDataNotify_MenuView | Event=view, EventKey/URL |
| 22 | TestParseDataNotify_ApprovalChange | Event=open_approval_change, ApprovalInfo nested fields |
| 23 | TestParseDataNotify_BatchJobResult | Event=batch_job_result, BatchJob nested fields |
| 24 | TestParseDataNotify_ExtContactChange | Event=change_external_contact, ChangeType/UserID |
| 25 | TestParseDataNotify_RawMsg | Unknown MsgType → DataRawMsg with RawXML |
| 26 | TestParseDataNotify_RawEvent | Unknown Event → DataRawEvent with RawXML |
| 27 | TestParseDataNotify_InvalidSignature | Bad msg_signature → error |
| 28 | TestParseDataNotify_InvalidXML | Corrupted XML → error |

### 5.4 Test Helper

New `buildDataNotifyRequest` helper, similar to existing `buildNotifyRequest`:

```go
func buildDataNotifyRequest(t *testing.T, crypto *wxcrypto.MsgCrypto, xmlBody string) *http.Request
```

Encrypts the XML body, wraps in envelope, computes signature, sets query params.

### 5.5 Coverage Target

≥ 85% for the package after changes.

---

## 6. Implementation Order

1. Extend `componentInner` with new fields + add 12 new Event structs to `struct.notify.go`
2. Extract `decryptNotify` helper from `ParseNotify` in `notify.go`
3. Add ChangeType secondary routing in `ParseNotify` for change_contact and change_external_contact
4. Update `notify_test.go` with 12 new ChangeType tests
5. Create `struct.data_notify.go` with DataEvent interface and all DTOs
6. Create `data_notify.go` with `ParseDataNotify` and routing logic
7. Create `data_notify_test.go` with ~16 tests
