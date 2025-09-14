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

	// 创建临时整型二维码
	log.Println("=== 创建临时整型二维码 ===")
	tempQRCode := &offiaccount.CreateQRCodeRequest{
		ExpireSeconds: 604800, // 7天有效期
		ActionName:    "QR_SCENE",
		ActionInfo: offiaccount.ActionInfo{
			Scene: offiaccount.Scene{
				SceneID: 123,
			},
		},
	}

	result, err := client.CreateQRCode(tempQRCode)
	if err != nil {
		log.Printf("创建临时整型二维码失败: %v", err)
	} else {
		log.Printf("二维码创建成功")
		log.Printf("Ticket: %s", result.Ticket)
		log.Printf("过期时间: %d秒", result.ExpireSeconds)
		log.Printf("二维码URL: %s", result.URL)

		// 获取二维码图片URL
		qrCodeURL := client.GetQRCodeURL(result.Ticket)
		log.Printf("二维码图片URL: %s", qrCodeURL)
	}

	// 创建临时字符串二维码
	log.Println("\n=== 创建临时字符串二维码 ===")
	tempStrQRCode := &offiaccount.CreateQRCodeRequest{
		ExpireSeconds: 604800, // 7天有效期
		ActionName:    "QR_STR_SCENE",
		ActionInfo: offiaccount.ActionInfo{
			Scene: offiaccount.Scene{
				SceneStr: "test_scene",
			},
		},
	}

	result2, err := client.CreateQRCode(tempStrQRCode)
	if err != nil {
		log.Printf("创建临时字符串二维码失败: %v", err)
	} else {
		log.Printf("二维码创建成功")
		log.Printf("Ticket: %s", result2.Ticket)
		log.Printf("过期时间: %d秒", result2.ExpireSeconds)
		log.Printf("二维码URL: %s", result2.URL)

		// 获取二维码图片URL
		qrCodeURL := client.GetQRCodeURL(result2.Ticket)
		log.Printf("二维码图片URL: %s", qrCodeURL)
	}

	// 创建永久整型二维码
	log.Println("\n=== 创建永久整型二维码 ===")
	permanentQRCode := &offiaccount.CreateQRCodeRequest{
		ActionName: "QR_LIMIT_SCENE",
		ActionInfo: offiaccount.ActionInfo{
			Scene: offiaccount.Scene{
				SceneID: 123,
			},
		},
	}

	result3, err := client.CreateQRCode(permanentQRCode)
	if err != nil {
		log.Printf("创建永久整型二维码失败: %v", err)
	} else {
		log.Printf("二维码创建成功")
		log.Printf("Ticket: %s", result3.Ticket)
		log.Printf("二维码URL: %s", result3.URL)

		// 获取二维码图片URL
		qrCodeURL := client.GetQRCodeURL(result3.Ticket)
		log.Printf("二维码图片URL: %s", qrCodeURL)
	}

	// 创建永久字符串二维码
	log.Println("\n=== 创建永久字符串二维码 ===")
	permanentStrQRCode := &offiaccount.CreateQRCodeRequest{
		ActionName: "QR_LIMIT_STR_SCENE",
		ActionInfo: offiaccount.ActionInfo{
			Scene: offiaccount.Scene{
				SceneStr: "permanent_scene",
			},
		},
	}

	result4, err := client.CreateQRCode(permanentStrQRCode)
	if err != nil {
		log.Printf("创建永久字符串二维码失败: %v", err)
	} else {
		log.Printf("二维码创建成功")
		log.Printf("Ticket: %s", result4.Ticket)
		log.Printf("二维码URL: %s", result4.URL)

		// 获取二维码图片URL
		qrCodeURL := client.GetQRCodeURL(result4.Ticket)
		log.Printf("二维码图片URL: %s", qrCodeURL)
	}
}
