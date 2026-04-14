//go:build ignore

package main

import (
	"context"
	"log"

	"github.com/godrealms/go-wechat-sdk/offiaccount"
)

func main() {
	ctx := context.Background()
	config := &offiaccount.Config{
		AppId:     "your app id",
		AppSecret: "your app secret",
	}
	client := offiaccount.NewClient(ctx, config)
	// 查询自定义菜单信息
	//selfMenuInfo, err := client.GetCurrentSelfMenuInfo(ctx)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//log.Printf("%+v", selfMenuInfo)
	// 获取自定义菜单配置
	//menus, err := client.GetMenu(ctx)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//log.Printf("%+v", menus)
	// 删除自定义菜单
	err := client.DeleteMenu(ctx)
	if err != nil {
		log.Fatal(err)
	}
	//info, err := client.GetRidInfo(ctx, "")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//log.Printf("%+v", info)
}
