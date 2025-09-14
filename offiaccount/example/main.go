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

	// 示例1: 获取 Access Token
	log.Println("=== 获取 Access Token ===")
	token := client.GetAccessToken()
	log.Printf("Access Token: %s", token)

	// 示例2: 获取用户列表
	log.Println("\n=== 获取用户列表 ===")
	users, err := client.GetFans("")
	if err != nil {
		log.Printf("获取用户列表失败: %v", err)
	} else {
		log.Printf("用户总数: %d", users.Total)
		if len(users.Data.Openid) > 0 {
			log.Printf("第一个用户 OpenID: %s", users.Data.Openid[0])
		}
	}

	// 示例3: 发送模板消息
	log.Println("\n=== 发送模板消息 ===")
	// 注意：需要先在微信公众平台配置模板消息
	template := &offiaccount.SubscribeMessageRequest{
		ToUser:     "user_openid", // 替换为实际的用户 openid
		TemplateID: "template_id", // 替换为实际的模板 ID
		Data: map[string]interface{}{
			"first": map[string]string{
				"value": "恭喜你购买成功！",
			},
			"product": map[string]string{
				"value": "巧克力",
			},
			"amount": map[string]string{
				"value": "39.8元",
			},
		},
		URL: "http://weixin.qq.com/download",
	}

	result, err := client.SendTemplateMessage(template)
	if err != nil {
		log.Printf("发送模板消息失败: %v", err)
	} else if result.ErrCode != 0 {
		log.Printf("发送模板消息失败: %s (错误码: %d)", result.ErrMsg, result.ErrCode)
	} else {
		log.Println("模板消息发送成功")
	}

	// 示例4: 非税支付 - 查询应收信息
	log.Println("\n=== 非税支付 - 查询应收信息 ===")
	feeInfo, err := client.QueryFee(&offiaccount.QueryFeeRequest{
		AppID:           config.AppId,
		ServiceID:       123,
		PaymentNoticeNo: "PAYMENT_NOTICE_NO",
		DepartmentCode:  "DEPARTMENT_CODE",
		RegionCode:      "REGION_CODE",
	})
	if err != nil {
		log.Printf("查询应收信息失败: %v", err)
	} else if feeInfo.ErrCode != 0 {
		log.Printf("查询应收信息失败: %s (错误码: %d)", feeInfo.ErrMsg, feeInfo.ErrCode)
	} else {
		log.Printf("应收金额: %d 分", feeInfo.Fee)
		log.Printf("用户姓名: %s", feeInfo.UserName)
	}

	// 示例5: 一码通 - 申请二维码
	log.Println("\n=== 一码通 - 申请二维码 ===")
	code, err := client.ApplyCode(&offiaccount.ApplyCodeRequest{
		CodeCount:        10000,
		IsvApplicationID: "OUT_REQUEST_NO",
	})
	if err != nil {
		log.Printf("申请二维码失败: %v", err)
	} else if code.Resp.ErrCode != 0 {
		log.Printf("申请二维码失败: %s (错误码: %d)", code.Resp.ErrMsg, code.Resp.ErrCode)
	} else {
		log.Printf("申请单号: %d", code.ApplicationID)
	}

	// 示例6: 医疗助手 - 发送城市服务消息
	log.Println("\n=== 医疗助手 - 发送城市服务消息 ===")
	result2, err := client.SendChannelMsg(&offiaccount.SendChannelMsgRequest{
		Status:     1501001,       // 预约挂号成功通知
		OpenID:     "USER_OPENID", // 替换为实际的用户 openid
		OrderID:    "ORDER_ID",
		MsgID:      "MSG_ID",
		AppID:      config.AppId,
		BusinessID: 150,
		BusinessInfo: &offiaccount.BusinessInfo{
			PatName:        "张三",
			DocName:        "李医生",
			DepartmentName: "内科",
		},
	})
	if err != nil {
		log.Printf("发送城市服务消息失败: %v", err)
	} else if result2.ErrCode != 0 {
		log.Printf("发送城市服务消息失败: %s (错误码: %d)", result2.ErrMsg, result2.ErrCode)
	} else {
		log.Println("城市服务消息发送成功")
	}

	// 示例7: 财政电子票据 - 查询授权信息
	log.Println("\n=== 财政电子票据 - 查询授权信息 ===")
	authData, err := client.GetFiscalAuthData(&offiaccount.GetFiscalAuthDataRequest{
		OrderID: "ORDER_ID",
		SPAppID: "SP_APP_ID",
	})
	if err != nil {
		log.Printf("查询财政电子票据授权信息失败: %v", err)
	} else if authData.ErrCode != 0 {
		log.Printf("查询财政电子票据授权信息失败: %s (错误码: %d)", authData.ErrMsg, authData.ErrCode)
	} else {
		log.Printf("发票状态: %s", authData.InvoiceStatus)
		log.Printf("授权时间: %d", authData.AuthTime)
	}

	log.Println("\n=== 示例运行完成 ===")
}
