package isv

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/url"
)

// GetJSAPITicket retrieves the enterprise jsapi_ticket used to sign wx.config calls.
func (cc *CorpClient) GetJSAPITicket(ctx context.Context) (*JSAPITicketResp, error) {
	var resp JSAPITicketResp
	if err := cc.doGet(ctx, "/cgi-bin/get_jsapi_ticket", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetAgentConfigTicket retrieves the application jsapi_ticket used to sign wx.agentConfig calls.
func (cc *CorpClient) GetAgentConfigTicket(ctx context.Context) (*JSAPITicketResp, error) {
	extra := url.Values{"type": {"agent_config"}}
	var resp JSAPITicketResp
	if err := cc.doGet(ctx, "/cgi-bin/ticket/get", extra, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SignJSAPI computes the JS-SDK signature. It is a pure computation function and makes no network requests.
func SignJSAPI(ticket, nonceStr, timestamp, pageURL string) string {
	s := fmt.Sprintf("jsapi_ticket=%s&noncestr=%s&timestamp=%s&url=%s",
		ticket, nonceStr, timestamp, pageURL)
	h := sha1.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}
