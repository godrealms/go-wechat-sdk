package isv

import "time"

// Event 是所有回调事件的统一接口。
type Event interface {
	isEvent()
}

// baseEvent 被所有具体事件嵌入,承载通用字段。
type baseEvent struct {
	SuiteID   string
	ReceiveAt time.Time
}

func (baseEvent) isEvent() {}

// SuiteTicketEvent —— 微信每 10 分钟推送一次,本包已自动持久化到 Store。
type SuiteTicketEvent struct {
	baseEvent
	SuiteTicket string
}

// CreateAuthEvent —— 企业授权成功。
type CreateAuthEvent struct {
	baseEvent
	AuthCode string
}

// ChangeAuthEvent —— 企业变更授权。
type ChangeAuthEvent struct {
	baseEvent
	AuthCorpID string
}

// CancelAuthEvent —— 企业取消授权。
type CancelAuthEvent struct {
	baseEvent
	AuthCorpID string
}

// ResetPermanentCodeEvent —— 重置永久授权码。
type ResetPermanentCodeEvent struct {
	baseEvent
	AuthCorpID string
}

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

// ShareAgentChangeEvent —— 共享应用变更。
type ShareAgentChangeEvent struct {
	baseEvent
	AuthCorpID string
	AgentID    string
}

// ChangeAppAdminEvent —— 应用管理员变更。
type ChangeAppAdminEvent struct {
	baseEvent
	AuthCorpID string
	UserID     string
	IsAdmin    bool
}

// RawEvent 是未知 InfoType 的兜底,调用方可以自行 unmarshal。
type RawEvent struct {
	baseEvent
	InfoType string
	RawXML   string
}
