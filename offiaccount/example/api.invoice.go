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

	// 获取用户发票抬头
	log.Println("=== 获取用户发票抬头 ===")
	titleURL, err := client.GetUserTitleUrl(&offiaccount.GetUserTitleUrlRequest{
		UserFill: 1, // 用户自己填写抬头
	})
	if err != nil {
		log.Printf("获取用户发票抬头失败: %v", err)
	} else {
		log.Println("用户发票抬头获取成功")
		log.Printf("URL: %s", titleURL.Url)
	}

	// 获取选择发票抬头链接
	log.Println("\n=== 获取选择发票抬头链接 ===")
	selectTitleURL, err := client.GetSelectTitleUrl(&offiaccount.GetSelectTitleUrlRequest{
		Attach:  "附加字段",
		BizName: "商户名称",
	})
	if err != nil {
		log.Printf("获取选择发票抬头链接失败: %v", err)
	} else {
		log.Println("选择发票抬头链接获取成功")
		log.Printf("URL: %s", selectTitleURL.Url)
	}

	// 查询发票信息示例（需要实际的发票card_id和encrypt_code）
	log.Println("\n=== 查询发票信息 ===")
	// 注意：需要替换为实际的card_id和encrypt_code才能运行此示例
	/*
		invoiceInfo, err := client.GetInvoiceInfo(&offiaccount.GetInvoiceInfoRequest{
			CardID:      "CARD_ID",
			EncryptCode: "ENCRYPT_CODE",
		})
		if err != nil {
			log.Printf("查询发票信息失败: %v", err)
		} else {
			log.Println("发票信息查询成功")
			log.Printf("发票ID: %s", invoiceInfo.CardID)
			log.Printf("发票状态: %s", invoiceInfo.UserInfo.ReimburseStatus)
		}
	*/

	// 更新发票状态示例（需要实际的发票card_id和code）
	log.Println("\n=== 更新发票状态 ===")
	// 注意：需要替换为实际的card_id和code才能运行此示例
	/*
		err = client.UpdateInvoiceStatus(&offiaccount.UpdateInvoiceStatusRequest{
			CardID:          "CARD_ID",
			Code:            "CODE",
			ReimburseStatus: "INVOICE_REIMBURSE_INIT",
		})
		if err != nil {
			log.Printf("更新发票状态失败: %v", err)
		} else {
			log.Println("发票状态更新成功")
		}
	*/

	// 查询财政电子票据授权信息
	log.Println("\n=== 查询财政电子票据授权信息 ===")
	// 注意：需要替换为实际的order_id和s_pappid才能运行此示例
	/*
		authData, err := client.GetFiscalAuthData(&offiaccount.GetFiscalAuthDataRequest{
			OrderID: "ORDER_ID",
			SPAppID: "SP_APP_ID",
		})
		if err != nil {
			log.Printf("查询财政电子票据授权信息失败: %v", err)
		} else {
			log.Println("财政电子票据授权信息查询成功")
			log.Printf("发票状态: %s", authData.InvoiceStatus)
			log.Printf("授权时间: %d", authData.AuthTime)
		}
	*/

	// 获取sdk临时票据
	log.Println("\n=== 获取sdk临时票据 ===")
	ticket, err := client.GetTicket()
	if err != nil {
		log.Printf("获取sdk临时票据失败: %v", err)
	} else {
		log.Println("sdk临时票据获取成功")
		log.Printf("票据: %s", ticket.Ticket)
		log.Printf("有效期: %d秒", ticket.ExpiresIn)
	}
}
