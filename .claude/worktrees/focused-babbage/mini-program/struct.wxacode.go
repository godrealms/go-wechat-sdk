package mini_program

import "github.com/godrealms/go-wechat-sdk/core"

// QRCodeRequest is the request for GetQRCode (limited to 100,000 codes)
type QRCodeRequest struct {
	Path      string         `json:"path"`
	Width     int            `json:"width,omitempty"`
	AutoColor bool           `json:"auto_color,omitempty"`
	LineColor map[string]int `json:"line_color,omitempty"`
	IsHyaline bool           `json:"is_hyaline,omitempty"`
}

// UnlimitedQRCodeRequest is the request for GetUnlimited (no count limit)
type UnlimitedQRCodeRequest struct {
	Scene       string         `json:"scene"`
	Page        string         `json:"page,omitempty"`
	CheckPath   bool           `json:"check_path,omitempty"`
	EnvVersion  string         `json:"env_version,omitempty"`
	Width       int            `json:"width,omitempty"`
	AutoColor   bool           `json:"auto_color,omitempty"`
	LineColor   map[string]int `json:"line_color,omitempty"`
	IsHyaline   bool           `json:"is_hyaline,omitempty"`
}

// CreateQRCodeRequest is the request for CreateQRCode (QR code, not mini-program code)
type CreateQRCodeRequest struct {
	Path  string `json:"path"`
	Width int    `json:"width,omitempty"`
}

// JumpWxa describes the mini-program page for scheme/urllink
type JumpWxa struct {
	Path       string `json:"path"`
	Query      string `json:"query"`
	EnvVersion string `json:"env_version,omitempty"`
}

// CloudBase describes cloud development config for urllink
type CloudBase struct {
	Env           string `json:"env"`
	Domain        string `json:"domain,omitempty"`
	Path          string `json:"path,omitempty"`
	Query         string `json:"query,omitempty"`
	ResourceAppid string `json:"resource_appid,omitempty"`
}

// GenerateSchemeRequest is the request for GenerateScheme
type GenerateSchemeRequest struct {
	JumpWxa        *JumpWxa `json:"jump_wxa,omitempty"`
	IsExpire       bool     `json:"is_expire,omitempty"`
	ExpireType     int      `json:"expire_type,omitempty"`
	ExpireTime     int64    `json:"expire_time,omitempty"`
	ExpireInterval int      `json:"expire_interval,omitempty"`
	EnvVersion     string   `json:"env_version,omitempty"`
}

// GenerateSchemeResult is the result of GenerateScheme
type GenerateSchemeResult struct {
	core.Resp
	OpenLink string `json:"openlink"`
}

// SchemeInfo contains scheme metadata
type SchemeInfo struct {
	AppId      string `json:"appid"`
	Path       string `json:"path"`
	Query      string `json:"query"`
	CreateTime int64  `json:"create_time"`
	ExpireTime int64  `json:"expire_time"`
	EnvVersion string `json:"env_version"`
	OpenLink   string `json:"openlink"`
}

// SchemeQuota contains scheme quota information
type SchemeQuota struct {
	LongTimeUsedQuota    int64 `json:"long_time_used_quota"`
	LongTimeDefaultQuota int64 `json:"long_time_default_quota"`
}

// QuerySchemeResult is the result of QueryScheme
type QuerySchemeResult struct {
	core.Resp
	SchemeInfo  *SchemeInfo  `json:"scheme_info"`
	SchemeQuota *SchemeQuota `json:"scheme_quota"`
}

// GenerateUrlLinkRequest is the request for GenerateUrlLink
type GenerateUrlLinkRequest struct {
	Path           string     `json:"path,omitempty"`
	Query          string     `json:"query,omitempty"`
	IsExpire       bool       `json:"is_expire,omitempty"`
	ExpireType     int        `json:"expire_type,omitempty"`
	ExpireTime     int64      `json:"expire_time,omitempty"`
	ExpireInterval int        `json:"expire_interval,omitempty"`
	CloudBase      *CloudBase `json:"cloud_base,omitempty"`
	EnvVersion     string     `json:"env_version,omitempty"`
}

// GenerateUrlLinkResult is the result of GenerateUrlLink
type GenerateUrlLinkResult struct {
	core.Resp
	UrlLink string `json:"url_link"`
}

// UrlLinkInfo contains URL Link metadata
type UrlLinkInfo struct {
	AppId      string     `json:"appid"`
	Path       string     `json:"path"`
	Query      string     `json:"query"`
	CreateTime int64      `json:"create_time"`
	ExpireTime int64      `json:"expire_time"`
	EnvVersion string     `json:"env_version"`
	UrlLink    string     `json:"url_link"`
	CloudBase  *CloudBase `json:"cloud_base"`
}

// UrlLinkQuota contains URL Link quota information
type UrlLinkQuota struct {
	LongTimeUsedQuota    int64 `json:"long_time_used_quota"`
	LongTimeDefaultQuota int64 `json:"long_time_default_quota"`
}

// QueryUrlLinkResult is the result of QueryUrlLink
type QueryUrlLinkResult struct {
	core.Resp
	UrlLinkInfo  *UrlLinkInfo  `json:"url_link_info"`
	UrlLinkQuota *UrlLinkQuota `json:"url_link_quota"`
}

// GenerateShortLinkRequest is the request for GenerateShortLink
type GenerateShortLinkRequest struct {
	PageUrl   string `json:"page_url"`
	PageTitle string `json:"page_title,omitempty"`
	Permanent bool   `json:"permanent,omitempty"`
}

// GenerateShortLinkResult is the result of GenerateShortLink
type GenerateShortLinkResult struct {
	core.Resp
	Link string `json:"link"`
}
