package mini_program

import "context"

// Color 小程序码线条颜色。
type Color struct {
	R int `json:"r"`
	G int `json:"g"`
	B int `json:"b"`
}

// GetWxaCodeReq 获取小程序码请求（有限个，适合推广）。
type GetWxaCodeReq struct {
	Path      string `json:"path"`
	Width     int    `json:"width,omitempty"`
	AutoColor bool   `json:"auto_color,omitempty"`
	LineColor *Color `json:"line_color,omitempty"`
	IsHyaline bool   `json:"is_hyaline,omitempty"`
}

// GetWxaCodeUnlimitReq 获取小程序码请求（无限个，scene 参数）。
type GetWxaCodeUnlimitReq struct {
	Scene      string `json:"scene"`
	Page       string `json:"page,omitempty"`
	Width      int    `json:"width,omitempty"`
	AutoColor  bool   `json:"auto_color,omitempty"`
	LineColor  *Color `json:"line_color,omitempty"`
	IsHyaline  bool   `json:"is_hyaline,omitempty"`
	CheckPath  bool   `json:"check_path,omitempty"`
	EnvVersion string `json:"env_version,omitempty"`
}

// CreateQRCodeReq 获取小程序二维码请求。
type CreateQRCodeReq struct {
	Path  string `json:"path"`
	Width int    `json:"width,omitempty"`
}

// GetWxaCode 获取小程序码（有限个，适合固定页面推广）。返回 PNG 图片字节。
func (c *Client) GetWxaCode(ctx context.Context, req *GetWxaCodeReq) ([]byte, error) {
	return c.doPostRaw(ctx, "/wxa/getwxacode", req)
}

// GetWxaCodeUnlimit 获取小程序码（无限个，通过 scene 参数区分）。返回 PNG 图片字节。
func (c *Client) GetWxaCodeUnlimit(ctx context.Context, req *GetWxaCodeUnlimitReq) ([]byte, error) {
	return c.doPostRaw(ctx, "/wxa/getwxacodeunlimit", req)
}

// CreateQRCode 获取小程序二维码（有限个）。返回 PNG 图片字节。
func (c *Client) CreateQRCode(ctx context.Context, req *CreateQRCodeReq) ([]byte, error) {
	return c.doPostRaw(ctx, "/cgi-bin/wxaapp/createwxaqrcode", req)
}
