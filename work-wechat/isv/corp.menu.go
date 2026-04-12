package isv

import (
	"context"
	"net/url"
	"strconv"
)

// CreateMenu creates a custom menu for the specified agent.
func (cc *CorpClient) CreateMenu(ctx context.Context, agentID int, req *CreateMenuReq) error {
	extra := url.Values{"agentid": {strconv.Itoa(agentID)}}
	return cc.doPostExtra(ctx, "/cgi-bin/menu/create", extra, req, nil)
}

// GetMenu retrieves the current custom menu for the specified agent.
func (cc *CorpClient) GetMenu(ctx context.Context, agentID int) (*MenuResp, error) {
	extra := url.Values{"agentid": {strconv.Itoa(agentID)}}
	var resp MenuResp
	if err := cc.doGet(ctx, "/cgi-bin/menu/get", extra, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteMenu deletes the custom menu for the specified agent.
func (cc *CorpClient) DeleteMenu(ctx context.Context, agentID int) error {
	extra := url.Values{"agentid": {strconv.Itoa(agentID)}}
	return cc.doGet(ctx, "/cgi-bin/menu/delete", extra, nil)
}
