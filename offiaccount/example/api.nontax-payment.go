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
	_ = offiaccount.NewClient(ctx, config)

	// 查询应收信息
	log.Println("=== 查询应收信息 ===")
	// 注意：需要替换为实际的参数才能运行此示例
	/*
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
			log.Println("应收信息查询成功")
			log.Printf("应收金额: %d 分", feeInfo.Fee)
			log.Printf("用户姓名: %s", feeInfo.UserName)
			log.Printf("缴费子项目数量: %d", len(feeInfo.Items))
		}
	*/

	// 缴费支付下单
	log.Println("\n=== 缴费支付下单 ===")
	// 注意：需要替换为实际的参数才能运行此示例
	/*
		order, err := client.UnifiedOrder(&offiaccount.UnifiedOrderRequest{
			AppID:           config.AppId,
			Desc:            "缴费描述",
			Fee:             100, // 单位：分
			IP:              "127.0.0.1",
			PaymentNoticeNo: "PAYMENT_NOTICE_NO",
			DepartmentCode:  "DEPARTMENT_CODE",
			DepartmentName:  "DEPARTMENT_NAME",
			RegionCode:      "REGION_CODE",
			Items: []offiaccount.NonTaxItem{
				{
					No:       1,
					ItemID:   "ITEM_ID",
					ItemName: "项目名称",
					Fee:      100,
				},
			},
			PaymentNoticeCreateTime: time.Now().Unix(),
			Scene: "biz",
		})
		if err != nil {
			log.Printf("缴费支付下单失败: %v", err)
		} else if order.ErrCode != 0 {
			log.Printf("缴费支付下单失败: %s (错误码: %d)", order.ErrMsg, order.ErrCode)
		} else {
			log.Println("缴费支付下单成功")
			log.Printf("支付订单号: %s", order.OrderID)
			log.Printf("支付跳转链接: %s", order.PayURL)
		}
	*/

	// 下载缴费对账单
	log.Println("\n=== 下载缴费对账单 ===")
	// 注意：需要替换为实际的参数才能运行此示例
	/*
		bill, err := client.DownloadBill(&offiaccount.DownloadBillRequest{
			AppID:  config.AppId,
			Date:   "20230101",
			Type:   "ALL",
			Scene:  "biz",
		})
		if err != nil {
			log.Printf("下载缴费对账单失败: %v", err)
		} else if bill.ErrCode != 0 {
			log.Printf("下载缴费对账单失败: %s (错误码: %d)", bill.ErrMsg, bill.ErrCode)
		} else {
			log.Println("缴费对账单下载成功")
			log.Printf("账单数据: %s", bill.Data)
		}
	*/

	// 获取缴费订单列表
	log.Println("\n=== 获取缴费订单列表 ===")
	// 注意：需要替换为实际的payment_notice_no才能运行此示例
	/*
		orderList, err := client.GetOrderList(&offiaccount.GetOrderListRequest{
			PaymentNoticeNo: "PAYMENT_NOTICE_NO",
		})
		if err != nil {
			log.Printf("获取缴费订单列表失败: %v", err)
		} else if orderList.ErrCode != 0 {
			log.Printf("获取缴费订单列表失败: %s (错误码: %d)", orderList.ErrMsg, orderList.ErrCode)
		} else {
			log.Println("缴费订单列表获取成功")
			log.Printf("订单ID列表: %v", orderList.OrderIDList)
			log.Printf("支付订单ID列表: %v", orderList.PayOrderIDList)
		}
	*/

	// 获取缴费订单详情
	log.Println("\n=== 获取缴费订单详情 ===")
	// 注意：需要替换为实际的order_id才能运行此示例
	/*
		orderDetail, err := client.GetOrder(&offiaccount.GetOrderRequest{
			OrderID: "ORDER_ID",
		})
		if err != nil {
			log.Printf("获取缴费订单详情失败: %v", err)
		} else if orderDetail.ErrCode != 0 {
			log.Printf("获取缴费订单详情失败: %s (错误码: %d)", orderDetail.ErrMsg, orderDetail.ErrCode)
		} else {
			log.Println("缴费订单详情获取成功")
			log.Printf("订单状态: %d", orderDetail.Status)
			log.Printf("订单金额: %d 分", orderDetail.Fee)
		}
	*/
}
