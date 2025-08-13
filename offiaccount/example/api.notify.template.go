package main

import (
	"context"

	"github.com/godrealms/go-wechat-sdk/offiaccount"
)

func main() {
	// 构建订阅消息请求
	request := offiaccount.SubscribeMessageRequest{
		ToUser:     "user_openid_here",
		TemplateID: "template_id_here",
		URL:        "https://example.com/redirect",
		MiniProgram: &offiaccount.MiniProgram{
			AppID:    "mini_program_appid",
			PagePath: "pages/index/index",
		},
		Data: map[string]interface{}{
			"thing1": offiaccount.ThingData{
				Value: "订单商品名称",
			},
			"character_string2": offiaccount.CharacterStringData{
				Value: "ORDER123456789",
			},
			"time3": offiaccount.TimeData{
				Value: "2024-01-01 15:30:00",
			},
			"amount4": offiaccount.AmountData{
				Value: "¥99.99元",
			},
			"phone_number5": offiaccount.PhoneNumberData{
				Value: "+86-138-0000-0000",
			},
			"const6": offiaccount.ConstData{
				Value: "支付成功",
			},
		},
		ClientMsgID: "unique_client_msg_id_123",
	}

	ctx := context.Background()
	config := &offiaccount.Config{
		AppId:     "your app id",
		AppSecret: "your app secret",
	}
	client := offiaccount.NewClient(ctx, config)
	client.SendTemplateMessage(&request)
}
