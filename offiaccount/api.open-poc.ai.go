package offiaccount

import (
	"context"
	"fmt"
	"net/url"
)

// TranslateContent 微信翻译
// lfrom: 源语言，zh_CN 或 en_US
// lto: 目标语言，zh_CN 或 en_US
// content: 源内容(utf8格式，最大600Byte)
func (c *Client) TranslateContent(ctx context.Context, lfrom, lto, content string) (*TranslateContentResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/media/voice/translatecontent?access_token=%s&lfrom=%s&lto=%s",
		token, url.QueryEscape(lfrom), url.QueryEscape(lto))

	// 发送请求
	var result TranslateContentResult
	if err := c.Https.Post(ctx, path, []byte(content), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// AddVoiceToRecoForText 上传语音文件进行转文字识别
// format: 文件格式（只支持mp3，16k，单声道，最大1M）
// voiceID: 语音唯一标识
// lang: 语言，zh_CN 或 en_US，默认中文
// media: 语音内容
func (c *Client) AddVoiceToRecoForText(ctx context.Context, format, voiceID, lang string, media []byte) (*Resp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/media/voice/addvoicetorecofortext?access_token=%s&format=%s&voice_id=%s&lang=%s",
		token, url.QueryEscape(format), url.QueryEscape(voiceID), url.QueryEscape(lang))

	// 发送请求
	var result Resp
	if err := c.Https.Post(ctx, path, media, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// QueryRecoResultForText 查询语音转文字结果
// voiceID: 语音唯一标识
// lang: 语言，zh_CN 或 en_US，默认中文
func (c *Client) QueryRecoResultForText(ctx context.Context, voiceID, lang string) (*QueryRecoResultForTextResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/media/voice/queryrecoresultfortext?access_token=%s&voice_id=%s&lang=%s",
		token, url.QueryEscape(voiceID), url.QueryEscape(lang))

	// 发送请求
	var result QueryRecoResultForTextResult
	if err := c.Https.Post(ctx, path, nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
