package isv

// ---------------------------------------------------------------------------
// 3.1 部门
// ---------------------------------------------------------------------------

// CreateDeptReq 创建部门请求。
type CreateDeptReq struct {
	Name     string `json:"name"`
	NameEn   string `json:"name_en,omitempty"`
	ParentID int    `json:"parentid"`
	Order    int    `json:"order,omitempty"`
	ID       int    `json:"id,omitempty"`
}

// CreateDeptResp 创建部门响应。
type CreateDeptResp struct {
	ID int `json:"id"`
}

// UpdateDeptReq 更新部门请求。
type UpdateDeptReq struct {
	ID       int    `json:"id"`
	Name     string `json:"name,omitempty"`
	NameEn   string `json:"name_en,omitempty"`
	ParentID int    `json:"parentid,omitempty"`
	Order    int    `json:"order,omitempty"`
}

// Department 部门信息（列表返回）。
type Department struct {
	ID               int      `json:"id"`
	Name             string   `json:"name"`
	NameEn           string   `json:"name_en"`
	DepartmentLeader []string `json:"department_leader"`
	ParentID         int      `json:"parentid"`
	Order            int      `json:"order"`
}

// ---------------------------------------------------------------------------
// 3.2 成员
// ---------------------------------------------------------------------------

// CreateUserReq 创建成员请求。
type CreateUserReq struct {
	UserID         string   `json:"userid"`
	Name           string   `json:"name"`
	Alias          string   `json:"alias,omitempty"`
	Mobile         string   `json:"mobile,omitempty"`
	Department     []int    `json:"department"`
	Order          []int    `json:"order,omitempty"`
	Position       string   `json:"position,omitempty"`
	Gender         string   `json:"gender,omitempty"`
	Email          string   `json:"email,omitempty"`
	BizMail        string   `json:"biz_mail,omitempty"`
	IsLeaderInDept []int    `json:"is_leader_in_dept,omitempty"`
	DirectLeader   []string `json:"direct_leader,omitempty"`
	Enable         int      `json:"enable,omitempty"`
	Telephone      string   `json:"telephone,omitempty"`
	Address        string   `json:"address,omitempty"`
	MainDepartment int      `json:"main_department,omitempty"`
	ToInvite       bool     `json:"to_invite,omitempty"`
}

// UpdateUserReq 更新成员请求。
type UpdateUserReq struct {
	UserID         string   `json:"userid"`
	Name           string   `json:"name,omitempty"`
	Alias          string   `json:"alias,omitempty"`
	Mobile         string   `json:"mobile,omitempty"`
	Department     []int    `json:"department,omitempty"`
	Order          []int    `json:"order,omitempty"`
	Position       string   `json:"position,omitempty"`
	Gender         string   `json:"gender,omitempty"`
	Email          string   `json:"email,omitempty"`
	BizMail        string   `json:"biz_mail,omitempty"`
	IsLeaderInDept []int    `json:"is_leader_in_dept,omitempty"`
	DirectLeader   []string `json:"direct_leader,omitempty"`
	Enable         int      `json:"enable,omitempty"`
	Telephone      string   `json:"telephone,omitempty"`
	Address        string   `json:"address,omitempty"`
	MainDepartment int      `json:"main_department,omitempty"`
}

// UserSimple 简单成员信息。
type UserSimple struct {
	UserID     string `json:"userid"`
	Name       string `json:"name"`
	Department []int  `json:"department"`
	OpenUserID string `json:"open_userid"`
}

// UserSimpleListResp simplelist 响应。
type UserSimpleListResp struct {
	UserList []UserSimple `json:"userlist"`
}

// UserDetail 详细成员信息（GetUser / ListUserDetail 共用）。
type UserDetail struct {
	UserID         string   `json:"userid"`
	Name           string   `json:"name"`
	Department     []int    `json:"department"`
	Order          []int    `json:"order"`
	Position       string   `json:"position"`
	Mobile         string   `json:"mobile"`
	Gender         string   `json:"gender"`
	Email          string   `json:"email"`
	BizMail        string   `json:"biz_mail"`
	IsLeaderInDept []int    `json:"is_leader_in_dept"`
	DirectLeader   []string `json:"direct_leader"`
	Avatar         string   `json:"avatar"`
	ThumbAvatar    string   `json:"thumb_avatar"`
	Telephone      string   `json:"telephone"`
	Alias          string   `json:"alias"`
	Address        string   `json:"address"`
	OpenUserID     string   `json:"open_userid"`
	MainDepartment int      `json:"main_department"`
	Status         int      `json:"status"`
	QRCode         string   `json:"qr_code"`
}

// UserDetailListResp user/list 响应。
type UserDetailListResp struct {
	UserList []UserDetail `json:"userlist"`
}

// ---------------------------------------------------------------------------
// 3.3 标签
// ---------------------------------------------------------------------------

// CreateTagReq 创建标签请求。
type CreateTagReq struct {
	TagName string `json:"tagname"`
	TagID   int    `json:"tagid,omitempty"`
}

// CreateTagResp 创建标签响应。
type CreateTagResp struct {
	TagID int `json:"tagid"`
}

// UpdateTagReq 更新标签请求。
type UpdateTagReq struct {
	TagID   int    `json:"tagid"`
	TagName string `json:"tagname"`
}

// Tag 标签信息。
type Tag struct {
	TagID   int    `json:"tagid"`
	TagName string `json:"tagname"`
}

// TagUser 标签成员。
type TagUser struct {
	UserID string `json:"userid"`
	Name   string `json:"name"`
}

// TagUsersResp tag/get 响应。
type TagUsersResp struct {
	TagName   string    `json:"tagname"`
	UserList  []TagUser `json:"userlist"`
	PartyList []int     `json:"partylist"`
}

// TagUsersModifyReq addtagusers / deltagusers 请求。
type TagUsersModifyReq struct {
	TagID     int      `json:"tagid"`
	UserList  []string `json:"userlist,omitempty"`
	PartyList []int    `json:"partylist,omitempty"`
}

// TagUsersModifyResp addtagusers / deltagusers 响应。
type TagUsersModifyResp struct {
	InvalidList  string `json:"invalidlist"`
	InvalidParty []int  `json:"invalidparty"`
}

// ---------------------------------------------------------------------------
// 3.4 邀请
// ---------------------------------------------------------------------------

// InviteReq batch/invite 请求。
type InviteReq struct {
	User  []string `json:"user,omitempty"`
	Party []int    `json:"party,omitempty"`
	Tag   []int    `json:"tag,omitempty"`
}

// InviteResp batch/invite 响应。
type InviteResp struct {
	InvalidUser  []string `json:"invaliduser"`
	InvalidParty []int    `json:"invalidparty"`
	InvalidTag   []int    `json:"invalidtag"`
}
