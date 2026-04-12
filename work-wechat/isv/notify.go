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
}

// ParseNotify 校验、解密、解析企业微信回调,并返回强类型事件。
//
// 对 suite_ticket 事件本函数自动调用 Store.PutSuiteTicket。
// 对未知 InfoType 返回 *RawEvent(Task 10 引入),不报错。
func (c *Client) ParseNotify(r *http.Request) (Event, error) {
	ctx := r.Context()

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

	var inner componentInner
	if err := xml.Unmarshal(plain, &inner); err != nil {
		return nil, fmt.Errorf("isv: parse inner: %w", err)
	}

	base := baseEvent{SuiteID: inner.SuiteID, ReceiveAt: time.Now()}

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
	}
	return nil, errors.New("isv: unknown InfoType " + inner.InfoType) // Task 10 replaces with *RawEvent
}
