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

// ChangeContactEvent —— 通讯录变更(成员/部门/标签)。
type ChangeContactEvent struct {
	baseEvent
	AuthCorpID string
	ChangeType string // create_user / update_user / delete_user / create_party / update_party / delete_party / update_tag
	UserID     string
	Name       string
	Department string
	NewUserID  string // 仅 update_user 在 userid 变更时出现
}

// ChangeExternalContactEvent —— 外部联系人变更。
type ChangeExternalContactEvent struct {
	baseEvent
	AuthCorpID     string
	ChangeType     string
	UserID         string
	ExternalUserID string
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
