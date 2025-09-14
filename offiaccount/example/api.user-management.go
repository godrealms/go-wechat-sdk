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

	// 获取用户列表
	log.Println("=== 获取用户列表 ===")
	users, err := client.GetFans("")
	if err != nil {
		log.Printf("获取用户列表失败: %v", err)
	} else {
		log.Printf("用户总数: %d", users.Total)
		if len(users.Data.Openid) > 0 {
			log.Printf("第一个用户 OpenID: %s", users.Data.Openid[0])
		}
	}

	// 批量获取用户信息
	log.Println("\n=== 批量获取用户信息 ===")
	if users != nil && len(users.Data.Openid) > 0 {
		// 取前两个用户进行示例
		var userList []*offiaccount.UserListItem
		for i, openid := range users.Data.Openid {
			if i >= 2 {
				break
			}
			userList = append(userList, &offiaccount.UserListItem{
				Openid:   openid,
				Language: "zh_CN",
			})
		}

		batchReq := &offiaccount.BatchGetUserInfoRequest{
			UserList: userList,
		}

		userInfos, err := client.BatchGetUserInfo(batchReq)
		if err != nil {
			log.Printf("批量获取用户信息失败: %v", err)
		} else {
			log.Printf("批量获取用户信息成功，共获取 %d 个用户信息", len(userInfos.UserInfoList))
			for _, userInfo := range userInfos.UserInfoList {
				log.Printf("用户: %s, 关注状态: %d", userInfo.Openid, userInfo.Subscribe)
			}
		}
	}

	// 用户标签管理
	log.Println("\n=== 用户标签管理 ===")
	// 创建标签
	createTagResult, err := client.CreateTag("测试标签")
	if err != nil {
		log.Printf("创建标签失败: %v", err)
	} else {
		log.Printf("标签创建成功，标签ID: %d", createTagResult.Tag.Id)

		// 删除刚刚创建的标签
		_, err = client.DeleteTag(createTagResult.Tag.Id)
		if err != nil {
			log.Printf("删除标签失败: %v", err)
		} else {
			log.Printf("标签删除成功")
		}
	}
}
