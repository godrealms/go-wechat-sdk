package isv

import (
	"context"
	"net/url"
	"strconv"
)

// GetAgent retrieves the details of an agent (application) by ID.
func (cc *CorpClient) GetAgent(ctx context.Context, agentID int) (*AgentDetail, error) {
	extra := url.Values{"agentid": {strconv.Itoa(agentID)}}
	var resp AgentDetail
	if err := cc.doGet(ctx, "/cgi-bin/agent/get", extra, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SetAgent updates an agent's properties (name, description, homepage, etc.).
func (cc *CorpClient) SetAgent(ctx context.Context, req *SetAgentReq) error {
	return cc.doPost(ctx, "/cgi-bin/agent/set", req, nil)
}
