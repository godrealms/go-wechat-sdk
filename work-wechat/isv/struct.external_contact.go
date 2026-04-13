package isv

// ExternalContact holds the details of an external contact (customer).
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

// FollowUser holds the information of an internal user who follows an external contact.
type FollowUser struct {
	UserID      string      `json:"userid"`
	Remark      string      `json:"remark"`
	Description string      `json:"description"`
	CreateTime  int64       `json:"createtime"`
	State       string      `json:"state"`
	Tags        []FollowTag `json:"tags"`
}

// FollowTag represents a tag that a follow user has applied to an external contact.
type FollowTag struct {
	GroupName string `json:"group_name"`
	TagName   string `json:"tag_name"`
	Type      int    `json:"type"`
}

// GetExternalContactResp is the response from GetExternalContact.
type GetExternalContactResp struct {
	ExternalContact ExternalContact `json:"external_contact"`
	FollowUser      []FollowUser    `json:"follow_user"`
}

// ListExternalContactResp is the response from ListExternalContact.
type ListExternalContactResp struct {
	ExternalUserID []string `json:"external_userid"`
}

// BatchGetExternalContactReq is the request for BatchGetExternalContactByUser.
type BatchGetExternalContactReq struct {
	UserIDList []string `json:"userid_list"`
	Cursor     string   `json:"cursor,omitempty"`
	Limit      int      `json:"limit,omitempty"`
}

// BatchGetExternalContactResp is the response from BatchGetExternalContactByUser.
type BatchGetExternalContactResp struct {
	ExternalContactList []GetExternalContactResp `json:"external_contact_list"`
	NextCursor          string                   `json:"next_cursor"`
}

// RemarkExternalContactReq is the request for RemarkExternalContact.
type RemarkExternalContactReq struct {
	UserID           string   `json:"userid"`
	ExternalUserID   string   `json:"external_userid"`
	Remark           string   `json:"remark,omitempty"`
	Description      string   `json:"description,omitempty"`
	RemarkCompany    string   `json:"remark_company,omitempty"`
	RemarkMobiles    []string `json:"remark_mobiles,omitempty"`
	RemarkPicMediaID string   `json:"remark_pic_mediaid,omitempty"`
}

// CorpTag represents an enterprise customer tag.
type CorpTag struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Order int    `json:"order"`
}

// CorpTagGroup represents a group of enterprise customer tags.
type CorpTagGroup struct {
	GroupID   string    `json:"group_id"`
	GroupName string    `json:"group_name"`
	Tag       []CorpTag `json:"tag"`
}

// GetCorpTagListReq is the request for GetCorpTagList.
type GetCorpTagListReq struct {
	TagID   []string `json:"tag_id,omitempty"`
	GroupID []string `json:"group_id,omitempty"`
}

// GetCorpTagListResp is the response from GetCorpTagList.
type GetCorpTagListResp struct {
	TagGroup []CorpTagGroup `json:"tag_group"`
}

// CorpTagInput holds the name and display order for a tag to create.
type CorpTagInput struct {
	Name  string `json:"name"`
	Order int    `json:"order,omitempty"`
}

// AddCorpTagReq is the request for AddCorpTag.
type AddCorpTagReq struct {
	GroupID   string         `json:"group_id,omitempty"`
	GroupName string         `json:"group_name,omitempty"`
	Tag       []CorpTagInput `json:"tag"`
}

// AddCorpTagResp is the response from AddCorpTag.
type AddCorpTagResp struct {
	TagGroup CorpTagGroup `json:"tag_group"`
}

// EditCorpTagReq is the request for EditCorpTag.
type EditCorpTagReq struct {
	ID    string `json:"id"`
	Name  string `json:"name,omitempty"`
	Order *int   `json:"order,omitempty"`
}

// DelCorpTagReq is the request for DelCorpTag.
type DelCorpTagReq struct {
	TagID   []string `json:"tag_id,omitempty"`
	GroupID []string `json:"group_id,omitempty"`
}

// MarkTagReq is the request for MarkTag.
type MarkTagReq struct {
	UserID         string   `json:"userid"`
	ExternalUserID string   `json:"external_userid"`
	AddTag         []string `json:"add_tag,omitempty"`
	RemoveTag      []string `json:"remove_tag,omitempty"`
}

// FollowUserListResp is the response from GetFollowUserList.
type FollowUserListResp struct {
	FollowUser []string `json:"follow_user"`
}
