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
