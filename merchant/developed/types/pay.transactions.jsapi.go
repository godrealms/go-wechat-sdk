package types

import (
	"crypto/rsa"
	"fmt"
	"github.com/godrealms/go-wechat-sdk/utils"
)

type TransactionsJsapiResp struct {
	PrepayId string `json:"prepay_id"`
}

type TransactionsJsapi struct {
	//【时间戳】标准北京时间，时区为东八区，自1970年1月1日 0点0分0秒以来的秒数。
	//	注意：部分系统取到的值为毫秒级，商户需要转换成秒(10位数字)。
	TimeStamp string `json:"timeStamp"`
	//【随机字符串】不长于32位。该值建议使用随机数算法生成。
	NonceStr string `json:"nonceStr"`
	//【预支付交易会话标识】JSAPI/小程序下单接口返回的prepay_id参数值，提交格式如：prepay_id=***
	Package string `json:"package"`
	//【签名类型】默认为RSA，仅支持RSA。
	SignType string `json:"signType"`
	//【签名值】使用字段appId、timeStamp、nonceStr、package计算得出的签名值，
	//	详细参考：小程序调起支付签名(https://pay.weixin.qq.com/doc/v3/merchant/4012365341)。
	//	注意：此处签名需使用实际调起支付小程序appid，且为JSAPI/小程序下单时传入的appid，
	//	微信支付会校验下单与调起支付所使用的appid的一致性。
	PaySign string `json:"paySign"`
}

func (j *TransactionsJsapi) GenerateSignature(appid string, privateKey *rsa.PrivateKey) error {
	str := fmt.Sprintf("%s\n%s\n%s\n%s\n", appid, j.TimeStamp, j.NonceStr, j.Package)
	sign, err := utils.SignSHA256WithRSA(str, privateKey)
	if err != nil {
		return err
	}
	j.PaySign = sign
	return nil
}
