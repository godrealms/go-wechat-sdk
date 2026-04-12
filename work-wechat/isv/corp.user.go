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

func (cc *CorpClient) CreateUser(ctx context.Context, req *CreateUserReq) error {
	return cc.doPost(ctx, "/cgi-bin/user/create", req, nil)
}

func (cc *CorpClient) UpdateUser(ctx context.Context, req *UpdateUserReq) error {
	return cc.doPost(ctx, "/cgi-bin/user/update", req, nil)
}

func (cc *CorpClient) DeleteUser(ctx context.Context, userID string) error {
	extra := url.Values{"userid": {userID}}
	return cc.doGet(ctx, "/cgi-bin/user/delete", extra, nil)
}

func (cc *CorpClient) GetUser(ctx context.Context, userID string) (*UserDetail, error) {
	extra := url.Values{"userid": {userID}}
	var resp UserDetail
	if err := cc.doGet(ctx, "/cgi-bin/user/get", extra, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

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
