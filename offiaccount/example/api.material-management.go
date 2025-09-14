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

	// 上传临时素材示例（需要有实际的图片文件）
	log.Println("=== 上传临时素材 ===")
	// 注意：需要确保有实际的图片文件才能运行此示例
	/*
		media, err := client.UploadTemporaryMedia("image", "./test.jpg")
		if err != nil {
			log.Printf("上传临时素材失败: %v", err)
		} else {
			log.Printf("临时素材上传成功，媒体ID: %s", media.MediaID)
			log.Printf("上传时间: %d", media.CreatedAt)
		}
	*/

	// 上传永久素材示例（需要有实际的图片文件）
	log.Println("\n=== 上传永久素材 ===")
	// 注意：需要确保有实际的图片文件才能运行此示例
	/*
		permanentMedia, err := client.UploadPermanentMedia("image", "./test.jpg")
		if err != nil {
			log.Printf("上传永久素材失败: %v", err)
		} else {
			log.Printf("永久素材上传成功，媒体ID: %s", permanentMedia.MediaID)
		}
	*/

	// 获取临时素材示例（需要有实际的media_id）
	log.Println("\n=== 获取临时素材 ===")
	// 注意：需要替换为实际的media_id才能运行此示例
	/*
		err := client.GetTemporaryMedia("MEDIA_ID", "./download.jpg")
		if err != nil {
			log.Printf("获取临时素材失败: %v", err)
		} else {
			log.Println("临时素材获取成功，已保存到 download.jpg")
		}
	*/

	// 获取永久素材示例（需要有实际的media_id）
	log.Println("\n=== 获取永久素材 ===")
	// 注意：需要替换为实际的media_id才能运行此示例
	/*
		media, err := client.GetPermanentMedia("MEDIA_ID")
		if err != nil {
			log.Printf("获取永久素材失败: %v", err)
		} else {
			log.Println("永久素材获取成功")
			log.Printf("素材信息: %+v", media)
		}
	*/

	// 获取素材总数
	log.Println("\n=== 获取素材总数 ===")
	count, err := client.GetMaterialCount()
	if err != nil {
		log.Printf("获取素材总数失败: %v", err)
	} else {
		log.Println("素材总数获取成功")
		log.Printf("语音素材数量: %d", count.VoiceCount)
		log.Printf("视频素材数量: %d", count.VideoCount)
		log.Printf("图片素材数量: %d", count.ImageCount)
		log.Printf("图文素材数量: %d", count.NewsCount)
	}

	// 获取素材列表
	log.Println("\n=== 获取素材列表 ===")
	listReq := &offiaccount.BatchGetMaterialRequest{
		Type:   "image",
		Offset: 0,
		Count:  10,
	}

	list, err := client.BatchGetMaterial(listReq)
	if err != nil {
		log.Printf("获取素材列表失败: %v", err)
	} else {
		log.Println("素材列表获取成功")
		log.Printf("素材总数: %d", list.TotalCount)
		log.Printf("本次获取数量: %d", list.ItemCount)
		log.Printf("素材项数量: %d", len(list.Item))
	}
}
