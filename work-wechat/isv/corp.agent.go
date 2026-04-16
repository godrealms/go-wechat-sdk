package isv

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// GetAgent retrieves the details of an agent (application) by ID.
func (cc *CorpClient) GetAgent(ctx context.Context, agentID int) (*AgentDetail, error) {
	if err := requirePositive("GetAgent", "agentID", agentID); err != nil {
		return nil, err
	}
	extra := url.Values{"agentid": {strconv.Itoa(agentID)}}
	var resp AgentDetail
	if err := cc.doGet(ctx, "/cgi-bin/agent/get", extra, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SetAgent updates an agent's properties (name, description, homepage, etc.).
func (cc *CorpClient) SetAgent(ctx context.Context, req *SetAgentReq) error {
	if req == nil {
		return fmt.Errorf("isv: SetAgent: req is required")
	}
	if err := requirePositive("SetAgent", "AgentID", req.AgentID); err != nil {
		return err
	}
	return cc.doPost(ctx, "/cgi-bin/agent/set", req, nil)
}
