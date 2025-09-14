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

	// 申请二维码
	log.Println("=== 申请二维码 ===")
	code, err := client.ApplyCode(&offiaccount.ApplyCodeRequest{
		CodeCount:        10000,
		IsvApplicationID: "OUT_REQUEST_NO",
	})
	if err != nil {
		log.Printf("申请二维码失败: %v", err)
	} else if code.Resp.ErrCode != 0 {
		log.Printf("申请二维码失败: %s (错误码: %d)", code.Resp.ErrMsg, code.Resp.ErrCode)
	} else {
		log.Println("二维码申请成功")
		log.Printf("申请单号: %d", code.ApplicationID)
	}

	// 查询二维码申请单
	log.Println("\n=== 查询二维码申请单 ===")
	// 注意：需要替换为实际的application_id才能运行此示例
	/*
		query, err := client.ApplyCodeQuery(&offiaccount.ApplyCodeQueryRequest{
			ApplicationID: 123456,
		})
		if err != nil {
			log.Printf("查询二维码申请单失败: %v", err)
		} else if query.Resp.ErrCode != 0 {
			log.Printf("查询二维码申请单失败: %s (错误码: %d)", query.Resp.ErrMsg, query.Resp.ErrCode)
		} else {
			log.Println("二维码申请单查询成功")
			log.Printf("申请单状态: %s", query.Status)
			log.Printf("二维码信息数量: %d", len(query.CodeGenerateList))
		}
	*/

	// 激活二维码
	log.Println("\n=== 激活二维码 ===")
	// 注意：需要替换为实际的参数才能运行此示例
	/*
		active, err := client.CodeActive(&offiaccount.CodeActiveRequest{
			ApplicationID: 123456,
			ActivityName:  "活动名称",
			ProductBrand:  "商品品牌",
			ProductTitle:  "商品标题",
			ProductCode:   "商品条码",
			WxaAppid:      "小程序的appid",
			WxaPath:       "小程序的path",
			CodeStart:     0,
			CodeEnd:       9999,
		})
		if err != nil {
			log.Printf("激活二维码失败: %v", err)
		} else if active.ErrCode != 0 {
			log.Printf("激活二维码失败: %s (错误码: %d)", active.ErrMsg, active.ErrCode)
		} else {
			log.Println("二维码激活成功")
		}
	*/

	// 查询二维码激活状态
	log.Println("\n=== 查询二维码激活状态 ===")
	// 注意：需要替换为实际的参数才能运行此示例
	/*
		activeQuery, err := client.CodeActiveQuery(&offiaccount.CodeActiveQueryRequest{
			ApplicationID: 123456,
			ActiveCode:    "ACTIVE_CODE",
		})
		if err != nil {
			log.Printf("查询二维码激活状态失败: %v", err)
		} else if activeQuery.Resp.ErrCode != 0 {
			log.Printf("查询二维码激活状态失败: %s (错误码: %d)", activeQuery.Resp.ErrMsg, activeQuery.Resp.ErrCode)
		} else {
			log.Println("二维码激活状态查询成功")
			log.Printf("激活状态: %s", activeQuery.Status)
		}
	*/

	// CODE_TICKET换CODE
	log.Println("\n=== CODE_TICKET换CODE ===")
	// 注意：需要替换为实际的ticket才能运行此示例
	/*
		ticketCode, err := client.TicketToCode(&offiaccount.TicketToCodeRequest{
			Ticket: "TICKET",
		})
		if err != nil {
			log.Printf("CODE_TICKET换CODE失败: %v", err)
		} else if ticketCode.Resp.ErrCode != 0 {
			log.Printf("CODE_TICKET换CODE失败: %s (错误码: %d)", ticketCode.Resp.ErrMsg, ticketCode.Resp.ErrCode)
		} else {
			log.Println("CODE_TICKET换CODE成功")
			log.Printf("营销码: %s", ticketCode.Code)
		}
	*/
}
