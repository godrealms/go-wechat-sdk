package offiaccount

import (
	"context"
	"fmt"
	"net/url"
)

// ImgAiCrop 图片智能裁剪
// imgURL: 要检测的图片 url，传这个则不用传 img 参数
// img: 图片文件，小于2M
// ratios: 宽高比；如果提供多个宽高比，请以英文逗号","分隔，最多支持5个宽高比
func (c *Client) ImgAiCrop(ctx context.Context, imgURL, ratios string, img []byte) (*ImgAiCropResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/cv/img/aicrop?access_token=%s", token)

	// 添加查询参数
	if imgURL != "" {
		path += "&img_url=" + url.QueryEscape(imgURL)
	}
	if ratios != "" {
		path += "&ratios=" + url.QueryEscape(ratios)
	}

	// 发送请求
	var result ImgAiCropResult
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

// ImgQrcode 二维码/条码识别
// imgURL: 图片URL地址
// img: 图片文件，限制小于 2 M
func (c *Client) ImgQrcode(ctx context.Context, imgURL string, img []byte) (*ImgQrcodeResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/cv/img/qrcode?access_token=%s", token)
	if imgURL != "" {
		path = fmt.Sprintf("%s&img_url=%s", path, url.QueryEscape(imgURL))
	}

	// 发送请求
	var result ImgQrcodeResult
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
