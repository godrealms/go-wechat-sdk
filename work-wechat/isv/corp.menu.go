package isv

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// CreateMenu creates a custom menu for the specified agent.
func (cc *CorpClient) CreateMenu(ctx context.Context, agentID int, req *CreateMenuReq) error {
	if err := requirePositive("CreateMenu", "agentID", agentID); err != nil {
		return err
	}
	if req == nil {
		return fmt.Errorf("isv: CreateMenu: req is required")
	}
	extra := url.Values{"agentid": {strconv.Itoa(agentID)}}
	return cc.doPostExtra(ctx, "/cgi-bin/menu/create", extra, req, nil)
}

// GetMenu retrieves the current custom menu for the specified agent.
func (cc *CorpClient) GetMenu(ctx context.Context, agentID int) (*MenuResp, error) {
	if err := requirePositive("GetMenu", "agentID", agentID); err != nil {
		return nil, err
	}
	extra := url.Values{"agentid": {strconv.Itoa(agentID)}}
	var resp MenuResp
	if err := cc.doGet(ctx, "/cgi-bin/menu/get", extra, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteMenu deletes the custom menu for the specified agent.
func (cc *CorpClient) DeleteMenu(ctx context.Context, agentID int) error {
	if err := requirePositive("DeleteMenu", "agentID", agentID); err != nil {
		return err
	}
	extra := url.Values{"agentid": {strconv.Itoa(agentID)}}
	return cc.doGet(ctx, "/cgi-bin/menu/delete", extra, nil)
}
