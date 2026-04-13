package isv

// GetApprovalTemplateReq is the request for GetApprovalTemplate.
type GetApprovalTemplateReq struct {
	TemplateID string `json:"template_id"`
}

// ApprovalTemplateResp is the response containing approval template details.
type ApprovalTemplateResp struct {
	TemplateNames   []ApprovalText  `json:"template_names"`
	TemplateContent ApprovalContent `json:"template_content"`
}

// ApprovalText holds a piece of multilingual text.
type ApprovalText struct {
	Text string `json:"text"`
	Lang string `json:"lang"`
}

// ApprovalContent holds the control definitions of an approval template.
type ApprovalContent struct {
	Controls []ApprovalControl `json:"controls"`
}

// ApprovalControl represents a single control in an approval template.
type ApprovalControl struct {
	Property ApprovalControlProperty `json:"property"`
	Config   *ApprovalControlConfig  `json:"config,omitempty"`
}

// ApprovalControlProperty holds the basic properties of an approval control.
type ApprovalControlProperty struct {
	Control string         `json:"control"`
	ID      string         `json:"id"`
	Title   []ApprovalText `json:"title"`
}

// ApprovalControlConfig holds optional configuration for an approval control.
type ApprovalControlConfig struct {
	Date     *ApprovalDateConfig     `json:"date,omitempty"`
	Selector *ApprovalSelectorConfig `json:"selector,omitempty"`
}

// ApprovalDateConfig holds the configuration for a date-type approval control.
type ApprovalDateConfig struct {
	Type string `json:"type"`
}

// ApprovalSelectorConfig holds the configuration for a selector-type approval control.
type ApprovalSelectorConfig struct {
	Type    string           `json:"type"`
	Options []ApprovalOption `json:"options"`
}

// ApprovalOption represents a single option in a selector-type approval control.
type ApprovalOption struct {
	Key   string         `json:"key"`
	Value []ApprovalText `json:"value"`
}

// ApplyEventReq is the request for ApplyEvent.
type ApplyEventReq struct {
	CreatorUserID       string         `json:"creator_userid"`
	TemplateID          string         `json:"template_id"`
	UseTemplateApprover int            `json:"use_template_approver"`
	ApplyData           ApplyData      `json:"apply_data"`
	SummaryList         []ApplySummary `json:"summary_list"`
}

// ApplyData holds the control values submitted in an approval application.
type ApplyData struct {
	Contents []ApplyContent `json:"contents"`
}

// ApplyContent holds the value for a single control in an approval application.
type ApplyContent struct {
	Control string     `json:"control"`
	ID      string     `json:"id"`
	Value   ApplyValue `json:"value"`
}

// ApplyValue holds the control value; fields are populated according to the control type.
type ApplyValue struct {
	Text     string              `json:"text,omitempty"`
	Date     *ApplyDateValue     `json:"date,omitempty"`
	Selector *ApplySelectorValue `json:"selector,omitempty"`
}

// ApplyDateValue holds the value for a date-type approval control.
type ApplyDateValue struct {
	Type      string `json:"type"`
	Timestamp string `json:"s_timestamp"`
}

// ApplySelectorValue holds the value for a selector-type approval control.
type ApplySelectorValue struct {
	Type    string             `json:"type"`
	Options []ApplySelectorOpt `json:"options"`
}

// ApplySelectorOpt identifies a selected option in a selector-type approval control.
type ApplySelectorOpt struct {
	Key string `json:"key"`
}

// ApplySummary holds the summary information shown on an approval record.
type ApplySummary struct {
	SummaryInfo []ApprovalText `json:"summary_info"`
}

// ApplyEventResp is the response from ApplyEvent.
type ApplyEventResp struct {
	SpNo string `json:"sp_no"`
}

// GetApprovalDetailReq is the request for GetApprovalDetail.
type GetApprovalDetailReq struct {
	SpNo string `json:"sp_no"`
}

// ApprovalDetailResp is the response containing the full details of an approval record.
type ApprovalDetailResp struct {
	Info ApprovalInfoDetail `json:"info"`
}

// ApprovalInfoDetail contains the full detail of an approval record.
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

// ApprovalApplyer holds the identity of the applicant.
type ApprovalApplyer struct {
	UserID  string `json:"userid"`
	PartyID string `json:"partyid"`
}

// ApprovalRecord represents a single approval node record.
type ApprovalRecord struct {
	SpStatus     int              `json:"sp_status"`
	ApproverAttr int              `json:"approverattr"`
	Details      []ApprovalDetail `json:"details"`
}

// ApprovalDetail holds the details of an individual approver action at a node.
type ApprovalDetail struct {
	Approver ApprovalApprover `json:"approver"`
	Speech   string           `json:"speech"`
	SpStatus int              `json:"sp_status"`
	SpTime   int64            `json:"sptime"`
}

// ApprovalApprover identifies the approver at an approval node.
type ApprovalApprover struct {
	UserID string `json:"userid"`
}

// GetApprovalDataReq is the request for GetApprovalData.
type GetApprovalDataReq struct {
	StartTime int64            `json:"starttime"`
	EndTime   int64            `json:"endtime"`
	Cursor    int              `json:"cursor,omitempty"`
	Size      int              `json:"size,omitempty"`
	Filters   []ApprovalFilter `json:"filters,omitempty"`
}

// ApprovalFilter specifies a filter condition for querying approval records.
type ApprovalFilter struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// GetApprovalDataResp is the response from GetApprovalData.
type GetApprovalDataResp struct {
	SpNoList      []string `json:"sp_no_list"`
	NewNextCursor int      `json:"new_next_cursor"`
}
