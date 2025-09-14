package offiaccount

import (
	"fmt"
	"net/url"
)

// MenuOcr 菜单识别
// imgURL: 要检测的图片 url，传这个则不用传 img 参数
// img: 图片文件，文件大小限制：小于2M
func (c *Client) MenuOcr(imgURL string, img []byte) (*MenuOcrResult, error) {
	// 构造请求URL
	path := "/cv/ocr/menu"
	if imgURL != "" {
		path = fmt.Sprintf("%s?img_url=%s", path, url.QueryEscape(imgURL))
	}

	// 发送请求
	var result MenuOcrResult
	var err error
	if img != nil {
		err = c.Https.Post(c.ctx, path, img, &result)
	} else {
		err = c.Https.Post(c.ctx, path, nil, &result)
	}
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// CommOcr 通用印刷体识别
// imgURL: 要检测的图片 url，传这个则不用传 img 参数
// img: form-data 中媒体文件标识，有filename、filelength、content-type等信息，传这个则不用传 img_url
func (c *Client) CommOcr(imgURL string, img []byte) (*CommOcrResult, error) {
	// 构造请求URL
	path := "/cv/ocr/comm"
	if imgURL != "" {
		path = fmt.Sprintf("%s?img_url=%s", path, url.QueryEscape(imgURL))
	}

	// 发送请求
	var result CommOcrResult
	var err error
	if img != nil {
		err = c.Https.Post(c.ctx, path, img, &result)
	} else {
		err = c.Https.Post(c.ctx, path, nil, &result)
	}
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// DrivingOcr 行驶证识别
// imgURL: 要检测的图片 url，传这个则不用传 img 参数
// img: form-data 中媒体文件标识，有filename、filelength、content-type等信息，传这个则不用传 img_url
func (c *Client) DrivingOcr(imgURL string, img []byte) (*DrivingOcrResult, error) {
	// 构造请求URL
	path := "/cv/ocr/driving"
	if imgURL != "" {
		path = fmt.Sprintf("%s?img_url=%s", path, url.QueryEscape(imgURL))
	}

	// 发送请求
	var result DrivingOcrResult
	var err error
	if img != nil {
		err = c.Https.Post(c.ctx, path, img, &result)
	} else {
		err = c.Https.Post(c.ctx, path, nil, &result)
	}
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// BankcardOcr 银行卡识别
// imgURL: 要检测的图片 url，传这个则不用传 img 参数
// img: form-data 中媒体文件标识，有filename、filelength、content-type等信息，传这个则不用传 img_url
func (c *Client) BankcardOcr(imgURL string, img []byte) (*BankcardOcrResult, error) {
	// 构造请求URL
	path := "/cv/ocr/bankcard"
	if imgURL != "" {
		path = fmt.Sprintf("%s?img_url=%s", path, url.QueryEscape(imgURL))
	}

	// 发送请求
	var result BankcardOcrResult
	var err error
	if img != nil {
		err = c.Https.Post(c.ctx, path, img, &result)
	} else {
		err = c.Https.Post(c.ctx, path, nil, &result)
	}
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// BizLicenseOcr 营业执照识别
// imgURL: 要检测的图片 url，传这个则不用传 img 参数
// img: form-data 中媒体文件标识，有filename、filelength、content-type等信息，传这个则不用传 img_url
func (c *Client) BizLicenseOcr(imgURL string, img []byte) (*BizLicenseOcrResult, error) {
	// 构造请求URL
	path := "/cv/ocr/bizlicense"
	if imgURL != "" {
		path = fmt.Sprintf("%s?img_url=%s", path, url.QueryEscape(imgURL))
	}

	// 发送请求
	var result BizLicenseOcrResult
	var err error
	if img != nil {
		err = c.Https.Post(c.ctx, path, img, &result)
	} else {
		err = c.Https.Post(c.ctx, path, nil, &result)
	}
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// DrivingLicenseOcr 驾驶证识别
// imgURL: 要检测的图片 url，传这个则不用传 img 参数
// img: form-data 中媒体文件标识，有filename、filelength、content-type等信息，传这个则不用传 img_url
func (c *Client) DrivingLicenseOcr(imgURL string, img []byte) (*DrivingLicenseOcrResult, error) {
	// 构造请求URL
	path := "/cv/ocr/drivinglicense"
	if imgURL != "" {
		path = fmt.Sprintf("%s?img_url=%s", path, url.QueryEscape(imgURL))
	}

	// 发送请求
	var result DrivingLicenseOcrResult
	var err error
	if img != nil {
		err = c.Https.Post(c.ctx, path, img, &result)
	} else {
		err = c.Https.Post(c.ctx, path, nil, &result)
	}
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// IdCardOcr 身份证识别
// imgURL: 要检测的图片 url，传这个则不用传 img 参数
// img: form-data 中媒体文件标识，有filename、filelength、content-type等信息，传这个则不用传 img_url
func (c *Client) IdCardOcr(imgURL string, img []byte) (*IdCardOcrResult, error) {
	// 构造请求URL
	path := "/cv/ocr/idcard"
	if imgURL != "" {
		path = fmt.Sprintf("%s?img_url=%s", path, url.QueryEscape(imgURL))
	}

	// 发送请求
	var result IdCardOcrResult
	var err error
	if img != nil {
		err = c.Https.Post(c.ctx, path, img, &result)
	} else {
		err = c.Https.Post(c.ctx, path, nil, &result)
	}
	if err != nil {
		return nil, err
	}

	return &result, nil
}
