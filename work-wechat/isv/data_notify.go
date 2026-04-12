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
	ChangeType     string         `xml:"ChangeType,omitempty"`
	UserID         string         `xml:"UserID,omitempty"`
	ExternalUserID string         `xml:"ExternalUserID,omitempty"`
	State          string         `xml:"State,omitempty"`
	WelcomeCode    string         `xml:"WelcomeCode,omitempty"`
	ApprovalInfo   ApprovalInfo   `xml:"ApprovalInfo"`
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
