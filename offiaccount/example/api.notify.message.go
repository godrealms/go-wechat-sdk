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
	// 上传图片
	//image, err := client.UploadImage("./bg.jpeg", "")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//log.Printf("%+v", image)
	// 发送文本消息
	//req := offiaccount.NewMassSendRequest(
	//	[]string{"o-wNfxCzha_pcGj7BPOC0rZZvqPE", "o-wNfxC7ZhqMus3jIbDkiLqAPaxo"},
	//	offiaccount.MsgTypeText,
	//).SetText("Hello World!").SetClientMsgID("25125773061322904")
	//resp, err := client.MassSend(req)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//log.Printf("%+v", resp)
	//resp, err := client.Preview(req)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//log.Printf("%+v", resp)

	req1 := offiaccount.NewMassSendToTag(offiaccount.MsgTypeText, "me").
		SetText("新年快乐！")
	resp, err := client.SendAll(req1)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%+v", resp)

	//info, err := client.GetRidInfo("689b66d3-57545a3a-00b87a82")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//log.Printf("%+v", info)
}
