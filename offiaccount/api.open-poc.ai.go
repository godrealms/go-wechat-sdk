package offiaccount

import (
	"fmt"
	"net/url"
)

// TranslateContent 微信翻译
// lfrom: 源语言，zh_CN 或 en_US
// lto: 目标语言，zh_CN 或 en_US
// content: 源内容(utf8格式，最大600Byte)
func (c *Client) TranslateContent(lfrom, lto, content string) (*TranslateContentResult, error) {
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/media/voice/translatecontent?lfrom=%s&lto=%s",
		url.QueryEscape(lfrom), url.QueryEscape(lto))

	// 发送请求
	var result TranslateContentResult
	err := c.Https.Post(c.ctx, path, []byte(content), &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// AddVoiceToRecoForText 上传语音文件进行转文字识别
// format: 文件格式（只支持mp3，16k，单声道，最大1M）
// voiceID: 语音唯一标识
// lang: 语言，zh_CN 或 en_US，默认中文
// media: 语音内容
func (c *Client) AddVoiceToRecoForText(format, voiceID, lang string, media []byte) (*Resp, error) {
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/media/voice/addvoicetorecofortext?format=%s&voice_id=%s&lang=%s",
		url.QueryEscape(format), url.QueryEscape(voiceID), url.QueryEscape(lang))

	// 发送请求
	var result Resp
	err := c.Https.Post(c.ctx, path, media, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// QueryRecoResultForText 查询语音转文字结果
// voiceID: 语音唯一标识
// lang: 语言，zh_CN 或 en_US，默认中文
func (c *Client) QueryRecoResultForText(voiceID, lang string) (*QueryRecoResultForTextResult, error) {
	// 构造请求URL
	path := fmt.Sprintf("/cgi-bin/media/voice/queryrecoresultfortext?voice_id=%s&lang=%s",
		url.QueryEscape(voiceID), url.QueryEscape(lang))

	// 发送请求
	var result QueryRecoResultForTextResult
	err := c.Https.Post(c.ctx, path, nil, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
