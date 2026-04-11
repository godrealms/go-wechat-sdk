package oplatform

// 本文件汇总代小程序开发管理 (WxaAdmin) 所有子族的请求/响应 DTO。
// 各子族（account/category/domain/tester/code/release）的结构体
// 按顺序追加到下面的分隔注释段。

// ----- account -----

type WxaSetNicknameReq struct {
	Nickname     string `json:"nick_name"`
	IDCard       string `json:"id_card,omitempty"`
	License      string `json:"license,omitempty"`
	NamingOther1 string `json:"naming_other_stuff_1,omitempty"`
	NamingOther2 string `json:"naming_other_stuff_2,omitempty"`
	NamingOther3 string `json:"naming_other_stuff_3,omitempty"`
	NamingOther4 string `json:"naming_other_stuff_4,omitempty"`
	NamingOther5 string `json:"naming_other_stuff_5,omitempty"`
}

type WxaSetNicknameResp struct {
	Wording string `json:"wording,omitempty"`
	AuditID string `json:"audit_id,omitempty"`
}

type WxaQueryNicknameResp struct {
	Nickname   string `json:"nickname"`
	AuditStat  int    `json:"audit_stat"`
	FailReason string `json:"fail_reason,omitempty"`
	CreateTime int64  `json:"create_time"`
	AuditTime  int64  `json:"audit_time"`
}

type WxaCheckNicknameResp struct {
	HitCondition bool   `json:"hit_condition"`
	Wording      string `json:"wording,omitempty"`
}

// ----- category -----

type WxaCategoryItem struct {
	First       int    `json:"first"`
	Second      int    `json:"second"`
	FirstName   string `json:"first_name,omitempty"`
	SecondName  string `json:"second_name,omitempty"`
	AuditStatus int    `json:"audit_status,omitempty"`
	AuditReason string `json:"audit_reason,omitempty"`
}

type WxaGetCategoryResp struct {
	CategoriesList []WxaCategoryItem `json:"categories_list"`
}

type WxaGetAllCategoriesResp struct {
	CategoriesList []struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Level    int    `json:"level"`
		Father   int    `json:"father"`
		Children []int  `json:"children,omitempty"`
	} `json:"categories_list"`
}

type WxaAddCategoryReq struct {
	Categories []WxaCategoryItem `json:"categories"`
}

type WxaModifyCategoryReq struct {
	First      int `json:"first"`
	Second     int `json:"second"`
	Certicates []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"certicates,omitempty"`
}

// ----- domain -----

type WxaModifyServerDomainReq struct {
	Action          string   `json:"action"` // add/delete/set/get/delete_legal_domain
	Requestdomain   []string `json:"requestdomain,omitempty"`
	Wsrequestdomain []string `json:"wsrequestdomain,omitempty"`
	Uploaddomain    []string `json:"uploaddomain,omitempty"`
	Downloaddomain  []string `json:"downloaddomain,omitempty"`
	Udpdomain       []string `json:"udpdomain,omitempty"`
	Tcpdomain       []string `json:"tcpdomain,omitempty"`
}

type WxaServerDomainResp struct {
	Requestdomain          []string `json:"requestdomain,omitempty"`
	Wsrequestdomain        []string `json:"wsrequestdomain,omitempty"`
	Uploaddomain           []string `json:"uploaddomain,omitempty"`
	Downloaddomain         []string `json:"downloaddomain,omitempty"`
	Udpdomain              []string `json:"udpdomain,omitempty"`
	Tcpdomain              []string `json:"tcpdomain,omitempty"`
	InvalidRequestdomain   []string `json:"invalid_requestdomain,omitempty"`
	InvalidWsrequestdomain []string `json:"invalid_wsrequestdomain,omitempty"`
	InvalidUploaddomain    []string `json:"invalid_uploaddomain,omitempty"`
	InvalidDownloaddomain  []string `json:"invalid_downloaddomain,omitempty"`
}

type WxaSetWebviewDomainReq struct {
	Action        string   `json:"action"`
	Webviewdomain []string `json:"webviewdomain,omitempty"`
}

type WxaDomainConfirmFile struct {
	FileName    string `json:"file_name"`
	FileContent string `json:"file_content"`
}

type WxaModifyDomainDirectlyReq struct {
	Action          string   `json:"action"`
	Requestdomain   []string `json:"requestdomain,omitempty"`
	Wsrequestdomain []string `json:"wsrequestdomain,omitempty"`
	Uploaddomain    []string `json:"uploaddomain,omitempty"`
	Downloaddomain  []string `json:"downloaddomain,omitempty"`
}

// ----- tester -----

type WxaBindTesterResp struct {
	UserStr string `json:"userstr"`
}

type WxaListTestersResp struct {
	Members []struct {
		UserStr string `json:"userstr"`
	} `json:"members"`
}
