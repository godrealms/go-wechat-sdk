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

// ----- code -----

type WxaCommitReq struct {
	TemplateID  int    `json:"template_id"`
	UserVersion string `json:"user_version"`
	UserDesc    string `json:"user_desc"`
	ExtJSON     string `json:"ext_json"`
}

type WxaGetPageResp struct {
	PageList []string `json:"page_list"`
}

type WxaGetCodeCategoryResp struct {
	CategoryList []struct {
		FirstClass  string `json:"first_class"`
		SecondClass string `json:"second_class"`
		ThirdClass  string `json:"third_class,omitempty"`
		FirstID     int    `json:"first_id"`
		SecondID    int    `json:"second_id"`
		ThirdID     int    `json:"third_id,omitempty"`
	} `json:"category_list"`
}

// ----- release -----

type WxaSubmitAuditReq struct {
	ItemList []struct {
		Address     string `json:"address"`
		Tag         string `json:"tag"`
		FirstClass  string `json:"first_class"`
		SecondClass string `json:"second_class"`
		ThirdClass  string `json:"third_class,omitempty"`
		FirstID     int    `json:"first_id"`
		SecondID    int    `json:"second_id"`
		ThirdID     int    `json:"third_id,omitempty"`
		Title       string `json:"title,omitempty"`
	} `json:"item_list,omitempty"`
	PreviewInfo *struct {
		VideoIDList []string `json:"video_id_list,omitempty"`
		PicIDList   []string `json:"pic_id_list,omitempty"`
	} `json:"preview_info,omitempty"`
	VersionDesc   string `json:"version_desc,omitempty"`
	FeedbackInfo  string `json:"feedback_info,omitempty"`
	FeedbackStuff string `json:"feedback_stuff,omitempty"`
}

type WxaSubmitAuditResp struct {
	AuditID int64 `json:"auditid"`
}

type WxaAuditStatus struct {
	AuditID         int64  `json:"auditid,omitempty"`
	Status          int    `json:"status"`
	Reason          string `json:"reason,omitempty"`
	ScreenShot      string `json:"screenshot,omitempty"`
	UserVersion     string `json:"user_version,omitempty"`
	UserDesc        string `json:"user_desc,omitempty"`
	SubmitAuditTime int64  `json:"submit_audit_time,omitempty"`
}

type WxaSupportVersionResp struct {
	NowVersion string `json:"now_version"`
	UVInfo     struct {
		Items []struct {
			Percentage float64 `json:"percentage"`
			Version    string  `json:"version"`
		} `json:"items"`
	} `json:"uv_info"`
}
