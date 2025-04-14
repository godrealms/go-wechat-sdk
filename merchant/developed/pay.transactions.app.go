package wechat

import (
	"context"
	"fmt"
	"github.com/godrealms/go-wechat-sdk/merchant/developed/types"
	"github.com/godrealms/go-wechat-sdk/utils"
	"time"
)

// TransactionsApp APP下单
func (c *Client) TransactionsApp(order *types.Transactions) (*types.TransactionsAppResponse, error) {
	response := &types.TransactionsAppResponse{}
	// 处理签名
	path := "/v3/pay/transactions/app"
	bodyJson := order.ToString()
	nonceStr := utils.GenerateHashBasedString(32)
	timestamp := time.Now().Unix()
	sign := fmt.Sprintf("%s\n%s\n%d\n%s\n%s\n", "POST", path, timestamp, nonceStr, bodyJson) // 签名体
	signature, err := utils.SignSHA256WithRSA(sign, c.privateKey)
	if err != nil {
		return nil, err
	}
	c.Http.Headers = map[string]string{
		"Accept":        "application/json",
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("WECHATPAY2-SHA256-RSA2048 mchid=\"%s\",nonce_str=\"%s\",signature=\"%s\",timestamp=\"%d\",serial_no=\"%s\"", c.Mchid, nonceStr, signature, timestamp, c.CertificateNumber),
	}
	err = c.Http.Post(context.Background(), path, order, response)
	return response, err
}

// ModifyTransactionsApp APP下单获取调起参数
func (c *Client) ModifyTransactionsApp(order *types.Transactions) (*types.ModifyAppResponse, error) {
	response, err := c.TransactionsApp(order)
	if err != nil {
		return nil, err
	}
	parameter := &types.ModifyAppResponse{
		AppId:        c.Appid,
		PartnerId:    c.Mchid,
		PrepayId:     response.PrepayId,
		PackageValue: "Sign=WXPay",
		NonceStr:     utils.GenerateHashBasedString(32),
		TimeStamp:    fmt.Sprintf("%d", time.Now().Unix()),
	}
	err = parameter.GenerateSignature(c.privateKey)
	return parameter, err
}
