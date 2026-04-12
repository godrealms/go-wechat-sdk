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
