package types

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

// Resource
// 【通知数据】通知资源数据。
type Resource struct {
	//【原始回调类型】加密前的对象类型，为transaction。
	OriginalType string `json:"original_type"`
	//【加密算法类型】回调数据密文的加密算法类型，目前为AEAD_AES_256_GCM，开发者需要使用同样类型的数据进行解密。
	Algorithm string `json:"algorithm"`
	//【数据密文】Base64编码后的回调数据密文，商户需Base64解码并使用APIV3密钥解密，具体参考如何解密证书和回调报文。
	Ciphertext string `json:"ciphertext"`
	//【附加数据】参与解密的附加数据，该字段可能为空。
	AssociatedData string `json:"associated_data"`
	//【随机串】参与解密的随机串。
	Nonce string `json:"nonce"`
}

// Transaction TransactionResource
// resource解密后字段
type Transaction struct {
	TransactionId   string             `json:"transaction_id"`
	Amount          *Amount            `json:"amount"`
	Mchid           string             `json:"mchid"`
	TradeState      TradeState         `json:"trade_state"`
	BankType        string             `json:"bank_type"`
	PromotionDetail []*PromotionDetail `json:"promotion_detail"`
	SuccessTime     time.Time          `json:"success_time"`
	Payer           *Payer             `json:"payer"`
	OutTradeNo      string             `json:"out_trade_no"`
	Appid           string             `json:"appid"`
	TradeStateDesc  string             `json:"trade_state_desc"`
	TradeType       string             `json:"trade_type"`
	Attach          string             `json:"attach"`
	SceneInfo       *SceneInfo         `json:"scene_info"`
}

// Notify is the decrypted payload of a WeChat Pay callback notification.
type Notify struct {
	//【通知ID】回调通知的唯一编号。
	Id string `json:"id"`
	//【通知创建时间】
	//	1、定义：本次回调通知创建的时间。
	//	2、格式：遵循rfc3339标准格式：yyyy-MM-DDTHH:mm:ss+TIMEZONE。
	//	yyyy-MM-DD 表示年月日；
	//	T 字符用于分隔日期和时间部分；
	//	HH:mm:ss 表示具体的时分秒；
	//	TIMEZONE 表示时区（例如，+08:00 对应东八区时间，即北京时间）。
	//	example：2015-05-20T13:29:35+08:00 表示北京时间2015年5月20日13点29分35秒。
	CreateTime time.Time `json:"create_time"`
	//【通知的类型】微信支付回调通知的类型。
	//	支付成功通知的类型为: TRANSACTION.SUCCESS。
	//	退款成功通知为: encrypt-resource
	ResourceType string `json:"resource_type"`
	//【通知的类型】微信支付回调通知的类型。
	//	支付成功通知的类型为TRANSACTION.SUCCESS。
	EventType string `json:"event_type"`
	//【回调摘要】微信支付对回调内容的摘要备注。
	Summary string `json:"summary"`
	//【通知数据】通知资源数据。
	Resource *Resource `json:"resource"`
}

// VerifySignature 验证签名
// timestamp: 时间戳
// nonce: 随机串
// body: 请求体
// signature: 签名
//
// Deprecated: 不要用这个方法。它存在三个独立的安全/正确性问题：
//  1. 静默吞错: 当 publicKey PEM 解析失败、不是 RSA 公钥、或 signature
//     base64 解码失败时，方法返回 false。调用方无法区分 "签名校验失败"
//     和 "调用方传错了公钥" 这两种完全不同的情况——后者意味着所有合法
//     回调都会被错误地拒绝。
//  2. 不跟随平台证书轮换: 该方法解析的是调用方传入的 PKIX 公钥，而
//     微信支付平台证书会按月轮换。一旦 WeChat 切换到新证书，所有
//     之前的回调都会被这个方法误判为无效。
//  3. 无时间戳/nonce 重放检查: 单纯校验签名不足以拒绝重放攻击。
//
// 正确做法: 调用 pay.Client.ParseNotification——它会用本地缓存或自动
// 拉取的最新平台证书校验签名、强制 ±5 分钟时间戳窗口、并返回结构化的
// 错误以区分以上三类失败。
func (n *Notify) VerifySignature(timestamp, nonce, signature, body, publicKey string) bool {
	// 构造验签串
	message := fmt.Sprintf("%s\n%s\n%s\n", timestamp, nonce, body)

	// 解析公钥
	block, _ := pem.Decode([]byte(publicKey))
	if block == nil {
		return false
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return false
	}

	rsaPublicKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return false
	}

	// Base64解码签名
	signatureBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false
	}

	// 计算消息摘要
	hashed := sha256.Sum256([]byte(message))

	// 验证签名
	err = rsa.VerifyPKCS1v15(rsaPublicKey, crypto.SHA256, hashed[:], signatureBytes)
	return err == nil
}

// DecryptResource 解密数据
func (n *Notify) DecryptResource(apiV3Key string) (*Transaction, error) {
	dataBytes, err := n.DecryptAES256GCM(apiV3Key)
	if err != nil {
		return nil, err
	}
	var transaction Transaction
	err = json.Unmarshal(dataBytes, &transaction)
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

// DecryptAES256GCM 使用 AEAD_AES_256_GCM 算法解密 n.Resource.Ciphertext。
// 官方实现已抽到 utils.DecryptAEADAES256GCM,此处为薄包装,保持 API 兼容。
//
// 详见 WeChat 文档:
// https://wechatpay-api.gitbook.io/wechatpay-api-v3/qian-ming-zhi-nan-1/zheng-shu-he-hui-tiao-bao-wen-jie-mi
func (n *Notify) DecryptAES256GCM(aesKey string) (plaintext []byte, err error) {
	if n.Resource == nil {
		return nil, fmt.Errorf("notify resource is nil")
	}
	return utils.DecryptAEADAES256GCM(aesKey, n.Resource.Nonce, n.Resource.AssociatedData, n.Resource.Ciphertext)
}

// IsPaymentSuccess 判断是否支付成功
func (n *Notify) IsPaymentSuccess() bool {
	return n.EventType == "TRANSACTION.SUCCESS"
}
