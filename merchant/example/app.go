//go:build ignore

package main

import (
	"context"
	"log"

	pay "github.com/godrealms/go-wechat-sdk/merchant/developed"
	"github.com/godrealms/go-wechat-sdk/merchant/developed/types"
	"github.com/godrealms/go-wechat-sdk/utils"
)

func main() {
	certificate, err := utils.LoadCertificateWithPath("certificate.pem")
	if err != nil {
		log.Fatalf("load certificate err: %s", err.Error())
	}
	privateKey, err := utils.LoadPrivateKeyWithPath("privateKey.pem")
	if err != nil {
		log.Fatalf("load private key err: %s", err.Error())
	}

	client, err := pay.NewClient(pay.Config{
		Appid:             "wx1234567890",
		Mchid:             "1900000001",
		CertificateNumber: "5157F09EFDC096DE15EBE81A47057A7232F1B8E1",
		APIv3Key:          "your_apiv3_key_32_bytes_long_xxx",
		PrivateKey:        privateKey,
		Certificate:       certificate,
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// （可选）启动时主动拉取一次平台证书，后续响应验签速度更快
	if _, err := client.FetchPlatformCertificates(ctx); err != nil {
		log.Printf("fetch platform certs failed: %v", err)
	}

	resp, err := client.ModifyTransactionsApp(ctx, &types.Transactions{})
	if err != nil {
		log.Fatalf("modify transactions app failed: %v", err)
	}
	log.Println(resp)
}
