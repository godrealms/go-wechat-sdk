package mini_program

import "context"

// JumpWxa 跳转小程序参数。
type JumpWxa struct {
	Path       string `json:"path"`
	Query      string `json:"query"`
	EnvVersion string `json:"env_version,omitempty"`
}

// GenerateSchemeReq 生成 URL Scheme 请求。
type GenerateSchemeReq struct {
	JumpWxa        *JumpWxa `json:"jump_wxa,omitempty"`
	IsExpire       bool     `json:"is_expire,omitempty"`
	ExpireType     int      `json:"expire_type,omitempty"`
	ExpireTime     int64    `json:"expire_time,omitempty"`
	ExpireInterval int      `json:"expire_interval,omitempty"`
}

// GenerateSchemeResp 生成 URL Scheme 响应。
type GenerateSchemeResp struct {
	OpenLink string `json:"openlink"`
}

// GenerateUrlLinkReq 生成 URL Link 请求。
type GenerateUrlLinkReq struct {
	Path           string `json:"path,omitempty"`
	Query          string `json:"query,omitempty"`
	EnvVersion     string `json:"env_version,omitempty"`
	IsExpire       bool   `json:"is_expire,omitempty"`
	ExpireType     int    `json:"expire_type,omitempty"`
	ExpireTime     int64  `json:"expire_time,omitempty"`
	ExpireInterval int    `json:"expire_interval,omitempty"`
}

// GenerateUrlLinkResp 生成 URL Link 响应。
type GenerateUrlLinkResp struct {
	URLLink string `json:"url_link"`
}

// GenerateShortLinkReq 生成 Short Link 请求。
type GenerateShortLinkReq struct {
	PageURL     string `json:"page_url"`
	PageTitle   string `json:"page_title,omitempty"`
	IsPermanent bool   `json:"is_permanent,omitempty"`
}

// GenerateShortLinkResp 生成 Short Link 响应。
type GenerateShortLinkResp struct {
	Link string `json:"link"`
}

// GenerateScheme 生成 URL Scheme（用于短信、邮件等外部拉起小程序）。
func (c *Client) GenerateScheme(ctx context.Context, req *GenerateSchemeReq) (*GenerateSchemeResp, error) {
	var resp GenerateSchemeResp
	if err := c.doPost(ctx, "/wxa/generatescheme", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GenerateUrlLink 生成 URL Link（适用于短信、邮件等无法使用 Scheme 的场景）。
func (c *Client) GenerateUrlLink(ctx context.Context, req *GenerateUrlLinkReq) (*GenerateUrlLinkResp, error) {
	var resp GenerateUrlLinkResp
	if err := c.doPost(ctx, "/wxa/generate_urllink", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GenerateShortLink 生成 Short Link（短链，适用于微信内传播）。
func (c *Client) GenerateShortLink(ctx context.Context, req *GenerateShortLinkReq) (*GenerateShortLinkResp, error) {
	var resp GenerateShortLinkResp
	if err := c.doPost(ctx, "/wxa/genwxashortlink", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
