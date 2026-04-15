package mini_program

import (
	"context"
	"fmt"
)

// Color specifies the RGB line color for a Mini Program QR code.
type Color struct {
	R int `json:"r"`
	G int `json:"g"`
	B int `json:"b"`
}

// GetWxaCodeReq is the request body for GetWxaCode (limited count, suitable for promotion pages).
type GetWxaCodeReq struct {
	Path      string `json:"path"`
	Width     int    `json:"width,omitempty"`
	AutoColor bool   `json:"auto_color,omitempty"`
	LineColor *Color `json:"line_color,omitempty"`
	IsHyaline bool   `json:"is_hyaline,omitempty"`
}

// GetWxaCodeUnlimitReq is the request body for GetWxaCodeUnlimit (unlimited count, distinguished by scene parameter).
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

// CreateQRCodeReq is the request body for CreateQRCode (limited count QR code).
type CreateQRCodeReq struct {
	Path  string `json:"path"`
	Width int    `json:"width,omitempty"`
}

// GetWxaCode returns a Mini Program QR code image (limited count, suitable for fixed-page promotion) as PNG bytes.
func (c *Client) GetWxaCode(ctx context.Context, req *GetWxaCodeReq) ([]byte, error) {
	if req == nil {
		return nil, fmt.Errorf("mini_program: GetWxaCode: req is required")
	}
	if req.Path == "" {
		return nil, fmt.Errorf("mini_program: GetWxaCode: req.Path is required")
	}
	return c.doPostRaw(ctx, "/wxa/getwxacode", req)
}

// GetWxaCodeUnlimit returns an unlimited Mini Program QR code image (distinguished by the scene parameter) as PNG bytes.
func (c *Client) GetWxaCodeUnlimit(ctx context.Context, req *GetWxaCodeUnlimitReq) ([]byte, error) {
	if req == nil {
		return nil, fmt.Errorf("mini_program: GetWxaCodeUnlimit: req is required")
	}
	if req.Scene == "" {
		return nil, fmt.Errorf("mini_program: GetWxaCodeUnlimit: req.Scene is required")
	}
	return c.doPostRaw(ctx, "/wxa/getwxacodeunlimit", req)
}

// CreateQRCode returns a limited-count Mini Program QR code image as PNG bytes.
func (c *Client) CreateQRCode(ctx context.Context, req *CreateQRCodeReq) ([]byte, error) {
	if req == nil {
		return nil, fmt.Errorf("mini_program: CreateQRCode: req is required")
	}
	if req.Path == "" {
		return nil, fmt.Errorf("mini_program: CreateQRCode: req.Path is required")
	}
	return c.doPostRaw(ctx, "/cgi-bin/wxaapp/createwxaqrcode", req)
}
