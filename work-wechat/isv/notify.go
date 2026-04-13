package isv

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// componentEnvelope 是外层加密信封 XML。
type componentEnvelope struct {
	XMLName xml.Name `xml:"xml"`
	Encrypt string   `xml:"Encrypt"`
}

// componentInner 是解密后的内层 XML,承载所有 InfoType 的字段。
// 对于不同 InfoType,只有部分字段有值;用 omitempty 可避免 unmarshal 报错。
type componentInner struct {
	XMLName        xml.Name `xml:"xml"`
	SuiteID        string   `xml:"SuiteId"`
	InfoType       string   `xml:"InfoType"`
	TimeStamp      int64    `xml:"TimeStamp"`
	SuiteTicket    string   `xml:"SuiteTicket,omitempty"`
	AuthCode       string   `xml:"AuthCode,omitempty"`
	AuthCorpID     string   `xml:"AuthCorpId,omitempty"`
	ChangeType     string   `xml:"ChangeType,omitempty"`
	UserID         string   `xml:"UserID,omitempty"`
	Name           string   `xml:"Name,omitempty"`
	Department     string   `xml:"Department,omitempty"`
	NewUserID      string   `xml:"NewUserID,omitempty"`
	ExternalUserID string   `xml:"ExternalUserID,omitempty"`
	AgentID        string   `xml:"AgentID,omitempty"`
	IsAdmin        int      `xml:"IsAdmin,omitempty"`
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
}

// decryptNotify reads, verifies the signature of, and decrypts the WeChat
// notification body from r. It closes r.Body before returning.
func (c *Client) decryptNotify(r *http.Request) ([]byte, error) {
	defer r.Body.Close()
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

// ParseNotify verifies, decrypts, and parses a WeChat Work ISV callback, returning a
// strongly typed Event. For suite_ticket events it automatically calls Store.PutSuiteTicket.
// Unknown InfoType values are returned as *RawEvent without error.
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
