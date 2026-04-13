package mini_program

import "context"

// JumpWxa specifies the Mini Program path and query to open via a URL Scheme.
type JumpWxa struct {
	Path       string `json:"path"`
	Query      string `json:"query"`
	EnvVersion string `json:"env_version,omitempty"`
}

// GenerateSchemeReq is the request body for generating a URL Scheme.
type GenerateSchemeReq struct {
	JumpWxa        *JumpWxa `json:"jump_wxa,omitempty"`
	IsExpire       bool     `json:"is_expire,omitempty"`
	ExpireType     int      `json:"expire_type,omitempty"`
	ExpireTime     int64    `json:"expire_time,omitempty"`
	ExpireInterval int      `json:"expire_interval,omitempty"`
}

// GenerateSchemeResp is the response from the URL Scheme generation API.
type GenerateSchemeResp struct {
	OpenLink string `json:"openlink"`
}

// GenerateUrlLinkReq is the request body for generating a URL Link.
type GenerateUrlLinkReq struct {
	Path           string `json:"path,omitempty"`
	Query          string `json:"query,omitempty"`
	EnvVersion     string `json:"env_version,omitempty"`
	IsExpire       bool   `json:"is_expire,omitempty"`
	ExpireType     int    `json:"expire_type,omitempty"`
	ExpireTime     int64  `json:"expire_time,omitempty"`
	ExpireInterval int    `json:"expire_interval,omitempty"`
}

// GenerateUrlLinkResp is the response from the URL Link generation API.
type GenerateUrlLinkResp struct {
	URLLink string `json:"url_link"`
}

// GenerateShortLinkReq is the request body for generating a Short Link.
type GenerateShortLinkReq struct {
	PageURL     string `json:"page_url"`
	PageTitle   string `json:"page_title,omitempty"`
	IsPermanent bool   `json:"is_permanent,omitempty"`
}

// GenerateShortLinkResp is the response from the Short Link generation API.
type GenerateShortLinkResp struct {
	Link string `json:"link"`
}

// GenerateScheme generates a URL Scheme for launching the Mini Program from SMS, email, or other external contexts.
func (c *Client) GenerateScheme(ctx context.Context, req *GenerateSchemeReq) (*GenerateSchemeResp, error) {
	var resp GenerateSchemeResp
	if err := c.doPost(ctx, "/wxa/generatescheme", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GenerateUrlLink generates a URL Link suitable for contexts where URL Schemes cannot be used (e.g. SMS, email).
func (c *Client) GenerateUrlLink(ctx context.Context, req *GenerateUrlLinkReq) (*GenerateUrlLinkResp, error) {
	var resp GenerateUrlLinkResp
	if err := c.doPost(ctx, "/wxa/generate_urllink", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GenerateShortLink generates a Short Link for sharing within WeChat.
func (c *Client) GenerateShortLink(ctx context.Context, req *GenerateShortLinkReq) (*GenerateShortLinkResp, error) {
	var resp GenerateShortLinkResp
	if err := c.doPost(ctx, "/wxa/genwxashortlink", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
