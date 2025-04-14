package wechat

import (
	"context"
	"fmt"
	"github.com/godrealms/go-wechat-sdk/merchant/developed/types"
	"github.com/godrealms/go-wechat-sdk/utils"
	"time"
)

func (c *Client) TransactionsJsapi(order *types.Transactions) (*types.TransactionsJsapiResp, error) {
	path := "/v3/pay/transactions/jsapi"
	timestamp := time.Now().Unix()
	body := order.ToString()
	nonceStr := utils.GenerateHashBasedString(32)
	sign := fmt.Sprintf("%s\n%s\n%d\n%s\n%s\n", "POST", path, timestamp, nonceStr, body)
	signature, err := utils.SignSHA256WithRSA(sign, c.privateKey)
	if err != nil {
		return nil, err
	}
	c.Http.Headers = map[string]string{
		"Accept":        "application/json",
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("WECHATPAY2-SHA256-RSA2048 mchid=\"%s\",nonce_str=\"%s\",signature=\"%s\",timestamp=\"%d\",serial_no=\"%s\"", c.Mchid, nonceStr, signature, timestamp, c.CertificateNumber),
	}

	response := &types.TransactionsJsapiResp{}
	err = c.Http.Post(context.Background(), path, order, response)
	return response, err
}

// ModifyTransactionsJsapi JsAPI下单获取调起参数
func (c *Client) ModifyTransactionsJsapi(order *types.Transactions) (*types.TransactionsJsapi, error) {
	response, err := c.TransactionsJsapi(order)
	if err != nil {
		return nil, err
	}
	parameter := &types.TransactionsJsapi{
		TimeStamp: fmt.Sprintf("%d", time.Now().Unix()),
		NonceStr:  utils.GenerateHashBasedString(32),
		Package:   fmt.Sprintf("prepay_id=%s", response.PrepayId),
		SignType:  "RSA",
	}
	err = parameter.GenerateSignature(c.Appid, c.privateKey)
	return parameter, err
}
