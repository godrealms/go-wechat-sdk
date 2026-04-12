package isv

// ExternalContact 客户详情。
type ExternalContact struct {
	ExternalUserID string `json:"external_userid"`
	Name           string `json:"name"`
	Position       string `json:"position"`
	Avatar         string `json:"avatar"`
	CorpName       string `json:"corp_name"`
	CorpFullName   string `json:"corp_full_name"`
	Type           int    `json:"type"`
	Gender         int    `json:"gender"`
	UnionID        string `json:"unionid"`
}

// FollowUser 跟进人信息。
type FollowUser struct {
	UserID      string      `json:"userid"`
	Remark      string      `json:"remark"`
	Description string      `json:"description"`
	CreateTime  int64       `json:"createtime"`
	State       string      `json:"state"`
	Tags        []FollowTag `json:"tags"`
}

// FollowTag 跟进人给客户打的标签。
type FollowTag struct {
	GroupName string `json:"group_name"`
	TagName   string `json:"tag_name"`
	Type      int    `json:"type"`
}

// GetExternalContactResp 获取客户详情响应。
type GetExternalContactResp struct {
	ExternalContact ExternalContact `json:"external_contact"`
	FollowUser      []FollowUser    `json:"follow_user"`
}

// ListExternalContactResp 获取客户列表响应。
type ListExternalContactResp struct {
	ExternalUserID []string `json:"external_userid"`
}

// BatchGetExternalContactReq 批量获取客户详情请求。
type BatchGetExternalContactReq struct {
	UserIDList []string `json:"userid_list"`
	Cursor     string   `json:"cursor,omitempty"`
	Limit      int      `json:"limit,omitempty"`
}

// BatchGetExternalContactResp 批量获取客户详情响应。
type BatchGetExternalContactResp struct {
	ExternalContactList []GetExternalContactResp `json:"external_contact_list"`
	NextCursor          string                   `json:"next_cursor"`
}

// RemarkExternalContactReq 修改客户备注请求。
type RemarkExternalContactReq struct {
	UserID           string   `json:"userid"`
	ExternalUserID   string   `json:"external_userid"`
	Remark           string   `json:"remark,omitempty"`
	Description      string   `json:"description,omitempty"`
	RemarkCompany    string   `json:"remark_company,omitempty"`
	RemarkMobiles    []string `json:"remark_mobiles,omitempty"`
	RemarkPicMediaID string   `json:"remark_pic_mediaid,omitempty"`
}

// CorpTag 企业客户标签。
type CorpTag struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Order int    `json:"order"`
}

// CorpTagGroup 企业客户标签组。
type CorpTagGroup struct {
	GroupID   string    `json:"group_id"`
	GroupName string    `json:"group_name"`
	Tag       []CorpTag `json:"tag"`
}

// GetCorpTagListReq 获取标签库请求。
type GetCorpTagListReq struct {
	TagID   []string `json:"tag_id,omitempty"`
	GroupID []string `json:"group_id,omitempty"`
}

// GetCorpTagListResp 获取标签库响应。
type GetCorpTagListResp struct {
	TagGroup []CorpTagGroup `json:"tag_group"`
}

// AddCorpTagReq 添加企业客户标签请求。
type AddCorpTagReq struct {
	GroupID   string `json:"group_id,omitempty"`
	GroupName string `json:"group_name,omitempty"`
	Tag       []struct {
		Name  string `json:"name"`
		Order int    `json:"order,omitempty"`
	} `json:"tag"`
}

// AddCorpTagResp 添加企业客户标签响应。
type AddCorpTagResp struct {
	TagGroup CorpTagGroup `json:"tag_group"`
}

// EditCorpTagReq 编辑企业客户标签请求。
type EditCorpTagReq struct {
	ID    string `json:"id"`
	Name  string `json:"name,omitempty"`
	Order *int   `json:"order,omitempty"`
}

// DelCorpTagReq 删除企业客户标签请求。
type DelCorpTagReq struct {
	TagID   []string `json:"tag_id,omitempty"`
	GroupID []string `json:"group_id,omitempty"`
}

// MarkTagReq 编辑客户企业标签请求。
type MarkTagReq struct {
	UserID         string   `json:"userid"`
	ExternalUserID string   `json:"external_userid"`
	AddTag         []string `json:"add_tag,omitempty"`
	RemoveTag      []string `json:"remove_tag,omitempty"`
}

// FollowUserListResp 获取配置了客户联系功能的成员列表响应。
type FollowUserListResp struct {
	FollowUser []string `json:"follow_user"`
}
