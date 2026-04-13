package isv

import (
	"context"
	"net/url"
)

// GetExternalContact retrieves the details of an external contact by their user ID.
func (cc *CorpClient) GetExternalContact(ctx context.Context, externalUserID string) (*GetExternalContactResp, error) {
	extra := url.Values{"external_userid": {externalUserID}}
	var resp GetExternalContactResp
	if err := cc.doGet(ctx, "/cgi-bin/externalcontact/get", extra, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListExternalContact retrieves the list of external contacts for a given internal user.
func (cc *CorpClient) ListExternalContact(ctx context.Context, userID string) (*ListExternalContactResp, error) {
	extra := url.Values{"userid": {userID}}
	var resp ListExternalContactResp
	if err := cc.doGet(ctx, "/cgi-bin/externalcontact/list", extra, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// BatchGetExternalContactByUser retrieves external contact details in bulk for multiple internal users.
func (cc *CorpClient) BatchGetExternalContactByUser(ctx context.Context, req *BatchGetExternalContactReq) (*BatchGetExternalContactResp, error) {
	var resp BatchGetExternalContactResp
	if err := cc.doPost(ctx, "/cgi-bin/externalcontact/batch/get_by_user", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// RemarkExternalContact updates the remark information for an external contact.
func (cc *CorpClient) RemarkExternalContact(ctx context.Context, req *RemarkExternalContactReq) error {
	return cc.doPost(ctx, "/cgi-bin/externalcontact/remark", req, nil)
}

// GetCorpTagList retrieves the enterprise customer tag library.
func (cc *CorpClient) GetCorpTagList(ctx context.Context, req *GetCorpTagListReq) (*GetCorpTagListResp, error) {
	var resp GetCorpTagListResp
	if err := cc.doPost(ctx, "/cgi-bin/externalcontact/get_corp_tag_list", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AddCorpTag adds enterprise customer tags.
func (cc *CorpClient) AddCorpTag(ctx context.Context, req *AddCorpTagReq) (*AddCorpTagResp, error) {
	var resp AddCorpTagResp
	if err := cc.doPost(ctx, "/cgi-bin/externalcontact/add_corp_tag", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// EditCorpTag edits an enterprise customer tag.
func (cc *CorpClient) EditCorpTag(ctx context.Context, req *EditCorpTagReq) error {
	return cc.doPost(ctx, "/cgi-bin/externalcontact/edit_corp_tag", req, nil)
}

// DelCorpTag deletes enterprise customer tags.
func (cc *CorpClient) DelCorpTag(ctx context.Context, req *DelCorpTagReq) error {
	return cc.doPost(ctx, "/cgi-bin/externalcontact/del_corp_tag", req, nil)
}

// MarkTag adds or removes enterprise tags on an external contact.
func (cc *CorpClient) MarkTag(ctx context.Context, req *MarkTagReq) error {
	return cc.doPost(ctx, "/cgi-bin/externalcontact/mark_tag", req, nil)
}

// GetFollowUserList retrieves the list of members who have the customer contact feature configured.
func (cc *CorpClient) GetFollowUserList(ctx context.Context) (*FollowUserListResp, error) {
	var resp FollowUserListResp
	if err := cc.doGet(ctx, "/cgi-bin/externalcontact/get_follow_user_list", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
