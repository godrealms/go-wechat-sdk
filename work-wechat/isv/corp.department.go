package isv

import (
	"context"
	"net/url"
	"strconv"
)

func (cc *CorpClient) CreateDepartment(ctx context.Context, req *CreateDeptReq) (*CreateDeptResp, error) {
	var resp CreateDeptResp
	if err := cc.doPost(ctx, "/cgi-bin/department/create", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (cc *CorpClient) UpdateDepartment(ctx context.Context, req *UpdateDeptReq) error {
	return cc.doPost(ctx, "/cgi-bin/department/update", req, nil)
}

func (cc *CorpClient) DeleteDepartment(ctx context.Context, id int) error {
	extra := url.Values{"id": {strconv.Itoa(id)}}
	return cc.doGet(ctx, "/cgi-bin/department/delete", extra, nil)
}

func (cc *CorpClient) ListDepartment(ctx context.Context, id int) ([]Department, error) {
	extra := url.Values{"id": {strconv.Itoa(id)}}
	var resp struct {
		Department []Department `json:"department"`
	}
	if err := cc.doGet(ctx, "/cgi-bin/department/list", extra, &resp); err != nil {
		return nil, err
	}
	return resp.Department, nil
}
