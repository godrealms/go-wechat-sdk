package offiaccount

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils/wxcrypto"
)

// notifyReplayWindow is the ±window applied to inbound notify timestamps to
// reject replays. Matches the oplatform / work-wechat ISV notify handlers.
const notifyReplayWindow = 5 * time.Minute

// ErrNotifyNoCrypto is returned by ParseNotify when the caller passes a nil
// *MsgCrypto. Historically the code silently returned the raw body in that
// case, which accepts completely unsigned payloads; callers who need the
// legacy plaintext mode must now use ParseNotifyPlaintext and supply a token.
var ErrNotifyNoCrypto = errors.New("offiaccount: ParseNotify requires a non-nil *MsgCrypto; use ParseNotifyPlaintext(r, token) for plaintext mode")

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

// ParseNotify 解析微信推送过来的回调 XML（加密模式），返回明文字节。
//
//	r      - 原始 *http.Request；query 里必须带 msg_signature/timestamp/nonce
//	crypto - 必须非 nil；使用其 Token + AES key 做签名校验与解密
//
// 对 POST 请求做三件事：
//  1. 校验 timestamp 必须在当前时间 ±5 分钟以内（防重放）。
//  2. 校验 msg_signature。
//  3. 用 AES key 解密 Encrypt 字段。
//
// 对 GET 请求做两件事：
//  1. 校验 timestamp 必须在当前时间 ±5 分钟以内。
//  2. 校验 signature（注意是 signature 而不是 msg_signature）。
//
// 如果用的是「明文模式」请使用 ParseNotifyPlaintext。
func ParseNotify(r *http.Request, crypto *MsgCrypto) ([]byte, error) {
	if r == nil {
		return nil, fmt.Errorf("offiaccount: nil request")
	}
	if crypto == nil {
		return nil, ErrNotifyNoCrypto
	}
	q := r.URL.Query()

	if err := checkNotifyTimestamp(q.Get("timestamp")); err != nil {
		return nil, err
	}

	if r.Method == http.MethodGet {
		if err := VerifyServerToken(crypto.Token(), q.Get("signature"), q.Get("timestamp"), q.Get("nonce")); err != nil {
			return nil, err
		}
		return []byte(q.Get("echostr")), nil
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("offiaccount: read body: %w", err)
	}
	_ = r.Body.Close()

	var env EncryptedEnvelope
	if err := xml.Unmarshal(body, &env); err != nil {
		return nil, fmt.Errorf("offiaccount: parse envelope: %w", err)
	}
	if env.Encrypt == "" {
		return nil, fmt.Errorf("offiaccount: empty Encrypt field — use ParseNotifyPlaintext for plaintext mode")
	}
	if !crypto.VerifySignature(q.Get("msg_signature"), q.Get("timestamp"), q.Get("nonce"), env.Encrypt) {
		return nil, fmt.Errorf("offiaccount: msg_signature invalid")
	}
	plain, _, err := crypto.Decrypt(env.Encrypt)
	if err != nil {
		return nil, fmt.Errorf("offiaccount: decrypt: %w", err)
	}
	return plain, nil
}

// ParseNotifyPlaintext 解析「明文模式」下 WeChat 公众号推送过来的回调。
//
// 明文模式下 POST 请求不会带 msg_signature / 加密信封，只带 signature /
// timestamp / nonce（与 GET 接入签名算法相同），因此 SDK 需要调用方把当初
// 配置在公众号后台的 Token 传进来，才能校验来源。
//
//	r     - 原始 *http.Request
//	token - 公众号后台配置的 Token
//
// 成功返回原始请求体（POST）或 echostr（GET）。校验失败返回非 nil error。
func ParseNotifyPlaintext(r *http.Request, token string) ([]byte, error) {
	if r == nil {
		return nil, fmt.Errorf("offiaccount: nil request")
	}
	if token == "" {
		return nil, fmt.Errorf("offiaccount: ParseNotifyPlaintext requires a non-empty token")
	}
	q := r.URL.Query()
	if err := checkNotifyTimestamp(q.Get("timestamp")); err != nil {
		return nil, err
	}
	if err := VerifyServerToken(token, q.Get("signature"), q.Get("timestamp"), q.Get("nonce")); err != nil {
		return nil, err
	}
	if r.Method == http.MethodGet {
		return []byte(q.Get("echostr")), nil
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("offiaccount: read body: %w", err)
	}
	_ = r.Body.Close()
	return body, nil
}

// checkNotifyTimestamp validates a notify timestamp query string against the
// local clock with a ±5-minute window (replay window).
func checkNotifyTimestamp(ts string) error {
	if ts == "" {
		return fmt.Errorf("offiaccount: missing notify timestamp")
	}
	n, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return fmt.Errorf("offiaccount: invalid notify timestamp %q: %w", ts, err)
	}
	delta := time.Now().Unix() - n
	if delta < 0 {
		delta = -delta
	}
	if delta > int64(notifyReplayWindow/time.Second) {
		return fmt.Errorf("offiaccount: notify timestamp out of ±5min window: delta=%ds", delta)
	}
	return nil
}
