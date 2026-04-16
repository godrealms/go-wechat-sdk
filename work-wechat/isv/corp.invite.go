package isv

import (
	"context"
	"fmt"
)

// InviteUser sends batch invitations to users, departments, or tags.
func (cc *CorpClient) InviteUser(ctx context.Context, req *InviteReq) (*InviteResp, error) {
	if req == nil {
		return nil, fmt.Errorf("isv: InviteUser: req is required")
	}
	var resp InviteResp
	if err := cc.doPost(ctx, "/cgi-bin/batch/invite", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
