package offiaccount

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"

	"github.com/godrealms/go-wechat-sdk/utils/wxcrypto"
)

// VerifyServerToken 微信公众号接入验证：GET 请求签名 = sha1(sort([token,timestamp,nonce]))。
// 保留为包级函数，直接转发到 wxcrypto。
func VerifyServerToken(token, signature, timestamp, nonce string) error {
	return wxcrypto.VerifyServerToken(token, signature, timestamp, nonce)
}

// EncryptedEnvelope 是微信加密模式 POST 过来的外层 XML。
type EncryptedEnvelope struct {
	XMLName    xml.Name `xml:"xml"`
	ToUserName string   `xml:"ToUserName"`
	Encrypt    string   `xml:"Encrypt"`
}

// ParseNotify 解析微信推送过来的回调 XML，返回明文字节。
//
//	r      - 原始 *http.Request；query string 里应带 msg_signature/timestamp/nonce（加密模式）
//	crypto - 若为 nil 表示明文模式
func ParseNotify(r *http.Request, crypto *MsgCrypto) ([]byte, error) {
	if r == nil {
		return nil, fmt.Errorf("offiaccount: nil request")
	}
	q := r.URL.Query()
	if r.Method == http.MethodGet {
		token := ""
		if crypto != nil {
			token = crypto.Token()
		}
		if err := VerifyServerToken(token, q.Get("signature"), q.Get("timestamp"), q.Get("nonce")); err != nil {
			return nil, err
		}
		return []byte(q.Get("echostr")), nil
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	_ = r.Body.Close()

	if crypto == nil {
		return body, nil
	}

	var env EncryptedEnvelope
	if err := xml.Unmarshal(body, &env); err != nil {
		return nil, fmt.Errorf("parse envelope: %w", err)
	}
	if env.Encrypt == "" {
		return body, nil
	}
	if !crypto.VerifySignature(q.Get("msg_signature"), q.Get("timestamp"), q.Get("nonce"), env.Encrypt) {
		return nil, fmt.Errorf("offiaccount: msg_signature invalid")
	}
	plain, _, err := crypto.Decrypt(env.Encrypt)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}
	return plain, nil
}
