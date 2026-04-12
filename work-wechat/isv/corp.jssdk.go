package isv

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/url"
)

// GetJSAPITicket 获取企业的 jsapi_ticket（用于 wx.config 签名）。
func (cc *CorpClient) GetJSAPITicket(ctx context.Context) (*JSAPITicketResp, error) {
	var resp JSAPITicketResp
	if err := cc.doGet(ctx, "/cgi-bin/get_jsapi_ticket", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetAgentConfigTicket 获取应用的 jsapi_ticket（用于 wx.agentConfig 签名）。
func (cc *CorpClient) GetAgentConfigTicket(ctx context.Context) (*JSAPITicketResp, error) {
	extra := url.Values{"type": {"agent_config"}}
	var resp JSAPITicketResp
	if err := cc.doGet(ctx, "/cgi-bin/ticket/get", extra, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SignJSAPI 计算 JS-SDK 签名。纯计算函数，不发网络请求。
func SignJSAPI(ticket, nonceStr, timestamp, pageURL string) string {
	s := fmt.Sprintf("jsapi_ticket=%s&noncestr=%s&timestamp=%s&url=%s",
		ticket, nonceStr, timestamp, pageURL)
	h := sha1.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}
