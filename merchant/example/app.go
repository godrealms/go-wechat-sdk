package main

import (
	wechat "github.com/godrealms/go-wechat-sdk/merchant/developed"
	"github.com/godrealms/go-wechat-sdk/merchant/developed/types"
	"github.com/godrealms/go-wechat-sdk/utils"
	"log"
)

func main() {
	certificate, err := utils.LoadCertificateWithPath("certificate.pem") // 通过证书的文件路径加载证书
	//certificate, err := utils.LoadCertificate("certificate")// 通过证书的文本内容加载证书
	if err != nil {
		log.Fatalf("load certificate err:%s\n", err.Error())
	}
	//privateKey, err := utils.LoadPrivateKey("privateKey")             // 通过私钥的文本内容加载私钥
	privateKey, err := utils.LoadPrivateKeyWithPath("privateKey.pem") // 通过私钥的文件路径内容加载私钥
	if err != nil {
		log.Fatalf("load private key err:%s\n", err.Error())
	}
	//publicKey, err := utils.LoadPublicKey("publicKey")             // 通过私钥的文本内容加载私钥
	publicKey, err := utils.LoadPublicKeyWithPath("publicKey.pem") // 通过私钥的文件路径内容加载私钥
	if err != nil {
		log.Fatalf("load private key err:%s\n", err.Error())
	}

	client := wechat.NewWechatClient().
		WithAppid("wx1234567890").
		WithAPIv3Key("1234567890").
		WithCertificate(certificate).
		WithPrivateKey(privateKey).
		WithPublicKey(publicKey)

	//response, err := client.Transactions(&types.Transactions{}) // APP下单
	response, err := client.ModifyTransactionsApp(&types.Transactions{}) // APP下单并获取调起参数
	if err != nil {
		return
	}
	log.Println(response)
}
