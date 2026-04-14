package offiaccount

import (
	"context"
	"fmt"
	"net/url"
)

// MenuOcr 菜单识别
// imgURL: 要检测的图片 url，传这个则不用传 img 参数
// img: 图片文件，文件大小限制：小于2M
func (c *Client) MenuOcr(ctx context.Context, imgURL string, img []byte) (*MenuOcrResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/cv/ocr/menu?access_token=%s", token)
	if imgURL != "" {
		path = fmt.Sprintf("%s&img_url=%s", path, url.QueryEscape(imgURL))
	}

	// 发送请求
	var result MenuOcrResult
	if img != nil {
		err = c.Https.Post(ctx, path, img, &result)
	} else {
		err = c.Https.Post(ctx, path, nil, &result)
	}
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// CommOcr 通用印刷体识别
// imgURL: 要检测的图片 url，传这个则不用传 img 参数
// img: form-data 中媒体文件标识，有filename、filelength、content-type等信息，传这个则不用传 img_url
func (c *Client) CommOcr(ctx context.Context, imgURL string, img []byte) (*CommOcrResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/cv/ocr/comm?access_token=%s", token)
	if imgURL != "" {
		path = fmt.Sprintf("%s&img_url=%s", path, url.QueryEscape(imgURL))
	}

	// 发送请求
	var result CommOcrResult
	if img != nil {
		err = c.Https.Post(ctx, path, img, &result)
	} else {
		err = c.Https.Post(ctx, path, nil, &result)
	}
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// DrivingOcr 行驶证识别
// imgURL: 要检测的图片 url，传这个则不用传 img 参数
// img: form-data 中媒体文件标识，有filename、filelength、content-type等信息，传这个则不用传 img_url
func (c *Client) DrivingOcr(ctx context.Context, imgURL string, img []byte) (*DrivingOcrResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/cv/ocr/driving?access_token=%s", token)
	if imgURL != "" {
		path = fmt.Sprintf("%s&img_url=%s", path, url.QueryEscape(imgURL))
	}

	// 发送请求
	var result DrivingOcrResult
	if img != nil {
		err = c.Https.Post(ctx, path, img, &result)
	} else {
		err = c.Https.Post(ctx, path, nil, &result)
	}
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// BankcardOcr 银行卡识别
// imgURL: 要检测的图片 url，传这个则不用传 img 参数
// img: form-data 中媒体文件标识，有filename、filelength、content-type等信息，传这个则不用传 img_url
func (c *Client) BankcardOcr(ctx context.Context, imgURL string, img []byte) (*BankcardOcrResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/cv/ocr/bankcard?access_token=%s", token)
	if imgURL != "" {
		path = fmt.Sprintf("%s&img_url=%s", path, url.QueryEscape(imgURL))
	}

	// 发送请求
	var result BankcardOcrResult
	if img != nil {
		err = c.Https.Post(ctx, path, img, &result)
	} else {
		err = c.Https.Post(ctx, path, nil, &result)
	}
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// BizLicenseOcr 营业执照识别
// imgURL: 要检测的图片 url，传这个则不用传 img 参数
// img: form-data 中媒体文件标识，有filename、filelength、content-type等信息，传这个则不用传 img_url
func (c *Client) BizLicenseOcr(ctx context.Context, imgURL string, img []byte) (*BizLicenseOcrResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/cv/ocr/bizlicense?access_token=%s", token)
	if imgURL != "" {
		path = fmt.Sprintf("%s&img_url=%s", path, url.QueryEscape(imgURL))
	}

	// 发送请求
	var result BizLicenseOcrResult
	if img != nil {
		err = c.Https.Post(ctx, path, img, &result)
	} else {
		err = c.Https.Post(ctx, path, nil, &result)
	}
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// DrivingLicenseOcr 驾驶证识别
// imgURL: 要检测的图片 url，传这个则不用传 img 参数
// img: form-data 中媒体文件标识，有filename、filelength、content-type等信息，传这个则不用传 img_url
func (c *Client) DrivingLicenseOcr(ctx context.Context, imgURL string, img []byte) (*DrivingLicenseOcrResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/cv/ocr/drivinglicense?access_token=%s", token)
	if imgURL != "" {
		path = fmt.Sprintf("%s&img_url=%s", path, url.QueryEscape(imgURL))
	}

	// 发送请求
	var result DrivingLicenseOcrResult
	if img != nil {
		err = c.Https.Post(ctx, path, img, &result)
	} else {
		err = c.Https.Post(ctx, path, nil, &result)
	}
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// IdCardOcr 身份证识别
// imgURL: 要检测的图片 url，传这个则不用传 img 参数
// img: form-data 中媒体文件标识，有filename、filelength、content-type等信息，传这个则不用传 img_url
func (c *Client) IdCardOcr(ctx context.Context, imgURL string, img []byte) (*IdCardOcrResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/cv/ocr/idcard?access_token=%s", token)
	if imgURL != "" {
		path = fmt.Sprintf("%s&img_url=%s", path, url.QueryEscape(imgURL))
	}

	// 发送请求
	var result IdCardOcrResult
	if img != nil {
		err = c.Https.Post(ctx, path, img, &result)
	} else {
		err = c.Https.Post(ctx, path, nil, &result)
	}
	if err != nil {
		return nil, err
	}

	return &result, nil
}
