package isv

import "context"

// InviteUser sends batch invitations to users, departments, or tags.
func (cc *CorpClient) InviteUser(ctx context.Context, req *InviteReq) (*InviteResp, error) {
	var resp InviteResp
	if err := cc.doPost(ctx, "/cgi-bin/batch/invite", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
