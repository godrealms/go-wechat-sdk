package example

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
	selfMenuInfo, err := client.GetCurrentSelfMenuInfo()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%+v", selfMenuInfo)
}
