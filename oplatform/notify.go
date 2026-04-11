package oplatform

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
)

// componentEnvelope 外层加密信封。
type componentEnvelope struct {
	XMLName    xml.Name `xml:"xml"`
	ToUserName string   `xml:"ToUserName"`
	Encrypt    string   `xml:"Encrypt"`
}

// componentInner 解密后的内层 XML。
type componentInner struct {
	XMLName                      xml.Name `xml:"xml"`
	AppID                        string   `xml:"AppId"`
	CreateTime                   int64    `xml:"CreateTime"`
	InfoType                     string   `xml:"InfoType"`
	ComponentVerifyTicket        string   `xml:"ComponentVerifyTicket"`
	AuthorizerAppID              string   `xml:"AuthorizerAppid"`
	AuthorizationCode            string   `xml:"AuthorizationCode"`
	AuthorizationCodeExpiredTime int64    `xml:"AuthorizationCodeExpiredTime"`
	PreAuthCode                  string   `xml:"PreAuthCode"`
}

// ParseNotify 解析开放平台第三方平台推送的回调。
//
//	r       - 原始 *http.Request；query 必须带 msg_signature/timestamp/nonce
//	rawBody - 可选：若调用方已经读过 r.Body，可以把原始字节从这里传入；
//	          若为 nil，本方法会从 r.Body 读取
//
// 成功返回 *ComponentNotify；当 InfoType == component_verify_ticket 时，
// SDK 会自动把 ticket 写入 Store，调用方无需再处理。
func (c *Client) ParseNotify(r *http.Request, rawBody []byte) (*ComponentNotify, error) {
	if r == nil {
		return nil, fmt.Errorf("oplatform: nil request")
	}
	q := r.URL.Query()

	if rawBody == nil {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, fmt.Errorf("oplatform: read body: %w", err)
		}
		_ = r.Body.Close()
		rawBody = body
	}

	var env componentEnvelope
	if err := xml.Unmarshal(rawBody, &env); err != nil {
		return nil, fmt.Errorf("oplatform: parse envelope: %w", err)
	}
	if env.Encrypt == "" {
		return nil, fmt.Errorf("oplatform: empty Encrypt field")
	}
	if !c.crypto.VerifySignature(q.Get("msg_signature"), q.Get("timestamp"), q.Get("nonce"), env.Encrypt) {
		return nil, fmt.Errorf("oplatform: msg_signature invalid")
	}
	plain, _, err := c.crypto.Decrypt(env.Encrypt)
	if err != nil {
		return nil, fmt.Errorf("oplatform: decrypt: %w", err)
	}

	var inner componentInner
	if err := xml.Unmarshal(plain, &inner); err != nil {
		return nil, fmt.Errorf("oplatform: parse inner xml: %w", err)
	}

	notify := &ComponentNotify{
		AppID:                        inner.AppID,
		CreateTime:                   inner.CreateTime,
		InfoType:                     inner.InfoType,
		ComponentVerifyTicket:        inner.ComponentVerifyTicket,
		AuthorizerAppID:              inner.AuthorizerAppID,
		AuthorizationCode:            inner.AuthorizationCode,
		AuthorizationCodeExpiredTime: inner.AuthorizationCodeExpiredTime,
		PreAuthCode:                  inner.PreAuthCode,
		Raw:                          plain,
	}

	if notify.InfoType == "component_verify_ticket" && notify.ComponentVerifyTicket != "" {
		if err := c.store.SetVerifyTicket(context.Background(), notify.ComponentVerifyTicket); err != nil {
			return nil, fmt.Errorf("oplatform: store set verify ticket: %w", err)
		}
	}

	return notify, nil
}
