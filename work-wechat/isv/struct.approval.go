package isv

// GetApprovalTemplateReq 获取审批模板详情请求。
type GetApprovalTemplateReq struct {
	TemplateID string `json:"template_id"`
}

// ApprovalTemplateResp 审批模板详情响应。
type ApprovalTemplateResp struct {
	TemplateNames   []ApprovalText  `json:"template_names"`
	TemplateContent ApprovalContent `json:"template_content"`
}

// ApprovalText 多语言文本。
type ApprovalText struct {
	Text string `json:"text"`
	Lang string `json:"lang"`
}

// ApprovalContent 审批模板内容。
type ApprovalContent struct {
	Controls []ApprovalControl `json:"controls"`
}

// ApprovalControl 审批模板控件。
type ApprovalControl struct {
	Property ApprovalControlProperty `json:"property"`
	Config   ApprovalControlConfig   `json:"config,omitempty"`
}

// ApprovalControlProperty 控件属性。
type ApprovalControlProperty struct {
	Control string         `json:"control"`
	ID      string         `json:"id"`
	Title   []ApprovalText `json:"title"`
}

// ApprovalControlConfig 控件配置。
type ApprovalControlConfig struct {
	Date     *ApprovalDateConfig     `json:"date,omitempty"`
	Selector *ApprovalSelectorConfig `json:"selector,omitempty"`
}

// ApprovalDateConfig 日期控件配置。
type ApprovalDateConfig struct {
	Type string `json:"type"`
}

// ApprovalSelectorConfig 选择控件配置。
type ApprovalSelectorConfig struct {
	Type    string           `json:"type"`
	Options []ApprovalOption `json:"options"`
}

// ApprovalOption 选择控件选项。
type ApprovalOption struct {
	Key   string         `json:"key"`
	Value []ApprovalText `json:"value"`
}

// ApplyEventReq 提交审批申请请求。
type ApplyEventReq struct {
	CreatorUserID       string         `json:"creator_userid"`
	TemplateID          string         `json:"template_id"`
	UseTemplateApprover int            `json:"use_template_approver"`
	ApplyData           ApplyData      `json:"apply_data"`
	SummaryList         []ApplySummary `json:"summary_list"`
}

// ApplyData 审批申请数据。
type ApplyData struct {
	Contents []ApplyContent `json:"contents"`
}

// ApplyContent 审批申请控件值。
type ApplyContent struct {
	Control string     `json:"control"`
	ID      string     `json:"id"`
	Value   ApplyValue `json:"value"`
}

// ApplyValue 控件值（各类型共用，按需填充）。
type ApplyValue struct {
	Text     string              `json:"text,omitempty"`
	Date     *ApplyDateValue     `json:"date,omitempty"`
	Selector *ApplySelectorValue `json:"selector,omitempty"`
}

// ApplyDateValue 日期控件值。
type ApplyDateValue struct {
	Type      string `json:"type"`
	Timestamp string `json:"s_timestamp"`
}

// ApplySelectorValue 选择控件值。
type ApplySelectorValue struct {
	Type    string             `json:"type"`
	Options []ApplySelectorOpt `json:"options"`
}

// ApplySelectorOpt 选择控件选中项。
type ApplySelectorOpt struct {
	Key string `json:"key"`
}

// ApplySummary 审批摘要。
type ApplySummary struct {
	SummaryInfo []ApprovalText `json:"summary_info"`
}

// ApplyEventResp 提交审批申请响应。
type ApplyEventResp struct {
	SpNo string `json:"sp_no"`
}

// GetApprovalDetailReq 获取审批申请详情请求。
type GetApprovalDetailReq struct {
	SpNo string `json:"sp_no"`
}

// ApprovalDetailResp 审批申请详情响应。
type ApprovalDetailResp struct {
	Info ApprovalInfoDetail `json:"info"`
}

// ApprovalInfoDetail 审批单详情。
type ApprovalInfoDetail struct {
	SpNo       string           `json:"sp_no"`
	SpName     string           `json:"sp_name"`
	SpStatus   int              `json:"sp_status"`
	TemplateID string           `json:"template_id"`
	ApplyTime  int64            `json:"apply_time"`
	Applyer    ApprovalApplyer  `json:"applyer"`
	SpRecord   []ApprovalRecord `json:"sp_record"`
	ApplyData  ApplyData        `json:"apply_data"`
}

// ApprovalApplyer 申请人信息。
type ApprovalApplyer struct {
	UserID  string `json:"userid"`
	PartyID string `json:"partyid"`
}

// ApprovalRecord 审批节点记录。
type ApprovalRecord struct {
	SpStatus     int              `json:"sp_status"`
	ApproverAttr int              `json:"approverattr"`
	Details      []ApprovalDetail `json:"details"`
}

// ApprovalDetail 审批人详情。
type ApprovalDetail struct {
	Approver ApprovalApprover `json:"approver"`
	Speech   string           `json:"speech"`
	SpStatus int              `json:"sp_status"`
	SpTime   int64            `json:"sptime"`
}

// ApprovalApprover 审批人。
type ApprovalApprover struct {
	UserID string `json:"userid"`
}

// GetApprovalDataReq 批量获取审批单号请求。
type GetApprovalDataReq struct {
	StartTime int64            `json:"starttime"`
	EndTime   int64            `json:"endtime"`
	Cursor    int              `json:"cursor,omitempty"`
	Size      int              `json:"size,omitempty"`
	Filters   []ApprovalFilter `json:"filters,omitempty"`
}

// ApprovalFilter 审批单号筛选条件。
type ApprovalFilter struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// GetApprovalDataResp 批量获取审批单号响应。
type GetApprovalDataResp struct {
	SpNoList      []string `json:"sp_no_list"`
	NewNextCursor int      `json:"new_next_cursor"`
}
