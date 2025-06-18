package types

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
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

// Notify 支付成功回调通知
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
	//	示例：2015-05-20T13:29:35+08:00 表示北京时间2015年5月20日13点29分35秒。
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

// 常量定义
const (
	EventTypeTransactionSuccess = "TRANSACTION.SUCCESS" // 支付成功
	EventTypeRefundSuccess      = "REFUND.SUCCESS"      // 退款成功
)

// DecryptResource 解密回调通知资源数据
func (n *Notify) DecryptResource(apiV3Key string) (*Transaction, error) {
	if n.Resource == nil {
		return nil, fmt.Errorf("resource is nil")
	}

	// Base64解码密文
	ciphertext, err := base64.StdEncoding.DecodeString(n.Resource.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("base64 decode ciphertext failed: %v", err)
	}

	// Base64解码随机串
	nonce, err := base64.StdEncoding.DecodeString(n.Resource.Nonce)
	if err != nil {
		return nil, fmt.Errorf("base64 decode nonce failed: %v", err)
	}

	// 解密数据
	plaintext, err := decryptAESGCM([]byte(apiV3Key), nonce, ciphertext, []byte(n.Resource.AssociatedData))
	if err != nil {
		return nil, fmt.Errorf("decrypt failed: %v", err)
	}

	// 解析JSON
	var transaction Transaction
	if err := json.Unmarshal(plaintext, &transaction); err != nil {
		return nil, fmt.Errorf("unmarshal transaction failed: %v", err)
	}

	return &transaction, nil
}

// decryptAESGCM AES-GCM解密
func decryptAESGCM(key, nonce, ciphertext, additionalData []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, additionalData)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// IsPaymentSuccess 判断是否支付成功
func (n *Notify) IsPaymentSuccess() bool {
	return n.EventType == EventTypeTransactionSuccess
}

// IsRefundSuccess 判断是否退款成功
func (n *Notify) IsRefundSuccess() bool {
	return n.EventType == EventTypeRefundSuccess
}

// GetTransactionInfo 获取交易信息（需要先解密）
func (t *Transaction) GetTransactionInfo() map[string]interface{} {
	return map[string]interface{}{
		"transaction_id":   t.TransactionId,
		"out_trade_no":     t.OutTradeNo,
		"trade_state":      t.TradeState,
		"trade_state_desc": t.TradeStateDesc,
		"trade_type":       t.TradeType,
		"success_time":     t.SuccessTime,
		"total_amount":     t.Amount.Total,
		"payer_openid":     t.Payer.Openid,
		"attach":           t.Attach,
	}
}

// IsTradeSuccess 判断交易是否成功
func (t *Transaction) IsTradeSuccess() bool {
	return t.TradeState == TradeStateSUCCESS
}
