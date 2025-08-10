package offiaccount

import (
	"errors"
	"fmt"
	"net/url"
)

// CreateCustomMenu 创建自定义菜单
// 该接口用于创建公众号/服务号的自定义菜单。
func (c *Client) CreateCustomMenu(body *CustomMenu) error {
	path := fmt.Sprintf("/cgi-bin/menu/create?access_token=%s", c.GetAccessToken())
	result := &Resp{}
	err := c.Https.Post(c.ctx, path, body, result)
	if err != nil {
		return err
	} else if result.ErrCode != 0 {
		return errors.New(result.ErrMsg)
	}
	return nil
}

type T struct {
	// 菜单是否开启，0代表未开启，1代表开启
	IsMenuOpen int `json:"is_menu_open"`
	// 自定义菜单信息
	SelfMenuInfo struct {
		// 菜单按钮
		Button []struct {
			// 菜单的类型，具体参考新增 API 的描述
			Type string `json:"type,omitempty"`
			// 菜单名称
			Name string `json:"name"`
			// 对于不同的菜单类型，value的值意义不同。
			//	官网上设置的自定义菜单：
			//	Text:保存文字到value；
			//	Img、voice：保存mediaID到value；
			//	Video：保存视频下载链接到value；
			//	News：保存图文消息到news_info，
			//	同时保存mediaID到value；
			//	View：保存链接到url。
			//	使用API设置的自定义菜单： click、scancode_push、scancode_waitmsg、pic_sysphoto、pic_photo_or_album、 pic_weixin、location_select：保存值到key；
			//	view：保存链接到url
			Key string `json:"key,omitempty"`
			// 对于不同的菜单类型，value的值意义不同。
			// 官网上设置的自定义菜单：
			// 	Text:保存文字到value；
			//	Img、voice：保存mediaID到value；
			//	Video：保存视频下载链接到value；
			//	News：保存图文消息到news_info，同时保存mediaID到value；
			//	View：保存链接到url。
			//	使用API设置的自定义菜单： click、scancode_push、scancode_waitmsg、pic_sysphoto、pic_photo_or_album、 pic_weixin、location_select：保存值到key；
			//	view：保存链接到url
			Value string `json:"value,omitempty"`
			// 对于不同的菜单类型，value的值意义不同。
			//	官网上设置的自定义菜单：
			//	Text:保存文字到value；
			//	Img、voice：保存mediaID到value；
			//	Video：保存视频下载链接到value；
			//	News：保存图文消息到news_info，同时保存mediaID到value；
			//	View：保存链接到url。
			//	使用API设置的自定义菜单： click、scancode_push、scancode_waitmsg、pic_sysphoto、pic_photo_or_album、 pic_weixin、location_select：保存值到key；
			//	view：保存链接到url
			Url      string `json:"url,omitempty"`
			NewsInfo struct {
				// 图文消息列表
				List []struct {
					// // 标题
					Title string `json:"title"`
					// // 摘要
					Digest string `json:"digest"`
					// // 作者
					Author string `json:"author"`
					// // 是否显示封面，0为不显示，1为显示
					ShowCover int `json:"show_cover"`
					// // 封面图片的URL
					CoverUrl string `json:"cover_url"`
					// // 正文的URL
					ContentUrl string `json:"content_url"`
					// // 原文的URL，若置空则无查看原文入口
					SourceUrl string `json:"source_url"`
				} `json:"list"`
			} `json:"news_info,omitempty"`
			SubButton struct {
				//
				List []struct {
					Type     string `json:"type"`
					Name     string `json:"name"`
					Url      string `json:"url,omitempty"`
					Value    string `json:"value,omitempty"`
					NewsInfo struct {
						List []struct {
							Title      string `json:"title"`
							Author     string `json:"author"`
							Digest     string `json:"digest"`
							ShowCover  int    `json:"show_cover"`
							CoverUrl   string `json:"cover_url"`
							ContentUrl string `json:"content_url"`
							SourceUrl  string `json:"source_url"`
						} `json:"list"`
					} `json:"news_info,omitempty"`
				} `json:"list"`
			} `json:"sub_button,omitempty"`
		} `json:"button"`
	} `json:"selfmenu_info"`
}

// GetCurrentSelfMenuInfo 获取自定义菜单配置
// 本接口提供公众号当前使用的自定义菜单的配置，如果公众号是通过API调用设置的菜单，则返回菜单的开发配置，
// 而如果公众号是在公众平台官网通过网站功能发布菜单，则本接口返回运营者设置的菜单配置。
func (c *Client) GetCurrentSelfMenuInfo() (map[string]interface{}, error) {
	query := url.Values{
		"access_token": {c.GetAccessToken()},
	}
	result := make(map[string]interface{})
	err := c.Https.Get(c.ctx, "/cgi-bin/get_current_selfmenu_info", query, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
