package isv

// ── Agent management ──────────────────────────────────────────────

// AgentDetail is the response from GetAgent.
type AgentDetail struct {
	AgentID            int          `json:"agentid"`
	Name               string       `json:"name"`
	Description        string       `json:"description"`
	SquareLogoURL      string       `json:"square_logo_url"`
	RoundLogoURL       string       `json:"round_logo_url"`
	HomeURL            string       `json:"home_url"`
	RedirectDomain     string       `json:"redirect_domain"`
	IsReportEnter      int          `json:"isreportenter"`
	ReportLocationFlag int          `json:"report_location_flag"`
	AllowUserInfos     AllowUsers   `json:"allow_userinfos"`
	AllowParties       AllowParties `json:"allow_partys"`
	AllowTags          AllowTags    `json:"allow_tags"`
}

// AllowUsers contains the list of users in the agent's visibility scope.
type AllowUsers struct {
	User []AllowUser `json:"user"`
}

// AllowUser represents a single user in the visibility scope.
type AllowUser struct {
	UserID string `json:"userid"`
}

// AllowParties contains the list of departments in the agent's visibility scope.
type AllowParties struct {
	PartyID []int `json:"partyid"`
}

// AllowTags contains the list of tags in the agent's visibility scope.
type AllowTags struct {
	TagID []int `json:"tagid"`
}

// SetAgentReq is the request body for SetAgent.
type SetAgentReq struct {
	AgentID            int    `json:"agentid"`
	Name               string `json:"name,omitempty"`
	Description        string `json:"description,omitempty"`
	LogoMediaID        string `json:"logo_mediaid,omitempty"`
	HomeURL            string `json:"home_url,omitempty"`
	RedirectDomain     string `json:"redirect_domain,omitempty"`
	IsReportEnter      *int   `json:"isreportenter,omitempty"`
	ReportLocationFlag *int   `json:"report_location_flag,omitempty"`
}

// ── Custom menu ───────────────────────────────────────────────────

// MenuButton represents a single menu button (supports nesting via SubButton).
type MenuButton struct {
	Type      string       `json:"type,omitempty"`
	Name      string       `json:"name"`
	Key       string       `json:"key,omitempty"`
	URL       string       `json:"url,omitempty"`
	AppID     string       `json:"appid,omitempty"`
	PagePath  string       `json:"pagepath,omitempty"`
	SubButton []MenuButton `json:"sub_button,omitempty"`
}

// CreateMenuReq is the request body for CreateMenu.
type CreateMenuReq struct {
	Button []MenuButton `json:"button"`
}

// MenuResp is the response from GetMenu.
type MenuResp struct {
	Button []MenuButton `json:"button"`
}

// ── Media upload ──────────────────────────────────────────────────

// UploadMediaResp is the response from UploadMedia.
type UploadMediaResp struct {
	Type      string `json:"type"`
	MediaID   string `json:"media_id"`
	CreatedAt string `json:"created_at"`
}
