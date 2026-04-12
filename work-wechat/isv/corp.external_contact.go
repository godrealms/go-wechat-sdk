package isv

import (
	"context"
	"net/url"
)

// GetExternalContact 获取客户详情。
func (cc *CorpClient) GetExternalContact(ctx context.Context, externalUserID string) (*GetExternalContactResp, error) {
	extra := url.Values{"external_userid": {externalUserID}}
	var resp GetExternalContactResp
	if err := cc.doGet(ctx, "/cgi-bin/externalcontact/get", extra, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListExternalContact 获取客户列表（按跟进人）。
func (cc *CorpClient) ListExternalContact(ctx context.Context, userID string) (*ListExternalContactResp, error) {
	extra := url.Values{"userid": {userID}}
	var resp ListExternalContactResp
	if err := cc.doGet(ctx, "/cgi-bin/externalcontact/list", extra, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// BatchGetExternalContactByUser 批量获取客户详情。
func (cc *CorpClient) BatchGetExternalContactByUser(ctx context.Context, req *BatchGetExternalContactReq) (*BatchGetExternalContactResp, error) {
	var resp BatchGetExternalContactResp
	if err := cc.doPost(ctx, "/cgi-bin/externalcontact/batch/get_by_user", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// RemarkExternalContact 修改客户备注信息。
func (cc *CorpClient) RemarkExternalContact(ctx context.Context, req *RemarkExternalContactReq) error {
	return cc.doPost(ctx, "/cgi-bin/externalcontact/remark", req, nil)
}

// GetCorpTagList 获取企业标签库。
func (cc *CorpClient) GetCorpTagList(ctx context.Context, req *GetCorpTagListReq) (*GetCorpTagListResp, error) {
	var resp GetCorpTagListResp
	if err := cc.doPost(ctx, "/cgi-bin/externalcontact/get_corp_tag_list", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AddCorpTag 添加企业客户标签。
func (cc *CorpClient) AddCorpTag(ctx context.Context, req *AddCorpTagReq) (*AddCorpTagResp, error) {
	var resp AddCorpTagResp
	if err := cc.doPost(ctx, "/cgi-bin/externalcontact/add_corp_tag", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// EditCorpTag 编辑企业客户标签。
func (cc *CorpClient) EditCorpTag(ctx context.Context, req *EditCorpTagReq) error {
	return cc.doPost(ctx, "/cgi-bin/externalcontact/edit_corp_tag", req, nil)
}

// DelCorpTag 删除企业客户标签。
func (cc *CorpClient) DelCorpTag(ctx context.Context, req *DelCorpTagReq) error {
	return cc.doPost(ctx, "/cgi-bin/externalcontact/del_corp_tag", req, nil)
}

// MarkTag 编辑客户企业标签（给客户打/取消标签）。
func (cc *CorpClient) MarkTag(ctx context.Context, req *MarkTagReq) error {
	return cc.doPost(ctx, "/cgi-bin/externalcontact/mark_tag", req, nil)
}

// GetFollowUserList 获取配置了客户联系功能的成员列表。
func (cc *CorpClient) GetFollowUserList(ctx context.Context) (*FollowUserListResp, error) {
	var resp FollowUserListResp
	if err := cc.doGet(ctx, "/cgi-bin/externalcontact/get_follow_user_list", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
