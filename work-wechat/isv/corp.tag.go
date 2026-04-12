package isv

import (
	"context"
	"net/url"
	"strconv"
)

// CreateTag creates a tag on behalf of the authorized enterprise.
func (cc *CorpClient) CreateTag(ctx context.Context, req *CreateTagReq) (*CreateTagResp, error) {
	var resp CreateTagResp
	if err := cc.doPost(ctx, "/cgi-bin/tag/create", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateTag updates a tag on behalf of the authorized enterprise.
func (cc *CorpClient) UpdateTag(ctx context.Context, req *UpdateTagReq) error {
	return cc.doPost(ctx, "/cgi-bin/tag/update", req, nil)
}

// DeleteTag deletes a tag by ID on behalf of the authorized enterprise.
func (cc *CorpClient) DeleteTag(ctx context.Context, tagID int) error {
	extra := url.Values{"tagid": {strconv.Itoa(tagID)}}
	return cc.doGet(ctx, "/cgi-bin/tag/delete", extra, nil)
}

// ListTag lists all tags of the authorized enterprise.
func (cc *CorpClient) ListTag(ctx context.Context) ([]Tag, error) {
	var resp struct {
		TagList []Tag `json:"taglist"`
	}
	if err := cc.doGet(ctx, "/cgi-bin/tag/list", nil, &resp); err != nil {
		return nil, err
	}
	return resp.TagList, nil
}

// GetTagUsers retrieves the user and department list for a given tag.
func (cc *CorpClient) GetTagUsers(ctx context.Context, tagID int) (*TagUsersResp, error) {
	extra := url.Values{"tagid": {strconv.Itoa(tagID)}}
	var resp TagUsersResp
	if err := cc.doGet(ctx, "/cgi-bin/tag/get", extra, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AddTagUsers adds users and departments to a tag.
func (cc *CorpClient) AddTagUsers(ctx context.Context, req *TagUsersModifyReq) (*TagUsersModifyResp, error) {
	var resp TagUsersModifyResp
	if err := cc.doPost(ctx, "/cgi-bin/tag/addtagusers", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DelTagUsers removes users and departments from a tag.
func (cc *CorpClient) DelTagUsers(ctx context.Context, req *TagUsersModifyReq) (*TagUsersModifyResp, error) {
	var resp TagUsersModifyResp
	if err := cc.doPost(ctx, "/cgi-bin/tag/deltagusers", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
