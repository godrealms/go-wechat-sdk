package isv

// ---- workbench ----

// WBKeyDataItem 关键数据条目。
type WBKeyDataItem struct {
	Key      string `json:"key"`
	Data     string `json:"data"`
	JumpURL  string `json:"jump_url,omitempty"`
	PagePath string `json:"pagepath,omitempty"`
}

// WBKeyData 关键数据型工作台。
type WBKeyData struct {
	Items []WBKeyDataItem `json:"items"`
}

// WBImage 图片型工作台。
type WBImage struct {
	URL      string `json:"url"`
	JumpURL  string `json:"jump_url,omitempty"`
	PagePath string `json:"pagepath,omitempty"`
}

// WBListItem 列表条目。
type WBListItem struct {
	Title    string `json:"title"`
	JumpURL  string `json:"jump_url,omitempty"`
	PagePath string `json:"pagepath,omitempty"`
}

// WBList 列表型工作台。
type WBList struct {
	Items []WBListItem `json:"items"`
}

// WBWebview 网页型工作台。
type WBWebview struct {
	URL      string `json:"url"`
	JumpURL  string `json:"jump_url,omitempty"`
	PagePath string `json:"pagepath,omitempty"`
}

// WorkbenchTemplateReq 是 agent/set_workbench_template 的请求体。
type WorkbenchTemplateReq struct {
	AgentID         int        `json:"agentid"`
	Type            string     `json:"type"` // key_data / image / list / webview / normal
	KeyData         *WBKeyData `json:"key_data,omitempty"`
	Image           *WBImage   `json:"image,omitempty"`
	List            *WBList    `json:"list,omitempty"`
	Webview         *WBWebview `json:"webview,omitempty"`
	ReplaceUserData bool       `json:"replace_user_data,omitempty"`
}

// WorkbenchTemplateResp 是 agent/get_workbench_template 的响应。
type WorkbenchTemplateResp struct {
	AgentID         int        `json:"agentid"`
	Type            string     `json:"type"`
	KeyData         *WBKeyData `json:"key_data,omitempty"`
	Image           *WBImage   `json:"image,omitempty"`
	List            *WBList    `json:"list,omitempty"`
	Webview         *WBWebview `json:"webview,omitempty"`
	ReplaceUserData bool       `json:"replace_user_data"`
}

// WorkbenchDataReq 是 agent/set_workbench_data 的请求体。
type WorkbenchDataReq struct {
	AgentID int        `json:"agentid"`
	UserID  string     `json:"userid"`
	Type    string     `json:"type"`
	KeyData *WBKeyData `json:"key_data,omitempty"`
	Image   *WBImage   `json:"image,omitempty"`
	List    *WBList    `json:"list,omitempty"`
	Webview *WBWebview `json:"webview,omitempty"`
}
