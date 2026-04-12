package isv

import (
	"context"
	"net/url"
	"strconv"
)

func boolToStr(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// CreateUser creates a user on behalf of the authorized enterprise.
func (cc *CorpClient) CreateUser(ctx context.Context, req *CreateUserReq) error {
	return cc.doPost(ctx, "/cgi-bin/user/create", req, nil)
}

// UpdateUser updates a user on behalf of the authorized enterprise.
func (cc *CorpClient) UpdateUser(ctx context.Context, req *UpdateUserReq) error {
	return cc.doPost(ctx, "/cgi-bin/user/update", req, nil)
}

// DeleteUser deletes a user by userid on behalf of the authorized enterprise.
func (cc *CorpClient) DeleteUser(ctx context.Context, userID string) error {
	extra := url.Values{"userid": {userID}}
	return cc.doGet(ctx, "/cgi-bin/user/delete", extra, nil)
}

// GetUser retrieves detailed user info by userid.
func (cc *CorpClient) GetUser(ctx context.Context, userID string) (*UserDetail, error) {
	extra := url.Values{"userid": {userID}}
	var resp UserDetail
	if err := cc.doGet(ctx, "/cgi-bin/user/get", extra, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListUserSimple lists users in a department with basic info (userid + name).
func (cc *CorpClient) ListUserSimple(ctx context.Context, deptID int, fetchChild bool) (*UserSimpleListResp, error) {
	extra := url.Values{
		"department_id": {strconv.Itoa(deptID)},
		"fetch_child":   {boolToStr(fetchChild)},
	}
	var resp UserSimpleListResp
	if err := cc.doGet(ctx, "/cgi-bin/user/simplelist", extra, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListUserDetail lists users in a department with full detail.
func (cc *CorpClient) ListUserDetail(ctx context.Context, deptID int, fetchChild bool) (*UserDetailListResp, error) {
	extra := url.Values{
		"department_id": {strconv.Itoa(deptID)},
		"fetch_child":   {boolToStr(fetchChild)},
	}
	var resp UserDetailListResp
	if err := cc.doGet(ctx, "/cgi-bin/user/list", extra, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
