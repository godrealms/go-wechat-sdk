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

	// 创建自定义菜单
	log.Println("=== 创建自定义菜单 ===")
	button1 := &offiaccount.CreateMenuButton{
		Type: "click",
		Name: "今日歌曲",
		Key:  "V1001_TODAY_MUSIC",
	}

	err := client.CreateCustomMenu(button1)
	if err != nil {
		log.Printf("创建自定义菜单失败: %v", err)
	} else {
		log.Println("自定义菜单创建成功")
	}

	// 查询自定义菜单
	log.Println("\n=== 查询自定义菜单 ===")
	menuInfo, err := client.GetMenu()
	if err != nil {
		log.Printf("查询自定义菜单失败: %v", err)
	} else {
		log.Println("自定义菜单查询成功")
		log.Printf("菜单信息: %+v", menuInfo)
	}

	// 查询当前自定义菜单信息
	log.Println("\n=== 查询当前自定义菜单信息 ===")
	selfMenuInfo, err := client.GetCurrentSelfMenuInfo()
	if err != nil {
		log.Printf("查询当前自定义菜单信息失败: %v", err)
	} else {
		log.Println("当前自定义菜单信息查询成功")
		log.Printf("菜单信息: %+v", selfMenuInfo)
	}

	// 删除自定义菜单
	log.Println("\n=== 删除自定义菜单 ===")
	err = client.DeleteMenu()
	if err != nil {
		log.Printf("删除自定义菜单失败: %v", err)
	} else {
		log.Println("自定义菜单删除成功")
	}
}
