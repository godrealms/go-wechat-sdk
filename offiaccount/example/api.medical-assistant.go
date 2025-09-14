package main

import (
	"context"
	"log"

	"github.com/godrealms/go-wechat-sdk/offiaccount"
)

func main() {
	// 初始化配置
	config := &offiaccount.Config{
		AppId:     "your-app-id",
		AppSecret: "your-app-secret",
	}

	// 创建客户端实例
	ctx := context.Background()
	client := offiaccount.NewClient(ctx, config)

	// 发送城市服务消息
	log.Println("=== 发送城市服务消息 ===")
	result, err := client.SendChannelMsg(&offiaccount.SendChannelMsgRequest{
		Status:     1501001, // 预约挂号成功通知
		OpenID:     "USER_OPENID",
		OrderID:    "ORDER_ID",
		MsgID:      "MSG_ID",
		AppID:      config.AppId,
		BusinessID: 150,
		BusinessInfo: &offiaccount.BusinessInfo{
			PatName:        "患者姓名",
			DocName:        "医生姓名",
			DepartmentName: "科室名称",
			RedirectPage: &offiaccount.RedirectPage{
				PageType: "web",
				URL:      "https://example.com/order/detail",
			},
		},
	})
	if err != nil {
		log.Printf("发送城市服务消息失败: %v", err)
	} else if result.ErrCode != 0 {
		log.Printf("发送城市服务消息失败: %s (错误码: %d)", result.ErrMsg, result.ErrCode)
	} else {
		log.Println("城市服务消息发送成功")
	}
}
