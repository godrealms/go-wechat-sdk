package offiaccount

import (
	"context"
	"fmt"
	"net/url"
)

// ocrRequest is the shared dispatch helper for all /cv/ocr/* endpoints.
//
// WeChat's OCR APIs accept EITHER an img_url query parameter OR a
// multipart/form-data POST carrying the image bytes under the "img" field —
// not both at once. Official curl example:
//
//	curl -F 'img=@test.jpg' \
//	  "https://api.weixin.qq.com/cv/ocr/idcard?access_token=...&type=photo"
//
// The filename is arbitrary (we pass "image"). The same contract applies to
// /cv/img/aicrop and /cv/img/qrcode.
//
// The previous implementation passed `img []byte` through doPost, which
// JSON-marshaled it and base64-encoded the file — WeChat always rejected
// those requests. Routing through doPostMultipartFile / doPostRaw fixes that.
func (c *Client) ocrRequest(ctx context.Context, endpoint, imgURL string, img []byte, result any) error {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("%s?access_token=%s", endpoint, token)
	if imgURL != "" {
		path += "&img_url=" + url.QueryEscape(imgURL)
	}

	if img != nil {
		return c.doPostMultipartFile(ctx, path, "img", "image", img, result)
	}
	// 仅 img_url 模式，空 body POST
	return c.doPostRaw(ctx, path, nil, "", result)
}

// MenuOcr 菜单识别
// imgURL: 要检测的图片 url，传这个则不用传 img 参数
// img: 图片文件，文件大小限制：小于2M
func (c *Client) MenuOcr(ctx context.Context, imgURL string, img []byte) (*MenuOcrResult, error) {
	var result MenuOcrResult
	if err := c.ocrRequest(ctx, "/cv/ocr/menu", imgURL, img, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CommOcr 通用印刷体识别
// imgURL: 要检测的图片 url，传这个则不用传 img 参数
// img: 图片文件（multipart/form-data "img" 字段），传这个则不用传 img_url
func (c *Client) CommOcr(ctx context.Context, imgURL string, img []byte) (*CommOcrResult, error) {
	var result CommOcrResult
	if err := c.ocrRequest(ctx, "/cv/ocr/comm", imgURL, img, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DrivingOcr 行驶证识别
// imgURL: 要检测的图片 url，传这个则不用传 img 参数
// img: 图片文件（multipart/form-data "img" 字段），传这个则不用传 img_url
func (c *Client) DrivingOcr(ctx context.Context, imgURL string, img []byte) (*DrivingOcrResult, error) {
	var result DrivingOcrResult
	if err := c.ocrRequest(ctx, "/cv/ocr/driving", imgURL, img, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// BankcardOcr 银行卡识别
// imgURL: 要检测的图片 url，传这个则不用传 img 参数
// img: 图片文件（multipart/form-data "img" 字段），传这个则不用传 img_url
func (c *Client) BankcardOcr(ctx context.Context, imgURL string, img []byte) (*BankcardOcrResult, error) {
	var result BankcardOcrResult
	if err := c.ocrRequest(ctx, "/cv/ocr/bankcard", imgURL, img, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// BizLicenseOcr 营业执照识别
// imgURL: 要检测的图片 url，传这个则不用传 img 参数
// img: 图片文件（multipart/form-data "img" 字段），传这个则不用传 img_url
func (c *Client) BizLicenseOcr(ctx context.Context, imgURL string, img []byte) (*BizLicenseOcrResult, error) {
	var result BizLicenseOcrResult
	if err := c.ocrRequest(ctx, "/cv/ocr/bizlicense", imgURL, img, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DrivingLicenseOcr 驾驶证识别
// imgURL: 要检测的图片 url，传这个则不用传 img 参数
// img: 图片文件（multipart/form-data "img" 字段），传这个则不用传 img_url
func (c *Client) DrivingLicenseOcr(ctx context.Context, imgURL string, img []byte) (*DrivingLicenseOcrResult, error) {
	var result DrivingLicenseOcrResult
	if err := c.ocrRequest(ctx, "/cv/ocr/drivinglicense", imgURL, img, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// IdCardOcr 身份证识别
// imgURL: 要检测的图片 url，传这个则不用传 img 参数
// img: 图片文件（multipart/form-data "img" 字段），传这个则不用传 img_url
func (c *Client) IdCardOcr(ctx context.Context, imgURL string, img []byte) (*IdCardOcrResult, error) {
	var result IdCardOcrResult
	if err := c.ocrRequest(ctx, "/cv/ocr/idcard", imgURL, img, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
