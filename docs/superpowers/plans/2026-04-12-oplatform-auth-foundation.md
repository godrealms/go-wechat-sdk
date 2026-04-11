# oplatform Auth Foundation + Authorizer Framework + QR Login — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement WeChat Open Platform sub-projects 1 (component auth core), 2 (authorizer 代调用 via TokenSource injection), and 6 (QR Login).

**Architecture:** Single flat `oplatform` package; shared `Store` interface with `MemoryStore` default; `utils/wxcrypto` extracted from `offiaccount/crypto.go`; `TokenSource` interface injected into `offiaccount`/`mini_program` clients so one API surface works for both self-owned and authorized identities; lazy token refresh with explicit `RefreshAll`.

**Tech Stack:** Go 1.23, standard library only (AES-CBC, SHA1, encoding/xml, encoding/json, net/http), existing `utils.HTTP` helper, `httptest` for integration tests.

**Spec:** `docs/superpowers/specs/2026-04-12-oplatform-auth-foundation-design.md`

**Module path:** `github.com/godrealms/go-wechat-sdk`

---

## Conventions

- Every task ends with a commit. Keep commits atomic and small.
- TDD: write the failing test first, verify failure, then minimal code to pass, verify pass, commit.
- Use `go test ./...` to verify nothing else broke before committing.
- All file paths are relative to the repo root `/Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk`.
- Never use `gofmt -w`; just write consistent formatting from the start (tabs, blank lines between funcs).

---

## Task 1: Extract `utils/wxcrypto` package

**Goal:** Move Biz Msg Crypt algorithm out of `offiaccount` into a shared `utils/wxcrypto` package. Preserve all behavior; verify via tests.

**Files:**
- Create: `utils/wxcrypto/doc.go`
- Create: `utils/wxcrypto/msgcrypt.go`
- Create: `utils/wxcrypto/msgcrypt_test.go`

- [ ] **Step 1.1: Create package doc**

Create `utils/wxcrypto/doc.go`:

```go
// Package wxcrypto 实现微信公众号/开放平台"消息加解密"(Biz Msg Crypt)。
//
// 参考文档:
//
//	https://developers.weixin.qq.com/doc/offiaccount/Message_Management/Message_encryption_and_decryption.html
//
// 算法流程:
//  1. 校验 msg_signature = sha1(sort([token, timestamp, nonce, encrypted])).
//  2. 解密 encrypted: base64 -> AES-256-CBC(iv=aesKey[:16]) -> PKCS#7 unpad
//     -> 16 字节随机前缀 + 4 字节网络序长度 + 明文 + 发送方 appid.
//  3. 加密反之。
//
// 本包被 offiaccount 和 oplatform 共同引用。offiaccount 保留了
// 原有 MsgCrypto/ParseNotify 导出符号的薄别名，不破坏外部调用点。
package wxcrypto
```

- [ ] **Step 1.2: Write the failing test file**

Create `utils/wxcrypto/msgcrypt_test.go`:

```go
package wxcrypto

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"io"
	"sort"
	"strings"
	"testing"
)

func genEncodingAESKey(t *testing.T) string {
	t.Helper()
	raw := make([]byte, 32)
	for i := range raw {
		raw[i] = byte(i)
	}
	k := base64.StdEncoding.EncodeToString(raw)
	return strings.TrimSuffix(k, "=")
}

func TestMsgCrypto_EncryptDecryptRoundtrip(t *testing.T) {
	key := genEncodingAESKey(t)
	mc, err := New("tk", key, "wxappid")
	if err != nil {
		t.Fatal(err)
	}
	plain := []byte(`<xml><Content><![CDATA[hello]]></Content></xml>`)
	cipher, err := mc.Encrypt(plain)
	if err != nil {
		t.Fatal(err)
	}
	got, fromAppid, err := mc.Decrypt(cipher)
	if err != nil {
		t.Fatal(err)
	}
	if fromAppid != "wxappid" {
		t.Errorf("fromAppid: got %q", fromAppid)
	}
	if string(got) != string(plain) {
		t.Errorf("roundtrip mismatch: %q vs %q", got, plain)
	}
}

func TestMsgCrypto_SignatureDeterministic(t *testing.T) {
	mc, _ := New("tk", genEncodingAESKey(t), "wxappid")
	sig1 := mc.Signature("1700000000", "nonceA", "encX")
	sig2 := mc.Signature("1700000000", "nonceA", "encX")
	if sig1 != sig2 || sig1 == "" {
		t.Errorf("signature not deterministic: %s vs %s", sig1, sig2)
	}
	if !mc.VerifySignature(sig1, "1700000000", "nonceA", "encX") {
		t.Error("expected VerifySignature to succeed")
	}
	if mc.VerifySignature("bogus", "1700000000", "nonceA", "encX") {
		t.Error("expected VerifySignature to fail for bad sig")
	}
}

func TestMsgCrypto_BuildEncryptedReply(t *testing.T) {
	mc, _ := New("tk", genEncodingAESKey(t), "wxappid")
	reply, err := mc.BuildEncryptedReply([]byte(`<xml><MsgType>text</MsgType></xml>`), "1700000001", "nonce1")
	if err != nil {
		t.Fatal(err)
	}
	var env struct {
		Encrypt      string `xml:"Encrypt"`
		MsgSignature string `xml:"MsgSignature"`
		Nonce        string `xml:"Nonce"`
	}
	if err := xml.Unmarshal(reply, &env); err != nil {
		t.Fatal(err)
	}
	if !mc.VerifySignature(env.MsgSignature, "1700000001", "nonce1", env.Encrypt) {
		t.Error("reply signature should verify")
	}
	plain, _, err := mc.Decrypt(env.Encrypt)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(plain), "text") {
		t.Errorf("unexpected decrypted: %s", plain)
	}
}

func TestVerifyServerToken(t *testing.T) {
	sig := simpleSHA1("tk", "1700", "n1")
	if err := VerifyServerToken("tk", sig, "1700", "n1"); err != nil {
		t.Errorf("expected ok, got %v", err)
	}
	if err := VerifyServerToken("tk", "wrong", "1700", "n1"); err == nil {
		t.Error("expected error")
	}
}

func TestMsgCrypto_Decrypt_AppidMismatch(t *testing.T) {
	mcA, _ := New("tk", genEncodingAESKey(t), "appA")
	mcB, _ := New("tk", genEncodingAESKey(t), "appB")
	cipher, err := mcA.Encrypt([]byte(`<xml/>`))
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := mcB.Decrypt(cipher); err == nil {
		t.Error("expected appid mismatch error")
	}
}

func simpleSHA1(token, ts, nonce string) string {
	parts := []string{token, ts, nonce}
	sort.Strings(parts)
	h := sha1.New()
	_, _ = io.WriteString(h, strings.Join(parts, ""))
	return hex.EncodeToString(h.Sum(nil))
}
```

- [ ] **Step 1.3: Run the failing test**

Run: `go test ./utils/wxcrypto/...`
Expected: Build errors — `New`, `MsgCrypto`, `VerifyServerToken` undefined.

- [ ] **Step 1.4: Create `utils/wxcrypto/msgcrypt.go`**

Copy the implementation below verbatim. This is the existing `offiaccount/crypto.go` body plus `VerifyServerToken` pulled over from `offiaccount/notify.go`, with the package name changed and constructor renamed from `NewMsgCrypto` to `New` (idiomatic for constructor of package-matching type).

```go
package wxcrypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
)

// MsgCrypto 持有 token + AESKey + AppId，做签名/加解密。线程安全。
type MsgCrypto struct {
	token  string
	appid  string
	aesKey []byte // base64 解码后的 32 字节 key
	iv     []byte // AES-CBC IV，微信规定使用 aesKey[:16]
}

// New 从配置 Token / EncodingAESKey / AppId 构造 MsgCrypto。
// EncodingAESKey 必须是 43 个字符，解码后得到 32 字节 AES-256 密钥。
func New(token, encodingAESKey, appid string) (*MsgCrypto, error) {
	if token == "" || appid == "" {
		return nil, errors.New("wxcrypto: token and appid are required")
	}
	if len(encodingAESKey) != 43 {
		return nil, fmt.Errorf("wxcrypto: EncodingAESKey must be 43 chars, got %d", len(encodingAESKey))
	}
	key, err := base64.StdEncoding.DecodeString(encodingAESKey + "=")
	if err != nil {
		return nil, fmt.Errorf("wxcrypto: decode EncodingAESKey: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("wxcrypto: decoded AES key must be 32 bytes, got %d", len(key))
	}
	return &MsgCrypto{token: token, appid: appid, aesKey: key, iv: key[:16]}, nil
}

// Token 返回原始 token，便于校验 GET 接入签名。
func (m *MsgCrypto) Token() string { return m.token }

// Signature 按微信规范对 [token, timestamp, nonce, encryptedMsg] 计算 sha1 签名。
func (m *MsgCrypto) Signature(timestamp, nonce, encryptedMsg string) string {
	parts := []string{m.token, timestamp, nonce, encryptedMsg}
	sort.Strings(parts)
	h := sha1.New()
	_, _ = io.WriteString(h, strings.Join(parts, ""))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifySignature 校验 msg_signature。
func (m *MsgCrypto) VerifySignature(msgSignature, timestamp, nonce, encryptedMsg string) bool {
	return SubtleConstEq(m.Signature(timestamp, nonce, encryptedMsg), msgSignature)
}

// SubtleConstEq 常数时间比较两个字符串。
func SubtleConstEq(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var v byte
	for i := 0; i < len(a); i++ {
		v |= a[i] ^ b[i]
	}
	return v == 0
}

// Decrypt 解密 encryptedMsg（base64 字符串），返回明文 XML 字节。
// 同时返回发送方 appid，用于校验来源。
func (m *MsgCrypto) Decrypt(encryptedMsg string) (plaintext []byte, fromAppid string, err error) {
	cipherText, err := base64.StdEncoding.DecodeString(encryptedMsg)
	if err != nil {
		return nil, "", fmt.Errorf("decode encrypted msg: %w", err)
	}
	block, err := aes.NewCipher(m.aesKey)
	if err != nil {
		return nil, "", err
	}
	if len(cipherText)%aes.BlockSize != 0 {
		return nil, "", fmt.Errorf("ciphertext len %d not multiple of %d", len(cipherText), aes.BlockSize)
	}
	buf := make([]byte, len(cipherText))
	mode := cipher.NewCBCDecrypter(block, m.iv)
	mode.CryptBlocks(buf, cipherText)

	buf, err = pkcs7Unpad(buf, aes.BlockSize)
	if err != nil {
		return nil, "", err
	}
	if len(buf) < 20 {
		return nil, "", errors.New("decrypted payload too short")
	}
	buf = buf[16:]
	msgLen := binary.BigEndian.Uint32(buf[:4])
	if int(msgLen)+4 > len(buf) {
		return nil, "", fmt.Errorf("msgLen %d exceeds buffer %d", msgLen, len(buf)-4)
	}
	msg := buf[4 : 4+msgLen]
	fromAppid = string(buf[4+msgLen:])
	if m.appid != "" && fromAppid != m.appid {
		return nil, fromAppid, fmt.Errorf("appid mismatch: got %q want %q", fromAppid, m.appid)
	}
	return msg, fromAppid, nil
}

// Encrypt 加密明文 XML 字节，返回 base64 密文。
func (m *MsgCrypto) Encrypt(plaintext []byte) (string, error) {
	prefix := make([]byte, 16)
	if _, err := rand.Read(prefix); err != nil {
		return "", err
	}
	var lenBuf [4]byte
	binary.BigEndian.PutUint32(lenBuf[:], uint32(len(plaintext)))

	var body bytes.Buffer
	body.Write(prefix)
	body.Write(lenBuf[:])
	body.Write(plaintext)
	body.WriteString(m.appid)

	padded := pkcs7Pad(body.Bytes(), aes.BlockSize)
	block, err := aes.NewCipher(m.aesKey)
	if err != nil {
		return "", err
	}
	ciphertext := make([]byte, len(padded))
	mode := cipher.NewCBCEncrypter(block, m.iv)
	mode.CryptBlocks(ciphertext, padded)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// BuildEncryptedReply 构造微信要求的加密响应 XML。
func (m *MsgCrypto) BuildEncryptedReply(plaintext []byte, timestamp, nonce string) ([]byte, error) {
	encrypted, err := m.Encrypt(plaintext)
	if err != nil {
		return nil, err
	}
	signature := m.Signature(timestamp, nonce, encrypted)
	return xml.Marshal(&encryptedReply{
		Encrypt:      cdata{encrypted},
		MsgSignature: cdata{signature},
		TimeStamp:    timestamp,
		Nonce:        cdata{nonce},
	})
}

// VerifyServerToken 微信接入验证：校验 GET 请求 signature = sha1(sort([token,timestamp,nonce])).
func VerifyServerToken(token, signature, timestamp, nonce string) error {
	parts := []string{token, timestamp, nonce}
	sort.Strings(parts)
	h := sha1.New()
	_, _ = io.WriteString(h, strings.Join(parts, ""))
	got := hex.EncodeToString(h.Sum(nil))
	if !SubtleConstEq(got, signature) {
		return fmt.Errorf("wxcrypto: bad signature")
	}
	return nil
}

type cdata struct {
	Value string `xml:",cdata"`
}

type encryptedReply struct {
	XMLName      xml.Name `xml:"xml"`
	Encrypt      cdata    `xml:"Encrypt"`
	MsgSignature cdata    `xml:"MsgSignature"`
	TimeStamp    string   `xml:"TimeStamp"`
	Nonce        cdata    `xml:"Nonce"`
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	pad := blockSize - len(data)%blockSize
	padding := bytes.Repeat([]byte{byte(pad)}, pad)
	return append(data, padding...)
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 || len(data)%blockSize != 0 {
		return nil, errors.New("invalid padded data length")
	}
	pad := int(data[len(data)-1])
	if pad == 0 || pad > blockSize {
		return nil, fmt.Errorf("invalid pad byte %d", pad)
	}
	for i := len(data) - pad; i < len(data); i++ {
		if int(data[i]) != pad {
			return nil, errors.New("invalid pkcs7 padding")
		}
	}
	return data[:len(data)-pad], nil
}
```

- [ ] **Step 1.5: Run tests to verify they pass**

Run: `go test ./utils/wxcrypto/...`
Expected: PASS for all 5 tests.

- [ ] **Step 1.6: Commit**

```bash
git add utils/wxcrypto/
git commit -m "feat(wxcrypto): extract Biz Msg Crypt into shared utils/wxcrypto package

Moves the AES-256-CBC encryption, PKCS#7 padding, SHA1 signature,
and encrypted-reply envelope logic out of offiaccount so oplatform
can reuse it without duplication. Constructor renamed New (was
NewMsgCrypto). offiaccount shim added in a follow-up commit.

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

## Task 2: offiaccount crypto shim — preserve external API

**Goal:** Replace `offiaccount/crypto.go` with a thin alias layer that re-exports `wxcrypto` symbols under their old names. Preserve every existing external call site.

**Files:**
- Modify: `offiaccount/crypto.go`
- Modify: `offiaccount/notify.go:88-93` (remove `tokenFromCrypto`; use `MsgCrypto.Token()`)
- Modify: `offiaccount/crypto_test.go` (keep as smoke test; delete the tests that duplicate wxcrypto coverage — retain only the integration-level `TestParseNotify_*` and `TestVerifyServerToken` which exercise the shim path)

- [ ] **Step 2.1: Rewrite `offiaccount/crypto.go` as a shim**

Replace the entire contents of `offiaccount/crypto.go` with:

```go
package offiaccount

import "github.com/godrealms/go-wechat-sdk/utils/wxcrypto"

// MsgCrypto is an alias for wxcrypto.MsgCrypto.
//
// 本类型是微信消息加解密的主入口。原实现已迁移到 utils/wxcrypto；
// 为了不破坏已经在生产使用的 offiaccount.NewMsgCrypto/*MsgCrypto 调用点，
// 这里保留类型别名和构造器转发。
type MsgCrypto = wxcrypto.MsgCrypto

// NewMsgCrypto 构造一个消息加解密器。等价于 wxcrypto.New。
func NewMsgCrypto(token, encodingAESKey, appid string) (*MsgCrypto, error) {
	return wxcrypto.New(token, encodingAESKey, appid)
}

// subtleConstEq 历史符号，保留以兼容同包内引用。新代码请用 wxcrypto.SubtleConstEq。
func subtleConstEq(a, b string) bool {
	return wxcrypto.SubtleConstEq(a, b)
}
```

- [ ] **Step 2.2: Update `offiaccount/notify.go` to use the shim**

The current `tokenFromCrypto(c *MsgCrypto)` reaches into the unexported `token` field. After the shim, the field is in another package and no longer accessible by name. Replace the helper with `MsgCrypto.Token()`.

Edit `offiaccount/notify.go` — replace lines 88-93 (`func tokenFromCrypto(...)` and its call site). The full updated file:

```go
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
```

- [ ] **Step 2.3: Slim down `offiaccount/crypto_test.go` to a shim smoke test**

Replace the entire contents of `offiaccount/crypto_test.go` with:

```go
package offiaccount

import (
	"encoding/base64"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// These tests ensure the offiaccount shim still exposes the same
// behavior through type aliasing. The full Biz Msg Crypt test suite
// lives in utils/wxcrypto.

func shimKey() string {
	raw := make([]byte, 32)
	for i := range raw {
		raw[i] = byte(i)
	}
	return strings.TrimSuffix(base64.StdEncoding.EncodeToString(raw), "=")
}

func TestOffiaccountMsgCryptoShim_Roundtrip(t *testing.T) {
	mc, err := NewMsgCrypto("tk", shimKey(), "wxappid")
	if err != nil {
		t.Fatal(err)
	}
	plain := []byte(`<xml><Content><![CDATA[hi]]></Content></xml>`)
	enc, err := mc.Encrypt(plain)
	if err != nil {
		t.Fatal(err)
	}
	got, _, err := mc.Decrypt(enc)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(plain) {
		t.Errorf("roundtrip mismatch")
	}
}

func TestParseNotify_EncryptedMode(t *testing.T) {
	mc, _ := NewMsgCrypto("tk", shimKey(), "wxappid")
	plain := []byte(`<xml><MsgType><![CDATA[text]]></MsgType><Content><![CDATA[hi]]></Content></xml>`)
	encrypted, err := mc.Encrypt(plain)
	if err != nil {
		t.Fatal(err)
	}
	sig := mc.Signature("1700", "n1", encrypted)
	bodyXML := `<xml><ToUserName>gh_x</ToUserName><Encrypt><![CDATA[` + encrypted + `]]></Encrypt></xml>`
	req := httptest.NewRequest(http.MethodPost,
		"/notify?msg_signature="+sig+"&timestamp=1700&nonce=n1",
		strings.NewReader(bodyXML))
	got, err := ParseNotify(req, mc)
	if err != nil {
		t.Fatal(err)
	}
	var env struct {
		MsgType string `xml:"MsgType"`
		Content string `xml:"Content"`
	}
	if err := xml.Unmarshal(got, &env); err != nil {
		t.Fatal(err)
	}
	if env.MsgType != "text" || env.Content != "hi" {
		t.Errorf("unexpected decrypted: %+v", env)
	}
}

func TestParseNotify_BadSignature(t *testing.T) {
	mc, _ := NewMsgCrypto("tk", shimKey(), "wxappid")
	encrypted, _ := mc.Encrypt([]byte(`<xml/>`))
	bodyXML := `<xml><Encrypt><![CDATA[` + encrypted + `]]></Encrypt></xml>`
	req := httptest.NewRequest(http.MethodPost,
		"/notify?msg_signature=deadbeef&timestamp=1700&nonce=n1",
		strings.NewReader(bodyXML))
	if _, err := ParseNotify(req, mc); err == nil {
		t.Error("expected signature error")
	}
}

func TestVerifyServerTokenShim(t *testing.T) {
	// Validate the shim delegation path (offiaccount.VerifyServerToken
	// forwards to wxcrypto.VerifyServerToken). Algorithm correctness
	// is tested in utils/wxcrypto.
	if err := VerifyServerToken("tk", "deadbeef", "1700", "n1"); err == nil {
		t.Error("expected error for bad signature")
	}
}
```

- [ ] **Step 2.4: Run full test suite**

Run: `go test ./offiaccount/... ./utils/wxcrypto/...`
Expected: PASS. If `offiaccount/notify.go` previously used `tokenFromCrypto`, the old reference is now gone — make sure no other file in `offiaccount/` still references it (`grep -n tokenFromCrypto offiaccount/*.go`).

- [ ] **Step 2.5: Run full repo build**

Run: `go build ./...`
Expected: clean build.

- [ ] **Step 2.6: Commit**

```bash
git add offiaccount/crypto.go offiaccount/crypto_test.go offiaccount/notify.go
git commit -m "refactor(offiaccount): alias crypto types to utils/wxcrypto

offiaccount.MsgCrypto is now a type alias for wxcrypto.MsgCrypto; the
shim preserves the public NewMsgCrypto / VerifyServerToken / ParseNotify
API unchanged. crypto_test.go is slimmed to a smoke test — the full
Biz Msg Crypt coverage lives in utils/wxcrypto.

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

## Task 3: offiaccount TokenSource injection

**Goal:** Add `offiaccount.TokenSource` interface and `WithTokenSource` option. When injected, `AccessTokenE` delegates to the source instead of calling `/cgi-bin/token`. When absent, behavior is unchanged.

**Files:**
- Create: `offiaccount/tokensource.go`
- Modify: `offiaccount/client.go` (add field, add Option, add variadic to `NewClient`, route `AccessTokenE`)
- Modify: `offiaccount/client_test.go` (add injection test)

- [ ] **Step 3.1: Write the failing test**

Append to `offiaccount/client_test.go`:

```go
type fakeTokenSource struct {
	token string
	err   error
	calls int
}

func (f *fakeTokenSource) AccessToken(ctx context.Context) (string, error) {
	f.calls++
	return f.token, f.err
}

func TestClient_AccessTokenE_UsesInjectedTokenSource(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("/cgi-bin/token must NOT be called when TokenSource is injected: %s", r.URL.Path)
	}))
	defer srv.Close()

	fake := &fakeTokenSource{token: "INJECTED"}
	c := NewClient(context.Background(), &Config{AppId: "appid"},
		WithHTTPClient(utils.NewHTTP(srv.URL, utils.WithTimeout(time.Second*3))),
		WithTokenSource(fake),
	)
	tok, err := c.AccessTokenE(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if tok != "INJECTED" {
		t.Errorf("got %q, want INJECTED", tok)
	}
	if fake.calls != 1 {
		t.Errorf("expected 1 call, got %d", fake.calls)
	}
}
```

Note: `WithHTTPClient` does not yet exist. The test uses a placeholder; you'll add `WithHTTPClient` as an `Option` in Step 3.4.

- [ ] **Step 3.2: Run the failing test**

Run: `go test ./offiaccount/ -run TestClient_AccessTokenE_UsesInjectedTokenSource`
Expected: build error — `WithTokenSource`, `WithHTTPClient`, `Option` undefined.

- [ ] **Step 3.3: Create `offiaccount/tokensource.go`**

```go
package offiaccount

import "context"

// TokenSource 是 access_token 的可注入来源。
// 当 Client 配置了 TokenSource 时，AccessTokenE 会直接委托给它，
// 不再调用 /cgi-bin/token。典型场景：开放平台代调用
// (oplatform.AuthorizerClient 实现本接口)。
type TokenSource interface {
	AccessToken(ctx context.Context) (string, error)
}
```

- [ ] **Step 3.4: Extend `offiaccount/client.go`**

Open `offiaccount/client.go` and make the following edits in place:

1. Add `tokenSource TokenSource` field to `Client` struct (after `expiresAt`).
2. Introduce an `Option` type and expose `WithTokenSource` and `WithHTTPClient` options.
3. Change `NewClient` signature from `NewClient(ctx context.Context, config *Config) *Client` to `NewClient(ctx context.Context, config *Config, opts ...Option) *Client`. Existing call sites continue to compile because variadic is zero-or-more.
4. Update `AccessTokenE` to delegate to `tokenSource` if set.

Full updated `client.go`:

```go
package offiaccount

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

type Config struct {
	AppId          string `json:"appId"`
	AppSecret      string `json:"appSecret"`
	Token          string `json:"token"`
	EncodingAESKey string `json:"encodingAESKey"`
}

// Client 微信公众号
type Client struct {
	ctx    context.Context
	Config *Config
	Https  *utils.HTTP

	tokenMutex  sync.RWMutex
	accessToken string
	expiresAt   time.Time

	tokenSource TokenSource
}

// Option 构造可选项。
type Option func(*Client)

// WithTokenSource 注入外部 access_token 来源（例如开放平台代调用）。
// 设置后 AccessTokenE 不再调用 /cgi-bin/token。
func WithTokenSource(ts TokenSource) Option {
	return func(c *Client) { c.tokenSource = ts }
}

// WithHTTPClient 允许替换底层 HTTP 客户端（主要用于测试注入）。
func WithHTTPClient(h *utils.HTTP) Option {
	return func(c *Client) {
		if h != nil {
			c.Https = h
		}
	}
}

// NewClient 创建客户端
func NewClient(ctx context.Context, config *Config, opts ...Option) *Client {
	if ctx == nil {
		ctx = context.Background()
	}
	c := &Client{
		ctx:    ctx,
		Config: config,
		Https:  utils.NewHTTP("https://api.weixin.qq.com", utils.WithTimeout(time.Second*30)),
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// GetAccessToken 旧版兼容入口：返回字符串，错误被吞掉。新代码请使用 AccessTokenE。
func (c *Client) GetAccessToken() string {
	token, _ := c.AccessTokenE(c.ctx)
	return token
}

// AccessTokenE 显式获取 access_token，错误会被传递给调用方。
// 如果注入了 TokenSource，则委托给它；否则走自有 /cgi-bin/token 流程。
func (c *Client) AccessTokenE(ctx context.Context) (string, error) {
	if ctx == nil {
		ctx = c.ctx
	}
	if c.tokenSource != nil {
		return c.tokenSource.AccessToken(ctx)
	}

	c.tokenMutex.RLock()
	if c.accessToken != "" && time.Now().Before(c.expiresAt) {
		token := c.accessToken
		c.tokenMutex.RUnlock()
		return token, nil
	}
	c.tokenMutex.RUnlock()

	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()

	if c.accessToken != "" && time.Now().Before(c.expiresAt) {
		return c.accessToken, nil
	}

	token, err := c.refreshAccessToken(ctx)
	if err != nil {
		return "", err
	}
	c.accessToken = token.AccessToken
	c.expiresAt = time.Now().Add(time.Duration(token.ExpiresIn-60) * time.Second)
	return c.accessToken, nil
}

// getAccessToken 旧版内部入口（保留以兼容现有代码）。
//
// Deprecated: 使用 AccessTokenE 来正确处理错误。
func (c *Client) getAccessToken() string {
	token, _ := c.AccessTokenE(c.ctx)
	return token
}

// refreshAccessToken 调用微信服务器刷新 access_token。
func (c *Client) refreshAccessToken(ctx context.Context) (*AccessToken, error) {
	if c.Config == nil || c.Config.AppId == "" || c.Config.AppSecret == "" {
		return nil, fmt.Errorf("offiaccount: AppId and AppSecret are required")
	}
	query := url.Values{
		"grant_type": {"client_credential"},
		"appid":      {c.Config.AppId},
		"secret":     {c.Config.AppSecret},
	}
	result := &AccessToken{}
	if err := c.Https.Get(ctx, "/cgi-bin/token", query, result); err != nil {
		return nil, fmt.Errorf("offiaccount: fetch access_token failed: %w", err)
	}
	if result.ErrCode != 0 {
		return nil, &WeixinError{ErrCode: result.ErrCode, ErrMsg: result.ErrMsg}
	}
	if result.AccessToken == "" {
		return nil, fmt.Errorf("offiaccount: empty access_token returned")
	}
	return result, nil
}
```

- [ ] **Step 3.5: Reconcile `client_test.go`'s `newClientWithBaseURL` helper**

The existing helper mutates `c.Https` after construction. Keep it as-is — still valid since the field is exported. Just ensure imports include `"github.com/godrealms/go-wechat-sdk/utils"`.

- [ ] **Step 3.6: Run tests**

Run: `go test ./offiaccount/...`
Expected: all tests PASS — including the new `TestClient_AccessTokenE_UsesInjectedTokenSource` and all pre-existing ones.

- [ ] **Step 3.7: Run full repo build**

Run: `go build ./...`
Expected: clean.

- [ ] **Step 3.8: Commit**

```bash
git add offiaccount/tokensource.go offiaccount/client.go offiaccount/client_test.go
git commit -m "feat(offiaccount): add TokenSource injection for authorizer delegation

Adds Option / WithTokenSource / WithHTTPClient and routes AccessTokenE
through the injected source when present. Default behavior (calling
/cgi-bin/token with AppSecret) is unchanged. NewClient signature
extended with variadic Option args — existing call sites remain
source-compatible.

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

## Task 4: mini-program TokenSource injection

**Goal:** Same shape as Task 3, but for `mini-program`.

**Files:**
- Create: `mini-program/tokensource.go`
- Modify: `mini-program/client.go` (add field, add option, route `AccessToken`)
- Modify: `mini-program/client_test.go` (add injection test)

- [ ] **Step 4.1: Write failing test**

Append to `mini-program/client_test.go`:

```go
type fakeTokenSource struct {
	token string
	err   error
	calls int
}

func (f *fakeTokenSource) AccessToken(ctx context.Context) (string, error) {
	f.calls++
	return f.token, f.err
}

func TestClient_AccessToken_UsesInjectedTokenSource(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("/cgi-bin/token must NOT be called when TokenSource is injected: %s", r.URL.Path)
	}))
	defer srv.Close()

	fake := &fakeTokenSource{token: "INJECTED_MP"}
	c, err := NewClient(Config{AppId: "wx", AppSecret: "sec"},
		WithHTTP(utils.NewHTTP(srv.URL, utils.WithTimeout(time.Second*3))),
		WithTokenSource(fake),
	)
	if err != nil {
		t.Fatal(err)
	}
	tok, err := c.AccessToken(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if tok != "INJECTED_MP" {
		t.Errorf("got %q, want INJECTED_MP", tok)
	}
	if fake.calls != 1 {
		t.Errorf("expected 1 call, got %d", fake.calls)
	}
}
```

- [ ] **Step 4.2: Run failing test**

Run: `go test ./mini-program/ -run TestClient_AccessToken_UsesInjectedTokenSource`
Expected: build error — `WithTokenSource` undefined.

- [ ] **Step 4.3: Create `mini-program/tokensource.go`**

```go
package mini_program

import "context"

// TokenSource 是 access_token 的可注入来源。
// 当 Client 配置了 TokenSource 时，AccessToken() 会直接委托给它，
// 不再调用 /cgi-bin/token。典型场景：开放平台代调用
// (oplatform.AuthorizerClient 实现本接口)。
type TokenSource interface {
	AccessToken(ctx context.Context) (string, error)
}
```

- [ ] **Step 4.4: Extend `mini-program/client.go`**

In `mini-program/client.go`:

1. Add `tokenSource TokenSource` field to `Client` (after `expiresAt`).
2. Add `WithTokenSource` Option.
3. In `AccessToken(ctx)`, delegate to `tokenSource` if set — return `c.tokenSource.AccessToken(ctx)` as the first branch.
4. Allow `AppSecret == ""` when `tokenSource != nil` — loosen the check in `NewClient`.

Edits:

Add after line 58 (end of `WithHTTP`):

```go
// WithTokenSource 注入外部 access_token 来源（例如开放平台代调用）。
func WithTokenSource(ts TokenSource) Option {
	return func(c *Client) { c.tokenSource = ts }
}
```

Replace the struct block (lines 29-38) to add the new field:

```go
// Client 小程序服务端客户端。并发安全。
type Client struct {
	cfg  Config
	http *utils.HTTP

	mu          sync.RWMutex
	accessToken string
	expiresAt   time.Time

	tokenSource TokenSource
}
```

Replace the `NewClient` validation (lines 47-49) to:

```go
	if cfg.AppId == "" {
		return nil, fmt.Errorf("mini_program: AppId is required")
	}
	// AppSecret is only required when no TokenSource is injected.
```

And move the AppSecret check into the token-fetch path so options can relax it. Full updated `NewClient`:

```go
// NewClient 构造客户端。
func NewClient(cfg Config, opts ...Option) (*Client, error) {
	if cfg.AppId == "" {
		return nil, fmt.Errorf("mini_program: AppId is required")
	}
	c := &Client{
		cfg:  cfg,
		http: utils.NewHTTP("https://api.weixin.qq.com", utils.WithTimeout(time.Second*30)),
	}
	for _, o := range opts {
		o(c)
	}
	if c.tokenSource == nil && cfg.AppSecret == "" {
		return nil, fmt.Errorf("mini_program: AppSecret is required when no TokenSource is injected")
	}
	return c, nil
}
```

Replace `AccessToken` (lines ~101-132) — insert the TokenSource branch at the top:

```go
// AccessToken 获取全局 access_token（带进程内缓存，提前 60 秒过期）。
// 注入 TokenSource 时直接委托。
func (c *Client) AccessToken(ctx context.Context) (string, error) {
	if c.tokenSource != nil {
		return c.tokenSource.AccessToken(ctx)
	}
	c.mu.RLock()
	if c.accessToken != "" && time.Now().Before(c.expiresAt) {
		t := c.accessToken
		c.mu.RUnlock()
		return t, nil
	}
	c.mu.RUnlock()
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.accessToken != "" && time.Now().Before(c.expiresAt) {
		return c.accessToken, nil
	}
	q := url.Values{
		"grant_type": {"client_credential"},
		"appid":      {c.cfg.AppId},
		"secret":     {c.cfg.AppSecret},
	}
	out := &accessTokenResp{}
	if err := c.http.Get(ctx, "/cgi-bin/token", q, out); err != nil {
		return "", fmt.Errorf("mini_program: fetch token: %w", err)
	}
	if out.ErrCode != 0 {
		return "", fmt.Errorf("mini_program: token errcode=%d errmsg=%s", out.ErrCode, out.ErrMsg)
	}
	if out.AccessToken == "" {
		return "", fmt.Errorf("mini_program: empty access_token")
	}
	c.accessToken = out.AccessToken
	c.expiresAt = time.Now().Add(time.Duration(out.ExpiresIn-60) * time.Second)
	return c.accessToken, nil
}
```

- [ ] **Step 4.5: Run tests**

Run: `go test ./mini-program/...`
Expected: PASS for all tests including the new injection test and the preexisting `TestClient_AccessTokenCaches` etc.

- [ ] **Step 4.6: Full repo build**

Run: `go build ./...`
Expected: clean.

- [ ] **Step 4.7: Commit**

```bash
git add mini-program/tokensource.go mini-program/client.go mini-program/client_test.go
git commit -m "feat(mini-program): add TokenSource injection for authorizer delegation

Mirrors the offiaccount change: WithTokenSource option + delegation
in AccessToken(). AppSecret validation moves from NewClient-time
to 'when no TokenSource is injected', so oplatform can construct a
Client with only AppId.

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

## Task 5: oplatform package skeleton — errors & struct types

**Goal:** Create the `oplatform` package, define error sentinels and the core DTOs used by later tasks. No client methods yet.

**Files:**
- Remove old file: `oplatform/client.go` (currently a 2-line placeholder — replaced in Task 6)
- Create: `oplatform/errors.go`
- Create: `oplatform/struct.component.go`
- Create: `oplatform/struct.authorizer.go`
- Create: `oplatform/struct.qrlogin.go`

- [ ] **Step 5.1: Create `oplatform/errors.go`**

```go
package oplatform

import (
	"errors"
	"fmt"
)

// WeixinError 微信业务错误 (errcode != 0).
type WeixinError struct {
	ErrCode int
	ErrMsg  string
}

func (e *WeixinError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("oplatform: errcode=%d errmsg=%s", e.ErrCode, e.ErrMsg)
}

// 常见哨兵错误。
var (
	// ErrNotFound 由 Store 实现返回，表示 key 不存在（非 I/O 错误）。
	ErrNotFound = errors.New("oplatform: not found")

	// ErrAuthorizerRevoked 当 refresh_token 失效 (errcode=61023) 时返回；
	// 调用方应删除 Store 中该 authorizer 的记录并重新引导授权。
	ErrAuthorizerRevoked = errors.New("oplatform: authorizer refresh_token revoked")

	// ErrVerifyTicketMissing 当 component_access_token 需要刷新但 Store
	// 尚未收到微信推送的 component_verify_ticket 时返回。
	ErrVerifyTicketMissing = errors.New("oplatform: component_verify_ticket not yet received")
)
```

- [ ] **Step 5.2: Create `oplatform/struct.component.go`**

```go
package oplatform

// component_access_token 响应
type componentTokenResp struct {
	ComponentAccessToken string `json:"component_access_token"`
	ExpiresIn            int64  `json:"expires_in"`
	ErrCode              int    `json:"errcode,omitempty"`
	ErrMsg               string `json:"errmsg,omitempty"`
}

// pre_auth_code 响应
type preAuthCodeResp struct {
	PreAuthCode string `json:"pre_auth_code"`
	ExpiresIn   int64  `json:"expires_in"`
	ErrCode     int    `json:"errcode,omitempty"`
	ErrMsg      string `json:"errmsg,omitempty"`
}

// ComponentNotify 回调解析结果。
type ComponentNotify struct {
	AppID      string // ComponentAppID
	CreateTime int64
	InfoType   string // component_verify_ticket / authorized / updateauthorized / unauthorized

	// component_verify_ticket 时填
	ComponentVerifyTicket string

	// authorized / updateauthorized / unauthorized 时填
	AuthorizerAppID              string
	AuthorizationCode            string
	AuthorizationCodeExpiredTime int64
	PreAuthCode                  string

	// 原始明文 XML
	Raw []byte
}
```

- [ ] **Step 5.3: Create `oplatform/struct.authorizer.go`**

```go
package oplatform

// AuthorizationInfo 对应 /cgi-bin/component/api_query_auth 的 authorization_info。
type AuthorizationInfo struct {
	AuthorizerAppID        string     `json:"authorizer_appid"`
	AuthorizerAccessToken  string     `json:"authorizer_access_token"`
	ExpiresIn              int64      `json:"expires_in"`
	AuthorizerRefreshToken string     `json:"authorizer_refresh_token"`
	FuncInfo               []FuncInfo `json:"func_info"`
}

type FuncInfo struct {
	FuncscopeCategory FuncscopeCategory `json:"funcscope_category"`
}

type FuncscopeCategory struct {
	ID int `json:"id"`
}

type queryAuthResp struct {
	AuthorizationInfo AuthorizationInfo `json:"authorization_info"`
	ErrCode           int               `json:"errcode,omitempty"`
	ErrMsg            string            `json:"errmsg,omitempty"`
}

// authorizer_access_token 刷新响应
type authorizerTokenResp struct {
	AuthorizerAccessToken  string `json:"authorizer_access_token"`
	AuthorizerRefreshToken string `json:"authorizer_refresh_token"`
	ExpiresIn              int64  `json:"expires_in"`
	ErrCode                int    `json:"errcode,omitempty"`
	ErrMsg                 string `json:"errmsg,omitempty"`
}

// AuthorizerInfo / Option 查询结构体
type AuthorizerInfo struct {
	NickName        string `json:"nick_name"`
	HeadImg         string `json:"head_img"`
	ServiceTypeInfo struct {
		ID int `json:"id"`
	} `json:"service_type_info"`
	VerifyTypeInfo struct {
		ID int `json:"id"`
	} `json:"verify_type_info"`
	UserName      string   `json:"user_name"`
	PrincipalName string   `json:"principal_name"`
	Alias         string   `json:"alias"`
	BusinessInfo  any      `json:"business_info,omitempty"`
	QrcodeURL     string   `json:"qrcode_url"`
	Signature     string   `json:"signature"`
	MiniProgramInfo any    `json:"MiniProgramInfo,omitempty"`
}

type AuthorizerInfoResp struct {
	AuthorizerInfo    AuthorizerInfo    `json:"authorizer_info"`
	AuthorizationInfo AuthorizationInfo `json:"authorization_info"`
	ErrCode           int               `json:"errcode,omitempty"`
	ErrMsg            string            `json:"errmsg,omitempty"`
}

type AuthorizerOption struct {
	AuthorizerAppID string `json:"authorizer_appid"`
	OptionName      string `json:"option_name"`
	OptionValue     string `json:"option_value"`
	ErrCode         int    `json:"errcode,omitempty"`
	ErrMsg          string `json:"errmsg,omitempty"`
}

type AuthorizerList struct {
	TotalCount int `json:"total_count"`
	List       []struct {
		AuthorizerAppID string `json:"authorizer_appid"`
		RefreshToken    string `json:"refresh_token"`
		AuthTime        int64  `json:"auth_time"`
	} `json:"list"`
	ErrCode int    `json:"errcode,omitempty"`
	ErrMsg  string `json:"errmsg,omitempty"`
}
```

- [ ] **Step 5.4: Create `oplatform/struct.qrlogin.go`**

```go
package oplatform

// QRLoginToken /sns/oauth2/access_token 响应。
type QRLoginToken struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	Scope        string `json:"scope"`
	UnionID      string `json:"unionid,omitempty"`
	ErrCode      int    `json:"errcode,omitempty"`
	ErrMsg       string `json:"errmsg,omitempty"`
}

// QRLoginUserInfo /sns/userinfo 响应。
type QRLoginUserInfo struct {
	OpenID     string   `json:"openid"`
	Nickname   string   `json:"nickname"`
	Sex        int      `json:"sex"`
	Province   string   `json:"province"`
	City       string   `json:"city"`
	Country    string   `json:"country"`
	HeadImgURL string   `json:"headimgurl"`
	Privilege  []string `json:"privilege"`
	UnionID    string   `json:"unionid,omitempty"`
	ErrCode    int      `json:"errcode,omitempty"`
	ErrMsg     string   `json:"errmsg,omitempty"`
}

type qrloginAuthResp struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}
```

- [ ] **Step 5.5: Delete the old placeholder `oplatform/client.go`**

Run:
```bash
rm oplatform/client.go
```

A new `client.go` is created in Task 6.

- [ ] **Step 5.6: Build check**

Run: `go build ./oplatform/...`
Expected: error — `package oplatform` has no files that define `Client`, but the struct files alone should compile. If you see "no Go files", the `rm` left nothing to build; that's fine — `go build` on an empty directory is a no-op. Moving to a file-based check:

Run: `go vet ./oplatform/...`
Expected: clean (or "no Go files" if nothing imports it yet — both acceptable).

- [ ] **Step 5.7: Commit**

```bash
git add oplatform/errors.go oplatform/struct.component.go oplatform/struct.authorizer.go oplatform/struct.qrlogin.go oplatform/client.go
git commit -m "feat(oplatform): add package skeleton with errors and DTOs

Introduces errors.go (WeixinError + ErrNotFound / ErrAuthorizerRevoked /
ErrVerifyTicketMissing) and the DTOs used by later tasks: component
token / pre_auth_code / authorization_info / authorizer_info /
authorizer_list / QR login responses. Removes the old placeholder
client.go.

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

## Task 6: oplatform Store interface + MemoryStore

**Goal:** Define `Store` contract and ship a thread-safe `MemoryStore`. This task also establishes the `AuthorizerTokens` value type used everywhere downstream.

**Files:**
- Create: `oplatform/store.go`
- Create: `oplatform/store_test.go`

- [ ] **Step 6.1: Write failing test `store_test.go`**

```go
package oplatform

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestMemoryStore_VerifyTicket(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	if _, err := s.GetVerifyTicket(ctx); !errors.Is(err, ErrNotFound) {
		t.Errorf("empty store should return ErrNotFound, got %v", err)
	}
	if err := s.SetVerifyTicket(ctx, "TICKET1"); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetVerifyTicket(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got != "TICKET1" {
		t.Errorf("got %q", got)
	}
}

func TestMemoryStore_ComponentToken(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	if _, _, err := s.GetComponentToken(ctx); !errors.Is(err, ErrNotFound) {
		t.Errorf("empty store should return ErrNotFound, got %v", err)
	}
	exp := time.Now().Add(time.Hour)
	if err := s.SetComponentToken(ctx, "CTOK", exp); err != nil {
		t.Fatal(err)
	}
	tok, gotExp, err := s.GetComponentToken(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if tok != "CTOK" {
		t.Errorf("token mismatch: %q", tok)
	}
	if !gotExp.Equal(exp) {
		t.Errorf("expiry mismatch")
	}
}

func TestMemoryStore_AuthorizerCRUD(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	if _, err := s.GetAuthorizer(ctx, "wxA"); !errors.Is(err, ErrNotFound) {
		t.Errorf("empty store should return ErrNotFound, got %v", err)
	}
	tokens := AuthorizerTokens{
		AccessToken:  "aA",
		RefreshToken: "rA",
		ExpireAt:     time.Now().Add(time.Hour),
	}
	if err := s.SetAuthorizer(ctx, "wxA", tokens); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetAuthorizer(ctx, "wxA")
	if err != nil {
		t.Fatal(err)
	}
	if got.AccessToken != "aA" || got.RefreshToken != "rA" {
		t.Errorf("mismatch: %+v", got)
	}

	// List should contain wxA
	ids, err := s.ListAuthorizerAppIDs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 || ids[0] != "wxA" {
		t.Errorf("list mismatch: %v", ids)
	}

	// Delete
	if err := s.DeleteAuthorizer(ctx, "wxA"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetAuthorizer(ctx, "wxA"); !errors.Is(err, ErrNotFound) {
		t.Errorf("after delete expected ErrNotFound, got %v", err)
	}
}

func TestMemoryStore_ConcurrentAccess(t *testing.T) {
	s := NewMemoryStore()
	ctx := context.Background()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_ = s.SetVerifyTicket(ctx, "t")
			_, _ = s.GetVerifyTicket(ctx)
			_ = s.SetAuthorizer(ctx, "wx", AuthorizerTokens{
				AccessToken: "a", RefreshToken: "r", ExpireAt: time.Now(),
			})
			_, _ = s.GetAuthorizer(ctx, "wx")
			_, _ = s.ListAuthorizerAppIDs(ctx)
		}(i)
	}
	wg.Wait()
}
```

- [ ] **Step 6.2: Run failing test**

Run: `go test ./oplatform/ -run MemoryStore`
Expected: build error — undefined `Store`, `NewMemoryStore`, `AuthorizerTokens`.

- [ ] **Step 6.3: Create `oplatform/store.go`**

```go
package oplatform

import (
	"context"
	"sync"
	"time"
)

// AuthorizerTokens 是单个授权方 (authorizer) 的一组 token。
type AuthorizerTokens struct {
	AccessToken  string
	RefreshToken string
	ExpireAt     time.Time
}

// Store 持久化 component_verify_ticket、component_access_token
// 以及每个 authorizer 的 refresh_token / access_token / expire_at。
//
// SDK 内置 MemoryStore 供测试和本地开发使用；生产应实现 Redis/MySQL
// 版本并通过 WithStore 注入。
//
// Get* 方法在 key 不存在时应返回 ErrNotFound。
type Store interface {
	GetVerifyTicket(ctx context.Context) (string, error)
	SetVerifyTicket(ctx context.Context, ticket string) error

	GetComponentToken(ctx context.Context) (token string, expireAt time.Time, err error)
	SetComponentToken(ctx context.Context, token string, expireAt time.Time) error

	GetAuthorizer(ctx context.Context, appid string) (AuthorizerTokens, error)
	SetAuthorizer(ctx context.Context, appid string, tokens AuthorizerTokens) error
	DeleteAuthorizer(ctx context.Context, appid string) error
	ListAuthorizerAppIDs(ctx context.Context) ([]string, error)
}

// MemoryStore 是 Store 接口的线程安全内存实现。进程重启后所有数据丢失，
// 仅适合测试或本地开发。
type MemoryStore struct {
	mu sync.RWMutex

	verifyTicket string

	componentToken    string
	componentExpireAt time.Time

	authorizers map[string]AuthorizerTokens
}

// NewMemoryStore 构造一个空的内存 Store。
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{authorizers: make(map[string]AuthorizerTokens)}
}

func (m *MemoryStore) GetVerifyTicket(ctx context.Context) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.verifyTicket == "" {
		return "", ErrNotFound
	}
	return m.verifyTicket, nil
}

func (m *MemoryStore) SetVerifyTicket(ctx context.Context, ticket string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.verifyTicket = ticket
	return nil
}

func (m *MemoryStore) GetComponentToken(ctx context.Context) (string, time.Time, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.componentToken == "" {
		return "", time.Time{}, ErrNotFound
	}
	return m.componentToken, m.componentExpireAt, nil
}

func (m *MemoryStore) SetComponentToken(ctx context.Context, token string, expireAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.componentToken = token
	m.componentExpireAt = expireAt
	return nil
}

func (m *MemoryStore) GetAuthorizer(ctx context.Context, appid string) (AuthorizerTokens, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.authorizers[appid]
	if !ok {
		return AuthorizerTokens{}, ErrNotFound
	}
	return t, nil
}

func (m *MemoryStore) SetAuthorizer(ctx context.Context, appid string, tokens AuthorizerTokens) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.authorizers[appid] = tokens
	return nil
}

func (m *MemoryStore) DeleteAuthorizer(ctx context.Context, appid string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.authorizers, appid)
	return nil
}

func (m *MemoryStore) ListAuthorizerAppIDs(ctx context.Context) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]string, 0, len(m.authorizers))
	for k := range m.authorizers {
		out = append(out, k)
	}
	return out, nil
}
```

- [ ] **Step 6.4: Run tests**

Run: `go test ./oplatform/...`
Expected: 4 tests PASS.

Also run with race detector:
Run: `go test -race ./oplatform/...`
Expected: PASS, no race reports.

- [ ] **Step 6.5: Commit**

```bash
git add oplatform/store.go oplatform/store_test.go
git commit -m "feat(oplatform): add Store interface and MemoryStore

Defines the persistence contract for component_verify_ticket,
component_access_token, and per-authorizer refresh/access tokens.
MemoryStore is thread-safe (sync.RWMutex) and suitable for tests
and local development; production should implement Redis/MySQL.

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

## Task 7: oplatform Client + component_access_token lifecycle

**Goal:** Introduce `Config`, `Client`, `NewClient`, `Option`, and the component token lazy-refresh path (`ComponentAccessToken`, `RefreshComponentToken`).

**Files:**
- Create: `oplatform/client.go`
- Create: `oplatform/component.token.go`
- Create: `oplatform/component_test.go`

- [ ] **Step 7.1: Write failing test `component_test.go`**

```go
package oplatform

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

func testConfig() Config {
	return Config{
		ComponentAppID:     "wxcomp",
		ComponentAppSecret: "secret",
		Token:              "tk",
		EncodingAESKey:     "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQ", // 43 chars
	}
}

func newTestClient(t *testing.T, baseURL string, opts ...Option) *Client {
	t.Helper()
	opts = append(opts, WithHTTP(utils.NewHTTP(baseURL, utils.WithTimeout(time.Second*3))))
	c, err := NewClient(testConfig(), opts...)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestClient_ComponentAccessToken_LazyAndCaches(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/component/api_component_token") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		atomic.AddInt32(&calls, 1)
		_, _ = w.Write([]byte(`{"component_access_token":"CTOK","expires_in":7200}`))
	}))
	defer srv.Close()

	store := NewMemoryStore()
	_ = store.SetVerifyTicket(context.Background(), "TICKET")
	c := newTestClient(t, srv.URL, WithStore(store))

	for i := 0; i < 3; i++ {
		tok, err := c.ComponentAccessToken(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if tok != "CTOK" {
			t.Errorf("got %q", tok)
		}
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("expected 1 fetch, got %d", got)
	}
}

func TestClient_ComponentAccessToken_MissingTicket(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("should not reach the server when ticket is missing: %s", r.URL.Path)
	}))
	defer srv.Close()
	c := newTestClient(t, srv.URL) // default MemoryStore, no ticket

	_, err := c.ComponentAccessToken(context.Background())
	if !errors.Is(err, ErrVerifyTicketMissing) {
		t.Errorf("expected ErrVerifyTicketMissing, got %v", err)
	}
}

func TestClient_ComponentAccessToken_WeixinError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":40013,"errmsg":"invalid appid"}`))
	}))
	defer srv.Close()
	store := NewMemoryStore()
	_ = store.SetVerifyTicket(context.Background(), "TICKET")
	c := newTestClient(t, srv.URL, WithStore(store))

	_, err := c.ComponentAccessToken(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	var werr *WeixinError
	if !errors.As(err, &werr) || werr.ErrCode != 40013 {
		t.Errorf("expected WeixinError 40013, got %v", err)
	}
}

func TestClient_RefreshComponentToken_ForcesFetch(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		_, _ = w.Write([]byte(`{"component_access_token":"CTOK","expires_in":7200}`))
	}))
	defer srv.Close()
	store := NewMemoryStore()
	_ = store.SetVerifyTicket(context.Background(), "TICKET")
	c := newTestClient(t, srv.URL, WithStore(store))

	if _, err := c.ComponentAccessToken(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := c.RefreshComponentToken(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Errorf("expected 2 fetches after forced refresh, got %d", got)
	}
}
```

- [ ] **Step 7.2: Run failing test**

Run: `go test ./oplatform/ -run ComponentAccessToken`
Expected: build errors — `Config`, `Client`, `NewClient`, `Option`, `WithStore`, `WithHTTP`, `ComponentAccessToken`, `RefreshComponentToken` undefined.

- [ ] **Step 7.3: Create `oplatform/client.go`**

```go
package oplatform

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
	"github.com/godrealms/go-wechat-sdk/utils/wxcrypto"
)

// Config 开放平台第三方平台配置。
type Config struct {
	ComponentAppID     string // 第三方平台 appid
	ComponentAppSecret string // 第三方平台 secret
	Token              string // 回调签名 Token
	EncodingAESKey     string // 43 字符 AESKey
}

// Client 开放平台客户端。并发安全。
type Client struct {
	cfg    Config
	http   *utils.HTTP
	store  Store
	crypto *wxcrypto.MsgCrypto

	componentMu sync.Mutex // 保护 component access_token 刷新
	authMu      sync.Map   // map[string]*sync.Mutex per-authorizer 刷新锁
}

// Option 构造可选项。
type Option func(*Client)

// WithStore 注入自定义 Store 实现（默认 MemoryStore）。
func WithStore(s Store) Option {
	return func(c *Client) {
		if s != nil {
			c.store = s
		}
	}
}

// WithHTTP 注入自定义 utils.HTTP 客户端（测试常用）。
func WithHTTP(h *utils.HTTP) Option {
	return func(c *Client) {
		if h != nil {
			c.http = h
		}
	}
}

// NewClient 构造开放平台 Client。构造期间不发起任何网络请求。
func NewClient(cfg Config, opts ...Option) (*Client, error) {
	if cfg.ComponentAppID == "" {
		return nil, fmt.Errorf("oplatform: ComponentAppID is required")
	}
	if cfg.ComponentAppSecret == "" {
		return nil, fmt.Errorf("oplatform: ComponentAppSecret is required")
	}
	if cfg.Token == "" {
		return nil, fmt.Errorf("oplatform: Token is required")
	}
	crypto, err := wxcrypto.New(cfg.Token, cfg.EncodingAESKey, cfg.ComponentAppID)
	if err != nil {
		return nil, fmt.Errorf("oplatform: init crypto: %w", err)
	}
	c := &Client{
		cfg:    cfg,
		http:   utils.NewHTTP("https://api.weixin.qq.com", utils.WithTimeout(time.Second*30)),
		store:  NewMemoryStore(),
		crypto: crypto,
	}
	for _, o := range opts {
		o(c)
	}
	return c, nil
}

// Store 暴露底层 Store，便于外部检查/管理（例如清理失效 authorizer）。
func (c *Client) Store() Store { return c.store }

// HTTP 暴露底层 HTTP 客户端，便于自定义扩展（与 offiaccount/mini-program 保持一致）。
func (c *Client) HTTP() *utils.HTTP { return c.http }

// ComponentAppID 返回第三方平台 appid。
func (c *Client) ComponentAppID() string { return c.cfg.ComponentAppID }

// checkWeixinErr 如果 errcode != 0 则返回 *WeixinError，否则 nil。
func checkWeixinErr(errcode int, errmsg string) error {
	if errcode == 0 {
		return nil
	}
	return &WeixinError{ErrCode: errcode, ErrMsg: errmsg}
}

// touchContext 保证 context 非 nil。
func touchContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
```

- [ ] **Step 7.4: Create `oplatform/component.token.go`**

```go
package oplatform

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// componentTokenSafetyWindow 提前过期 60 秒，避免临界抖动。
const componentTokenSafetyWindow = 60 * time.Second

// ComponentAccessToken 读取 Store 中的 component_access_token；
// 过期则调用微信接口刷新并写回 Store。
func (c *Client) ComponentAccessToken(ctx context.Context) (string, error) {
	ctx = touchContext(ctx)

	// 先乐观读一次（无锁）
	if tok, expireAt, err := c.store.GetComponentToken(ctx); err == nil {
		if time.Now().Add(componentTokenSafetyWindow).Before(expireAt) {
			return tok, nil
		}
	} else if !errors.Is(err, ErrNotFound) {
		return "", fmt.Errorf("oplatform: store get component token: %w", err)
	}

	c.componentMu.Lock()
	defer c.componentMu.Unlock()

	// 双重检查：其他 goroutine 可能已经刷新
	if tok, expireAt, err := c.store.GetComponentToken(ctx); err == nil {
		if time.Now().Add(componentTokenSafetyWindow).Before(expireAt) {
			return tok, nil
		}
	} else if !errors.Is(err, ErrNotFound) {
		return "", fmt.Errorf("oplatform: store get component token: %w", err)
	}

	return c.fetchComponentTokenLocked(ctx)
}

// RefreshComponentToken 无视缓存强制刷新。
func (c *Client) RefreshComponentToken(ctx context.Context) error {
	ctx = touchContext(ctx)
	c.componentMu.Lock()
	defer c.componentMu.Unlock()
	_, err := c.fetchComponentTokenLocked(ctx)
	return err
}

// fetchComponentTokenLocked 必须在 componentMu 持锁时调用。
func (c *Client) fetchComponentTokenLocked(ctx context.Context) (string, error) {
	ticket, err := c.store.GetVerifyTicket(ctx)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", ErrVerifyTicketMissing
		}
		return "", fmt.Errorf("oplatform: store get verify ticket: %w", err)
	}

	body := map[string]string{
		"component_appid":         c.cfg.ComponentAppID,
		"component_appsecret":     c.cfg.ComponentAppSecret,
		"component_verify_ticket": ticket,
	}
	var resp componentTokenResp
	if err := c.http.Post(ctx, "/cgi-bin/component/api_component_token", body, &resp); err != nil {
		return "", fmt.Errorf("oplatform: api_component_token: %w", err)
	}
	if err := checkWeixinErr(resp.ErrCode, resp.ErrMsg); err != nil {
		return "", err
	}
	if resp.ComponentAccessToken == "" {
		return "", fmt.Errorf("oplatform: empty component_access_token")
	}

	expireAt := time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second)
	if err := c.store.SetComponentToken(ctx, resp.ComponentAccessToken, expireAt); err != nil {
		return "", fmt.Errorf("oplatform: store set component token: %w", err)
	}
	return resp.ComponentAccessToken, nil
}
```

- [ ] **Step 7.5: Run tests**

Run: `go test ./oplatform/...`
Expected: PASS for 4 component token tests plus the 4 MemoryStore tests from Task 6.

- [ ] **Step 7.6: Full build**

Run: `go build ./...`
Expected: clean.

- [ ] **Step 7.7: Commit**

```bash
git add oplatform/client.go oplatform/component.token.go oplatform/component_test.go
git commit -m "feat(oplatform): add Client and component_access_token lifecycle

NewClient wires utils.HTTP + MemoryStore + wxcrypto.MsgCrypto.
ComponentAccessToken is lazy with double-checked single-flight:
reads Store, refreshes via /cgi-bin/component/api_component_token
when expired, writes back. RefreshComponentToken forces a fetch.
Missing verify_ticket returns ErrVerifyTicketMissing.

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

## Task 8: oplatform pre_auth_code + authorize URLs

**Goal:** Implement `PreAuthCode`, `AuthorizeURL`, and `MobileAuthorizeURL`.

**Files:**
- Create: `oplatform/component.preauth.go`
- Modify: `oplatform/component_test.go` (add tests)

- [ ] **Step 8.1: Write failing tests**

Append to `oplatform/component_test.go`:

```go
func TestClient_PreAuthCode(t *testing.T) {
	var gotBody string
	mux := http.NewServeMux()
	mux.HandleFunc("/cgi-bin/component/api_component_token", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"component_access_token":"CTOK","expires_in":7200}`))
	})
	mux.HandleFunc("/cgi-bin/component/api_create_preauthcode", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("component_access_token") != "CTOK" {
			t.Errorf("missing component_access_token in query")
		}
		bb, _ := io.ReadAll(r.Body)
		gotBody = string(bb)
		_, _ = w.Write([]byte(`{"pre_auth_code":"PREAUTH","expires_in":600}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	store := NewMemoryStore()
	_ = store.SetVerifyTicket(context.Background(), "TICKET")
	c := newTestClient(t, srv.URL, WithStore(store))

	code, err := c.PreAuthCode(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if code != "PREAUTH" {
		t.Errorf("got %q", code)
	}
	if !strings.Contains(gotBody, `"component_appid":"wxcomp"`) {
		t.Errorf("unexpected body: %s", gotBody)
	}
}

func TestClient_AuthorizeURL(t *testing.T) {
	c, err := NewClient(testConfig())
	if err != nil {
		t.Fatal(err)
	}
	u := c.AuthorizeURL("PC", "https://example.com/cb", 3, "")
	if !strings.Contains(u, "component_appid=wxcomp") {
		t.Errorf("missing component_appid: %s", u)
	}
	if !strings.Contains(u, "pre_auth_code=PC") {
		t.Errorf("missing pre_auth_code: %s", u)
	}
	if !strings.Contains(u, "redirect_uri=https%3A%2F%2Fexample.com%2Fcb") {
		t.Errorf("redirect_uri not encoded: %s", u)
	}
	if !strings.Contains(u, "auth_type=3") {
		t.Errorf("missing auth_type: %s", u)
	}
	if strings.Contains(u, "biz_appid") {
		t.Errorf("biz_appid should be absent when empty: %s", u)
	}
}

func TestClient_AuthorizeURL_WithBizAppid(t *testing.T) {
	c, _ := NewClient(testConfig())
	u := c.AuthorizeURL("PC", "https://x/cb", 1, "wxbiz")
	if !strings.Contains(u, "biz_appid=wxbiz") {
		t.Errorf("biz_appid missing: %s", u)
	}
}

func TestClient_MobileAuthorizeURL(t *testing.T) {
	c, _ := NewClient(testConfig())
	u := c.MobileAuthorizeURL("PC", "https://x/cb", 3, "")
	if !strings.Contains(u, "action=bindcomponent") {
		t.Errorf("missing action=bindcomponent: %s", u)
	}
	if !strings.Contains(u, "pre_auth_code=PC") {
		t.Errorf("missing pre_auth_code: %s", u)
	}
}
```

Add `"io"` to the test file's import block if not yet present.

- [ ] **Step 8.2: Run failing test**

Run: `go test ./oplatform/ -run "PreAuthCode|AuthorizeURL"`
Expected: undefined `PreAuthCode`, `AuthorizeURL`, `MobileAuthorizeURL`.

- [ ] **Step 8.3: Create `oplatform/component.preauth.go`**

```go
package oplatform

import (
	"context"
	"fmt"
	"net/url"
)

// PreAuthCode 调用 /cgi-bin/component/api_create_preauthcode 换预授权码。
// 预授权码 TTL ~10 分钟，调用方应每次使用前重新调用本方法。
func (c *Client) PreAuthCode(ctx context.Context) (string, error) {
	ctx = touchContext(ctx)
	token, err := c.ComponentAccessToken(ctx)
	if err != nil {
		return "", err
	}
	q := url.Values{"component_access_token": {token}}
	body := map[string]string{"component_appid": c.cfg.ComponentAppID}

	var resp preAuthCodeResp
	path := "/cgi-bin/component/api_create_preauthcode?" + q.Encode()
	if err := c.http.Post(ctx, path, body, &resp); err != nil {
		return "", fmt.Errorf("oplatform: api_create_preauthcode: %w", err)
	}
	if err := checkWeixinErr(resp.ErrCode, resp.ErrMsg); err != nil {
		return "", err
	}
	if resp.PreAuthCode == "" {
		return "", fmt.Errorf("oplatform: empty pre_auth_code")
	}
	return resp.PreAuthCode, nil
}

// AuthorizeURL 构造 PC 版本的引导授权跳转 URL。
//
//	preAuthCode - 来自 PreAuthCode()
//	redirectURI - 授权完成后回调地址
//	authType    - 1=公众号 2=小程序 3=全部；当 bizAppid 非空时被忽略
//	bizAppid    - 指定授权方 appid；不指定传 ""
func (c *Client) AuthorizeURL(preAuthCode, redirectURI string, authType int, bizAppid string) string {
	q := url.Values{
		"component_appid": {c.cfg.ComponentAppID},
		"pre_auth_code":   {preAuthCode},
		"redirect_uri":    {redirectURI},
	}
	if bizAppid != "" {
		q.Set("biz_appid", bizAppid)
	} else {
		q.Set("auth_type", fmt.Sprintf("%d", authType))
	}
	return "https://mp.weixin.qq.com/cgi-bin/componentloginpage?" + q.Encode()
}

// MobileAuthorizeURL 构造移动端（扫码）授权跳转 URL。
func (c *Client) MobileAuthorizeURL(preAuthCode, redirectURI string, authType int, bizAppid string) string {
	q := url.Values{
		"action":          {"bindcomponent"},
		"no_scan":         {"1"},
		"component_appid": {c.cfg.ComponentAppID},
		"pre_auth_code":   {preAuthCode},
		"redirect_uri":    {redirectURI},
	}
	if bizAppid != "" {
		q.Set("biz_appid", bizAppid)
	} else {
		q.Set("auth_type", fmt.Sprintf("%d", authType))
	}
	return "https://mp.weixin.qq.com/safe/bindcomponent?" + q.Encode() + "#wechat_redirect"
}
```

- [ ] **Step 8.4: Run tests**

Run: `go test ./oplatform/...`
Expected: all tests PASS.

- [ ] **Step 8.5: Commit**

```bash
git add oplatform/component.preauth.go oplatform/component_test.go
git commit -m "feat(oplatform): add pre_auth_code + AuthorizeURL helpers

Implements PreAuthCode (calls api_create_preauthcode), AuthorizeURL
(PC entry) and MobileAuthorizeURL (bindcomponent entry). Handles both
auth_type and biz_appid branches.

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

## Task 9: oplatform QueryAuth + authorizer info family

**Goal:** Implement `QueryAuth` (authorization_code → authorizer tokens, with Store write-through), `GetAuthorizerInfo`, `GetAuthorizerOption`, `SetAuthorizerOption`, `GetAuthorizerList`.

**Files:**
- Create: `oplatform/component.authorize.go`
- Modify: `oplatform/component_test.go` (add tests)

- [ ] **Step 9.1: Write failing tests**

Append to `oplatform/component_test.go`:

```go
func TestClient_QueryAuth_PopulatesStore(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/cgi-bin/component/api_component_token", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"component_access_token":"CTOK","expires_in":7200}`))
	})
	mux.HandleFunc("/cgi-bin/component/api_query_auth", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
  "authorization_info": {
    "authorizer_appid": "wxAuthed",
    "authorizer_access_token": "ATOK",
    "expires_in": 7200,
    "authorizer_refresh_token": "RTOK",
    "func_info": [{"funcscope_category": {"id": 1}}]
  }
}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	store := NewMemoryStore()
	_ = store.SetVerifyTicket(context.Background(), "TICKET")
	c := newTestClient(t, srv.URL, WithStore(store))

	info, err := c.QueryAuth(context.Background(), "AUTHCODE")
	if err != nil {
		t.Fatal(err)
	}
	if info.AuthorizerAppID != "wxAuthed" {
		t.Errorf("appid mismatch: %+v", info)
	}
	// Store must be populated
	got, err := store.GetAuthorizer(context.Background(), "wxAuthed")
	if err != nil {
		t.Fatal(err)
	}
	if got.AccessToken != "ATOK" || got.RefreshToken != "RTOK" {
		t.Errorf("store mismatch: %+v", got)
	}
	if !got.ExpireAt.After(time.Now()) {
		t.Errorf("expire_at should be future, got %v", got.ExpireAt)
	}
}

func TestClient_GetAuthorizerInfo(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/cgi-bin/component/api_component_token", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"component_access_token":"CTOK","expires_in":7200}`))
	})
	mux.HandleFunc("/cgi-bin/component/api_get_authorizer_info", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
  "authorizer_info": {"nick_name":"biz","user_name":"gh_x","principal_name":"Acme"},
  "authorization_info": {"authorizer_appid":"wxAuthed","authorizer_refresh_token":"RTOK"}
}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	store := NewMemoryStore()
	_ = store.SetVerifyTicket(context.Background(), "TICKET")
	c := newTestClient(t, srv.URL, WithStore(store))

	got, err := c.GetAuthorizerInfo(context.Background(), "wxAuthed")
	if err != nil {
		t.Fatal(err)
	}
	if got.AuthorizerInfo.NickName != "biz" || got.AuthorizerInfo.PrincipalName != "Acme" {
		t.Errorf("unexpected: %+v", got)
	}
}

func TestClient_GetAuthorizerList(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/cgi-bin/component/api_component_token", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"component_access_token":"CTOK","expires_in":7200}`))
	})
	mux.HandleFunc("/cgi-bin/component/api_get_authorizer_list", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
  "total_count": 2,
  "list": [
    {"authorizer_appid":"wxA","refresh_token":"rA","auth_time":1700000000},
    {"authorizer_appid":"wxB","refresh_token":"rB","auth_time":1700000001}
  ]
}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	store := NewMemoryStore()
	_ = store.SetVerifyTicket(context.Background(), "TICKET")
	c := newTestClient(t, srv.URL, WithStore(store))

	list, err := c.GetAuthorizerList(context.Background(), 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if list.TotalCount != 2 || len(list.List) != 2 {
		t.Errorf("unexpected: %+v", list)
	}
}

func TestClient_GetSetAuthorizerOption(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/cgi-bin/component/api_component_token", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"component_access_token":"CTOK","expires_in":7200}`))
	})
	mux.HandleFunc("/cgi-bin/component/api_get_authorizer_option", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"authorizer_appid":"wxA","option_name":"voice_recognize","option_value":"1"}`))
	})
	mux.HandleFunc("/cgi-bin/component/api_set_authorizer_option", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	store := NewMemoryStore()
	_ = store.SetVerifyTicket(context.Background(), "TICKET")
	c := newTestClient(t, srv.URL, WithStore(store))

	opt, err := c.GetAuthorizerOption(context.Background(), "wxA", "voice_recognize")
	if err != nil {
		t.Fatal(err)
	}
	if opt.OptionValue != "1" {
		t.Errorf("unexpected: %+v", opt)
	}
	if err := c.SetAuthorizerOption(context.Background(), "wxA", "voice_recognize", "0"); err != nil {
		t.Fatal(err)
	}
}
```

- [ ] **Step 9.2: Run failing test**

Run: `go test ./oplatform/ -run "QueryAuth|Authorizer"`
Expected: undefined `QueryAuth`, `GetAuthorizerInfo`, `GetAuthorizerOption`, `SetAuthorizerOption`, `GetAuthorizerList`.

- [ ] **Step 9.3: Create `oplatform/component.authorize.go`**

```go
package oplatform

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// QueryAuth 用 authorization_code 换取 authorizer 的 access_token / refresh_token。
// 成功后自动写入 Store（key = authorizer_appid）。
func (c *Client) QueryAuth(ctx context.Context, authCode string) (*AuthorizationInfo, error) {
	ctx = touchContext(ctx)
	if authCode == "" {
		return nil, fmt.Errorf("oplatform: authCode is required")
	}
	token, err := c.ComponentAccessToken(ctx)
	if err != nil {
		return nil, err
	}
	q := url.Values{"component_access_token": {token}}
	body := map[string]string{
		"component_appid":    c.cfg.ComponentAppID,
		"authorization_code": authCode,
	}
	var resp queryAuthResp
	path := "/cgi-bin/component/api_query_auth?" + q.Encode()
	if err := c.http.Post(ctx, path, body, &resp); err != nil {
		return nil, fmt.Errorf("oplatform: api_query_auth: %w", err)
	}
	if err := checkWeixinErr(resp.ErrCode, resp.ErrMsg); err != nil {
		return nil, err
	}
	info := resp.AuthorizationInfo
	if info.AuthorizerAppID == "" {
		return nil, fmt.Errorf("oplatform: api_query_auth returned empty authorizer_appid")
	}
	tokens := AuthorizerTokens{
		AccessToken:  info.AuthorizerAccessToken,
		RefreshToken: info.AuthorizerRefreshToken,
		ExpireAt:     time.Now().Add(time.Duration(info.ExpiresIn) * time.Second),
	}
	if err := c.store.SetAuthorizer(ctx, info.AuthorizerAppID, tokens); err != nil {
		return nil, fmt.Errorf("oplatform: store set authorizer: %w", err)
	}
	return &info, nil
}

// GetAuthorizerInfo /cgi-bin/component/api_get_authorizer_info
func (c *Client) GetAuthorizerInfo(ctx context.Context, authorizerAppID string) (*AuthorizerInfoResp, error) {
	ctx = touchContext(ctx)
	token, err := c.ComponentAccessToken(ctx)
	if err != nil {
		return nil, err
	}
	q := url.Values{"component_access_token": {token}}
	body := map[string]string{
		"component_appid":  c.cfg.ComponentAppID,
		"authorizer_appid": authorizerAppID,
	}
	var resp AuthorizerInfoResp
	path := "/cgi-bin/component/api_get_authorizer_info?" + q.Encode()
	if err := c.http.Post(ctx, path, body, &resp); err != nil {
		return nil, fmt.Errorf("oplatform: api_get_authorizer_info: %w", err)
	}
	if err := checkWeixinErr(resp.ErrCode, resp.ErrMsg); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetAuthorizerOption /cgi-bin/component/api_get_authorizer_option
func (c *Client) GetAuthorizerOption(ctx context.Context, authorizerAppID, optionName string) (*AuthorizerOption, error) {
	ctx = touchContext(ctx)
	token, err := c.ComponentAccessToken(ctx)
	if err != nil {
		return nil, err
	}
	q := url.Values{"component_access_token": {token}}
	body := map[string]string{
		"component_appid":  c.cfg.ComponentAppID,
		"authorizer_appid": authorizerAppID,
		"option_name":      optionName,
	}
	var resp AuthorizerOption
	path := "/cgi-bin/component/api_get_authorizer_option?" + q.Encode()
	if err := c.http.Post(ctx, path, body, &resp); err != nil {
		return nil, fmt.Errorf("oplatform: api_get_authorizer_option: %w", err)
	}
	if err := checkWeixinErr(resp.ErrCode, resp.ErrMsg); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SetAuthorizerOption /cgi-bin/component/api_set_authorizer_option
func (c *Client) SetAuthorizerOption(ctx context.Context, authorizerAppID, optionName, optionValue string) error {
	ctx = touchContext(ctx)
	token, err := c.ComponentAccessToken(ctx)
	if err != nil {
		return err
	}
	q := url.Values{"component_access_token": {token}}
	body := map[string]string{
		"component_appid":  c.cfg.ComponentAppID,
		"authorizer_appid": authorizerAppID,
		"option_name":      optionName,
		"option_value":     optionValue,
	}
	var resp struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	path := "/cgi-bin/component/api_set_authorizer_option?" + q.Encode()
	if err := c.http.Post(ctx, path, body, &resp); err != nil {
		return fmt.Errorf("oplatform: api_set_authorizer_option: %w", err)
	}
	return checkWeixinErr(resp.ErrCode, resp.ErrMsg)
}

// GetAuthorizerList /cgi-bin/component/api_get_authorizer_list
func (c *Client) GetAuthorizerList(ctx context.Context, offset, count int) (*AuthorizerList, error) {
	ctx = touchContext(ctx)
	token, err := c.ComponentAccessToken(ctx)
	if err != nil {
		return nil, err
	}
	q := url.Values{"component_access_token": {token}}
	payload := struct {
		ComponentAppID string `json:"component_appid"`
		Offset         int    `json:"offset"`
		Count          int    `json:"count"`
	}{c.cfg.ComponentAppID, offset, count}

	var resp AuthorizerList
	path := "/cgi-bin/component/api_get_authorizer_list?" + q.Encode()
	if err := c.http.Post(ctx, path, payload, &resp); err != nil {
		return nil, fmt.Errorf("oplatform: api_get_authorizer_list: %w", err)
	}
	if err := checkWeixinErr(resp.ErrCode, resp.ErrMsg); err != nil {
		return nil, err
	}
	return &resp, nil
}
```

- [ ] **Step 9.4: Run tests**

Run: `go test ./oplatform/...`
Expected: all tests PASS.

- [ ] **Step 9.5: Commit**

```bash
git add oplatform/component.authorize.go oplatform/component_test.go
git commit -m "feat(oplatform): add QueryAuth and authorizer info/option/list

QueryAuth exchanges authorization_code for authorizer tokens and
writes them to Store. GetAuthorizerInfo/Option/SetOption/List are
thin wrappers over the corresponding api_* endpoints, all gated by
ComponentAccessToken.

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

## Task 10: oplatform AuthorizerClient — access token refresh

**Goal:** Implement `AuthorizerClient.AccessToken` (lazy + per-appid single-flight), `AuthorizerClient.Refresh`, and the `ErrAuthorizerRevoked` mapping from errcode 61023.

**Files:**
- Create: `oplatform/component.authorizer.token.go`
- Create: `oplatform/authorizer_test.go`

- [ ] **Step 10.1: Write failing test `authorizer_test.go`**

```go
package oplatform

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestAuthorizerClient_AccessToken_LazyAndCaches(t *testing.T) {
	var refreshCalls int32
	mux := http.NewServeMux()
	mux.HandleFunc("/cgi-bin/component/api_component_token", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"component_access_token":"CTOK","expires_in":7200}`))
	})
	mux.HandleFunc("/cgi-bin/component/api_authorizer_token", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&refreshCalls, 1)
		_, _ = w.Write([]byte(`{"authorizer_access_token":"A1","authorizer_refresh_token":"R1","expires_in":7200}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	store := NewMemoryStore()
	_ = store.SetVerifyTicket(context.Background(), "TICKET")
	// Pre-seed expired tokens so the first call triggers refresh
	_ = store.SetAuthorizer(context.Background(), "wxA", AuthorizerTokens{
		AccessToken:  "old",
		RefreshToken: "R0",
		ExpireAt:     time.Now().Add(-time.Minute),
	})
	c := newTestClient(t, srv.URL, WithStore(store))
	auth := c.Authorizer("wxA")

	for i := 0; i < 3; i++ {
		tok, err := auth.AccessToken(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if tok != "A1" {
			t.Errorf("got %q", tok)
		}
	}
	if got := atomic.LoadInt32(&refreshCalls); got != 1 {
		t.Errorf("expected 1 refresh, got %d", got)
	}
}

func TestAuthorizerClient_AccessToken_RefreshRevoked(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/cgi-bin/component/api_component_token", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"component_access_token":"CTOK","expires_in":7200}`))
	})
	mux.HandleFunc("/cgi-bin/component/api_authorizer_token", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":61023,"errmsg":"invalid refresh_token"}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	store := NewMemoryStore()
	_ = store.SetVerifyTicket(context.Background(), "TICKET")
	_ = store.SetAuthorizer(context.Background(), "wxA", AuthorizerTokens{
		AccessToken: "old", RefreshToken: "Rbad", ExpireAt: time.Now().Add(-time.Minute),
	})
	c := newTestClient(t, srv.URL, WithStore(store))
	auth := c.Authorizer("wxA")

	_, err := auth.AccessToken(context.Background())
	if !errors.Is(err, ErrAuthorizerRevoked) {
		t.Errorf("expected ErrAuthorizerRevoked, got %v", err)
	}
}

func TestAuthorizerClient_Refresh_ForcesFetch(t *testing.T) {
	var refreshCalls int32
	mux := http.NewServeMux()
	mux.HandleFunc("/cgi-bin/component/api_component_token", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"component_access_token":"CTOK","expires_in":7200}`))
	})
	mux.HandleFunc("/cgi-bin/component/api_authorizer_token", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&refreshCalls, 1)
		_, _ = w.Write([]byte(`{"authorizer_access_token":"A2","authorizer_refresh_token":"R2","expires_in":7200}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	store := NewMemoryStore()
	_ = store.SetVerifyTicket(context.Background(), "TICKET")
	_ = store.SetAuthorizer(context.Background(), "wxA", AuthorizerTokens{
		AccessToken: "cached", RefreshToken: "R0", ExpireAt: time.Now().Add(time.Hour),
	})
	c := newTestClient(t, srv.URL, WithStore(store))
	auth := c.Authorizer("wxA")

	// First call should hit cache (not call refresh)
	tok, err := auth.AccessToken(context.Background())
	if err != nil || tok != "cached" {
		t.Fatalf("expected cached, got %q err=%v", tok, err)
	}
	// Force refresh
	if err := auth.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := atomic.LoadInt32(&refreshCalls); got != 1 {
		t.Errorf("expected 1 refresh, got %d", got)
	}
	// Store now has A2
	stored, _ := store.GetAuthorizer(context.Background(), "wxA")
	if stored.AccessToken != "A2" || stored.RefreshToken != "R2" {
		t.Errorf("store mismatch: %+v", stored)
	}
}
```

- [ ] **Step 10.2: Run failing test**

Run: `go test ./oplatform/ -run Authorizer`
Expected: undefined `Authorizer`, `AuthorizerClient`, method `AccessToken`, method `Refresh`.

- [ ] **Step 10.3: Create `oplatform/component.authorizer.token.go`**

```go
package oplatform

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"
)

const authorizerTokenSafetyWindow = 60 * time.Second

// errcodeAuthorizerRevoked 微信返回 refresh_token 失效。
const errcodeAuthorizerRevoked = 61023

// AuthorizerClient 是代单个 authorizer 调用微信接口的句柄。
// 同时实现 offiaccount.TokenSource 和 mini_program.TokenSource。
type AuthorizerClient struct {
	c     *Client
	appID string
}

// Authorizer 返回一个指向某 authorizer 的句柄。构造不做 I/O。
func (c *Client) Authorizer(authorizerAppID string) *AuthorizerClient {
	return &AuthorizerClient{c: c, appID: authorizerAppID}
}

// AppID 返回被授权方 appid。
func (a *AuthorizerClient) AppID() string { return a.appID }

// AccessToken 返回 authorizer_access_token，过期自动刷新。
// 这个签名恰好匹配 offiaccount.TokenSource 和 mini_program.TokenSource。
func (a *AuthorizerClient) AccessToken(ctx context.Context) (string, error) {
	ctx = touchContext(ctx)

	// Lazy read
	tokens, err := a.c.store.GetAuthorizer(ctx, a.appID)
	if err == nil && time.Now().Add(authorizerTokenSafetyWindow).Before(tokens.ExpireAt) {
		return tokens.AccessToken, nil
	}
	if err != nil && !errors.Is(err, ErrNotFound) {
		return "", fmt.Errorf("oplatform: store get authorizer: %w", err)
	}

	mu := a.lockFor(a.appID)
	mu.Lock()
	defer mu.Unlock()

	// Double check
	tokens, err = a.c.store.GetAuthorizer(ctx, a.appID)
	if err == nil && time.Now().Add(authorizerTokenSafetyWindow).Before(tokens.ExpireAt) {
		return tokens.AccessToken, nil
	}
	if err != nil && !errors.Is(err, ErrNotFound) {
		return "", fmt.Errorf("oplatform: store get authorizer: %w", err)
	}

	return a.refreshLocked(ctx, tokens.RefreshToken)
}

// Refresh 强制刷新。
func (a *AuthorizerClient) Refresh(ctx context.Context) error {
	ctx = touchContext(ctx)

	mu := a.lockFor(a.appID)
	mu.Lock()
	defer mu.Unlock()

	tokens, err := a.c.store.GetAuthorizer(ctx, a.appID)
	if err != nil {
		return fmt.Errorf("oplatform: store get authorizer: %w", err)
	}
	_, err = a.refreshLocked(ctx, tokens.RefreshToken)
	return err
}

// refreshLocked must be called with the per-appid mutex held.
func (a *AuthorizerClient) refreshLocked(ctx context.Context, refreshToken string) (string, error) {
	if refreshToken == "" {
		return "", fmt.Errorf("oplatform: no refresh_token for authorizer %s", a.appID)
	}
	componentToken, err := a.c.ComponentAccessToken(ctx)
	if err != nil {
		return "", err
	}
	q := url.Values{"component_access_token": {componentToken}}
	body := map[string]string{
		"component_appid":          a.c.cfg.ComponentAppID,
		"authorizer_appid":         a.appID,
		"authorizer_refresh_token": refreshToken,
	}
	var resp authorizerTokenResp
	path := "/cgi-bin/component/api_authorizer_token?" + q.Encode()
	if err := a.c.http.Post(ctx, path, body, &resp); err != nil {
		return "", fmt.Errorf("oplatform: api_authorizer_token: %w", err)
	}
	if resp.ErrCode == errcodeAuthorizerRevoked {
		return "", ErrAuthorizerRevoked
	}
	if err := checkWeixinErr(resp.ErrCode, resp.ErrMsg); err != nil {
		return "", err
	}
	if resp.AuthorizerAccessToken == "" {
		return "", fmt.Errorf("oplatform: empty authorizer_access_token")
	}
	tokens := AuthorizerTokens{
		AccessToken:  resp.AuthorizerAccessToken,
		RefreshToken: resp.AuthorizerRefreshToken,
		ExpireAt:     time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second),
	}
	if tokens.RefreshToken == "" {
		// 微信并不总是下发新的 refresh_token，此时沿用旧的
		tokens.RefreshToken = refreshToken
	}
	if err := a.c.store.SetAuthorizer(ctx, a.appID, tokens); err != nil {
		return "", fmt.Errorf("oplatform: store set authorizer: %w", err)
	}
	return tokens.AccessToken, nil
}

func (a *AuthorizerClient) lockFor(appid string) *sync.Mutex {
	if mu, ok := a.c.authMu.Load(appid); ok {
		return mu.(*sync.Mutex)
	}
	mu := &sync.Mutex{}
	actual, _ := a.c.authMu.LoadOrStore(appid, mu)
	return actual.(*sync.Mutex)
}
```

- [ ] **Step 10.4: Run tests**

Run: `go test ./oplatform/ -run Authorizer`
Expected: 3 PASS.

Run full: `go test -race ./oplatform/...`
Expected: all PASS, no race.

- [ ] **Step 10.5: Commit**

```bash
git add oplatform/component.authorizer.token.go oplatform/authorizer_test.go
git commit -m "feat(oplatform): add AuthorizerClient lazy token refresh

AuthorizerClient reads/writes AuthorizerTokens through Store,
refreshes via /cgi-bin/component/api_authorizer_token when
expired, and uses per-appid sync.Mutex via Client.authMu
(sync.Map) to avoid cross-authorizer blocking. errcode 61023
maps to ErrAuthorizerRevoked so callers can re-trigger auth.

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

## Task 11: oplatform `authorizer.go` — OffiaccountClient / MiniProgramClient / RefreshAll

**Goal:** Bridge `AuthorizerClient` into the existing `offiaccount` and `mini_program` packages via `TokenSource` injection, and provide `Client.RefreshAll` for bulk warm-up.

**Files:**
- Create: `oplatform/authorizer.go`
- Modify: `oplatform/authorizer_test.go` (add bridge tests)

- [ ] **Step 11.1: Write failing tests**

Append to `oplatform/authorizer_test.go`:

```go
import (
	"github.com/godrealms/go-wechat-sdk/mini-program"
	"github.com/godrealms/go-wechat-sdk/offiaccount"
	"github.com/godrealms/go-wechat-sdk/utils"
)
```

> If the test file already has imports, merge these in. Go compiler will reject duplicate imports, so ensure only one import block.

Add the test funcs (keeping existing ones):

```go
func TestAuthorizerClient_OffiaccountClient_UsesAuthorizerToken(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/cgi-bin/component/api_component_token", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"component_access_token":"CTOK","expires_in":7200}`))
	})
	mux.HandleFunc("/cgi-bin/component/api_authorizer_token", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"authorizer_access_token":"AUTH_A","authorizer_refresh_token":"R","expires_in":7200}`))
	})
	// offiaccount.Client.AccessTokenE via TokenSource must NOT call /cgi-bin/token
	mux.HandleFunc("/cgi-bin/token", func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("cgi-bin/token should not be called through TokenSource")
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	store := NewMemoryStore()
	_ = store.SetVerifyTicket(context.Background(), "TICKET")
	_ = store.SetAuthorizer(context.Background(), "wxBiz", AuthorizerTokens{
		AccessToken: "x", RefreshToken: "R", ExpireAt: time.Now().Add(-time.Minute),
	})
	c := newTestClient(t, srv.URL, WithStore(store))
	auth := c.Authorizer("wxBiz")

	off := auth.OffiaccountClient(offiaccount.WithHTTPClient(utils.NewHTTP(srv.URL, utils.WithTimeout(time.Second*3))))
	tok, err := off.AccessTokenE(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if tok != "AUTH_A" {
		t.Errorf("offiaccount token should come from oplatform, got %q", tok)
	}
}

func TestAuthorizerClient_MiniProgramClient_UsesAuthorizerToken(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/cgi-bin/component/api_component_token", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"component_access_token":"CTOK","expires_in":7200}`))
	})
	mux.HandleFunc("/cgi-bin/component/api_authorizer_token", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"authorizer_access_token":"AUTH_MP","authorizer_refresh_token":"R","expires_in":7200}`))
	})
	mux.HandleFunc("/cgi-bin/token", func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("cgi-bin/token should not be called")
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	store := NewMemoryStore()
	_ = store.SetVerifyTicket(context.Background(), "TICKET")
	_ = store.SetAuthorizer(context.Background(), "wxMP", AuthorizerTokens{
		AccessToken: "x", RefreshToken: "R", ExpireAt: time.Now().Add(-time.Minute),
	})
	c := newTestClient(t, srv.URL, WithStore(store))
	auth := c.Authorizer("wxMP")

	mp, err := auth.MiniProgramClient(mini_program.WithHTTP(utils.NewHTTP(srv.URL, utils.WithTimeout(time.Second*3))))
	if err != nil {
		t.Fatal(err)
	}
	tok, err := mp.AccessToken(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if tok != "AUTH_MP" {
		t.Errorf("mini-program token should come from oplatform, got %q", tok)
	}
}

func TestClient_RefreshAll(t *testing.T) {
	var refreshCalls int32
	mux := http.NewServeMux()
	mux.HandleFunc("/cgi-bin/component/api_component_token", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"component_access_token":"CTOK","expires_in":7200}`))
	})
	mux.HandleFunc("/cgi-bin/component/api_authorizer_token", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&refreshCalls, 1)
		_, _ = w.Write([]byte(`{"authorizer_access_token":"NEW","authorizer_refresh_token":"R","expires_in":7200}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	store := NewMemoryStore()
	_ = store.SetVerifyTicket(context.Background(), "TICKET")
	_ = store.SetAuthorizer(context.Background(), "wxA", AuthorizerTokens{
		AccessToken: "old", RefreshToken: "R", ExpireAt: time.Now().Add(time.Hour),
	})
	_ = store.SetAuthorizer(context.Background(), "wxB", AuthorizerTokens{
		AccessToken: "old", RefreshToken: "R", ExpireAt: time.Now().Add(time.Hour),
	})
	c := newTestClient(t, srv.URL, WithStore(store))

	if err := c.RefreshAll(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := atomic.LoadInt32(&refreshCalls); got != 2 {
		t.Errorf("expected 2 refreshes, got %d", got)
	}
}
```

- [ ] **Step 11.2: Run failing test**

Run: `go test ./oplatform/ -run "OffiaccountClient|MiniProgramClient|RefreshAll"`
Expected: undefined `OffiaccountClient`, `MiniProgramClient`, `RefreshAll`.

- [ ] **Step 11.3: Create `oplatform/authorizer.go`**

```go
package oplatform

import (
	"context"
	"errors"
	"fmt"
	"strings"

	mini_program "github.com/godrealms/go-wechat-sdk/mini-program"
	"github.com/godrealms/go-wechat-sdk/offiaccount"
)

// Compile-time assertions that AuthorizerClient satisfies both TokenSource shapes.
var (
	_ offiaccount.TokenSource  = (*AuthorizerClient)(nil)
	_ mini_program.TokenSource = (*AuthorizerClient)(nil)
)

// OffiaccountClient 返回一个预先注入了 AuthorizerClient 作为 TokenSource
// 的 offiaccount.Client。之后调用 off.AccessTokenE / 菜单 / 模板消息 / 素材
// 等任意 offiaccount 方法时，底层 token 自动来自开放平台。
func (a *AuthorizerClient) OffiaccountClient(opts ...offiaccount.Option) *offiaccount.Client {
	allOpts := append([]offiaccount.Option{offiaccount.WithTokenSource(a)}, opts...)
	return offiaccount.NewClient(context.Background(), &offiaccount.Config{AppId: a.appID}, allOpts...)
}

// MiniProgramClient 返回一个预先注入了 AuthorizerClient 作为 TokenSource
// 的 mini_program.Client。
func (a *AuthorizerClient) MiniProgramClient(opts ...mini_program.Option) (*mini_program.Client, error) {
	allOpts := append([]mini_program.Option{mini_program.WithTokenSource(a)}, opts...)
	return mini_program.NewClient(mini_program.Config{AppId: a.appID}, allOpts...)
}

// RefreshAll 对 Store 中所有已登记的 authorizer 调用 Refresh。
// 用于启动预热或外部 cron 触发。单个 appid 失败不中断循环，
// 所有错误汇总后以多行错误字符串返回。
func (c *Client) RefreshAll(ctx context.Context) error {
	ctx = touchContext(ctx)
	ids, err := c.store.ListAuthorizerAppIDs(ctx)
	if err != nil {
		return fmt.Errorf("oplatform: list authorizers: %w", err)
	}
	var errs []string
	for _, id := range ids {
		auth := c.Authorizer(id)
		if err := auth.Refresh(ctx); err != nil {
			if errors.Is(err, ErrAuthorizerRevoked) {
				errs = append(errs, fmt.Sprintf("%s: revoked", id))
				continue
			}
			errs = append(errs, fmt.Sprintf("%s: %v", id, err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("oplatform: RefreshAll had %d failures:\n%s",
			len(errs), strings.Join(errs, "\n"))
	}
	return nil
}
```

- [ ] **Step 11.4: Run tests**

Run: `go test ./oplatform/...`
Expected: all tests PASS including bridge and RefreshAll.

- [ ] **Step 11.5: Full repo build + tests**

Run: `go build ./... && go test ./...`
Expected: clean.

- [ ] **Step 11.6: Commit**

```bash
git add oplatform/authorizer.go oplatform/authorizer_test.go
git commit -m "feat(oplatform): bridge AuthorizerClient into offiaccount/mini-program

AuthorizerClient now satisfies both offiaccount.TokenSource and
mini_program.TokenSource (compile-time checked). OffiaccountClient()
and MiniProgramClient() return pre-wired clients that transparently
use oplatform as their token source — the full offiaccount and
mini-program API surface becomes available for authorized users
without any new wrapper code.

Client.RefreshAll iterates Store.ListAuthorizerAppIDs and refreshes
each, aggregating errors so one bad authorizer does not block others.

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

## Task 12: oplatform `ParseNotify` — callback decryption with 4 InfoType branches

**Goal:** Parse the encrypted callback payload (verify_ticket push, authorized / updateauthorized / unauthorized events), auto-write `verify_ticket` to Store, return `*ComponentNotify`.

**Files:**
- Create: `oplatform/notify.go`
- Create: `oplatform/notify_test.go`

- [ ] **Step 12.1: Write failing test `notify_test.go`**

```go
package oplatform

import (
	"context"
	"encoding/xml"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// helper — build an encrypted POST request for a given plaintext XML payload
func buildEncryptedReq(t *testing.T, c *Client, plain string, ts, nonce string) *http.Request {
	t.Helper()
	encrypted, err := c.crypto.Encrypt([]byte(plain))
	if err != nil {
		t.Fatal(err)
	}
	sig := c.crypto.Signature(ts, nonce, encrypted)
	body := `<xml><ToUserName><![CDATA[wxcomp]]></ToUserName><Encrypt><![CDATA[` + encrypted + `]]></Encrypt></xml>`
	req := httptest.NewRequest(http.MethodPost,
		"/oplatform/notify?msg_signature="+sig+"&timestamp="+ts+"&nonce="+nonce,
		strings.NewReader(body))
	return req
}

func TestParseNotify_VerifyTicket_AutoWritesStore(t *testing.T) {
	c, _ := NewClient(testConfig())
	plain := `<xml>
<AppId><![CDATA[wxcomp]]></AppId>
<CreateTime>1700000000</CreateTime>
<InfoType><![CDATA[component_verify_ticket]]></InfoType>
<ComponentVerifyTicket><![CDATA[TICKET_ABC]]></ComponentVerifyTicket>
</xml>`
	req := buildEncryptedReq(t, c, plain, "1700000000", "nonceA")

	notify, err := c.ParseNotify(req, nil)
	if err != nil {
		t.Fatal(err)
	}
	if notify.InfoType != "component_verify_ticket" {
		t.Errorf("info_type mismatch: %q", notify.InfoType)
	}
	if notify.ComponentVerifyTicket != "TICKET_ABC" {
		t.Errorf("ticket mismatch: %q", notify.ComponentVerifyTicket)
	}
	got, err := c.store.GetVerifyTicket(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got != "TICKET_ABC" {
		t.Errorf("store ticket mismatch: %q", got)
	}
}

func TestParseNotify_Authorized(t *testing.T) {
	c, _ := NewClient(testConfig())
	plain := `<xml>
<AppId><![CDATA[wxcomp]]></AppId>
<CreateTime>1700000001</CreateTime>
<InfoType><![CDATA[authorized]]></InfoType>
<AuthorizerAppid><![CDATA[wxAuthed]]></AuthorizerAppid>
<AuthorizationCode><![CDATA[AC_CODE]]></AuthorizationCode>
<AuthorizationCodeExpiredTime>1700003601</AuthorizationCodeExpiredTime>
<PreAuthCode><![CDATA[PRE]]></PreAuthCode>
</xml>`
	req := buildEncryptedReq(t, c, plain, "1700000001", "nonceB")

	notify, err := c.ParseNotify(req, nil)
	if err != nil {
		t.Fatal(err)
	}
	if notify.InfoType != "authorized" {
		t.Errorf("info_type mismatch: %q", notify.InfoType)
	}
	if notify.AuthorizerAppID != "wxAuthed" || notify.AuthorizationCode != "AC_CODE" {
		t.Errorf("unexpected: %+v", notify)
	}
	if notify.AuthorizationCodeExpiredTime != 1700003601 {
		t.Errorf("expire time mismatch: %d", notify.AuthorizationCodeExpiredTime)
	}
}

func TestParseNotify_BadSignature(t *testing.T) {
	c, _ := NewClient(testConfig())
	encrypted, _ := c.crypto.Encrypt([]byte(`<xml><InfoType>x</InfoType></xml>`))
	body := `<xml><Encrypt><![CDATA[` + encrypted + `]]></Encrypt></xml>`
	req := httptest.NewRequest(http.MethodPost,
		"/oplatform/notify?msg_signature=deadbeef&timestamp=1700&nonce=n1",
		strings.NewReader(body))
	if _, err := c.ParseNotify(req, nil); err == nil {
		t.Error("expected signature error")
	}
}

func TestParseNotify_RawBodyOverride(t *testing.T) {
	c, _ := NewClient(testConfig())
	plain := `<xml><AppId><![CDATA[wxcomp]]></AppId><InfoType><![CDATA[unauthorized]]></InfoType><AuthorizerAppid><![CDATA[wxAuthed]]></AuthorizerAppid></xml>`
	encrypted, _ := c.crypto.Encrypt([]byte(plain))
	sig := c.crypto.Signature("1700", "n1", encrypted)
	body := []byte(`<xml><Encrypt><![CDATA[` + encrypted + `]]></Encrypt></xml>`)
	req := httptest.NewRequest(http.MethodPost,
		"/oplatform/notify?msg_signature="+sig+"&timestamp=1700&nonce=n1",
		strings.NewReader("")) // empty body; we pass via rawBody
	notify, err := c.ParseNotify(req, body)
	if err != nil {
		t.Fatal(err)
	}
	if notify.InfoType != "unauthorized" || notify.AuthorizerAppID != "wxAuthed" {
		t.Errorf("unexpected: %+v", notify)
	}
}

func TestParseNotify_NotFoundSentinel(t *testing.T) {
	// When Store is empty, GetVerifyTicket returns ErrNotFound; we want to
	// ensure the sentinel is usable via errors.Is in downstream code.
	store := NewMemoryStore()
	_, err := store.GetVerifyTicket(context.Background())
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
	// Also verify xml.Unmarshal sanity for fail-fast payloads
	var env struct {
		XMLName xml.Name `xml:"xml"`
	}
	if err := xml.Unmarshal([]byte(`<xml/>`), &env); err != nil {
		t.Fatal(err)
	}
}
```

- [ ] **Step 12.2: Run failing test**

Run: `go test ./oplatform/ -run ParseNotify`
Expected: undefined `Client.ParseNotify`.

- [ ] **Step 12.3: Create `oplatform/notify.go`**

```go
package oplatform

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
)

// componentEnvelope 外层加密信封（和 offiaccount 结构相同）。
type componentEnvelope struct {
	XMLName    xml.Name `xml:"xml"`
	ToUserName string   `xml:"ToUserName"`
	Encrypt    string   `xml:"Encrypt"`
}

// componentInner 解密后的内层 XML。
type componentInner struct {
	XMLName                      xml.Name `xml:"xml"`
	AppID                        string   `xml:"AppId"`
	CreateTime                   int64    `xml:"CreateTime"`
	InfoType                     string   `xml:"InfoType"`
	ComponentVerifyTicket        string   `xml:"ComponentVerifyTicket"`
	AuthorizerAppID              string   `xml:"AuthorizerAppid"`
	AuthorizationCode            string   `xml:"AuthorizationCode"`
	AuthorizationCodeExpiredTime int64    `xml:"AuthorizationCodeExpiredTime"`
	PreAuthCode                  string   `xml:"PreAuthCode"`
}

// ParseNotify 解析开放平台第三方平台推送的回调。
//
//	r       - 原始 *http.Request；query 必须带 msg_signature/timestamp/nonce
//	rawBody - 可选：若调用方已经读过 r.Body，可以把原始字节从这里传入；
//	          若为 nil，本方法会从 r.Body 读取
//
// 成功返回 *ComponentNotify；当 InfoType == component_verify_ticket 时，
// SDK 会自动把 ticket 写入 Store，调用方无需再处理。
func (c *Client) ParseNotify(r *http.Request, rawBody []byte) (*ComponentNotify, error) {
	if r == nil {
		return nil, fmt.Errorf("oplatform: nil request")
	}
	q := r.URL.Query()

	if rawBody == nil {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, fmt.Errorf("oplatform: read body: %w", err)
		}
		_ = r.Body.Close()
		rawBody = body
	}

	var env componentEnvelope
	if err := xml.Unmarshal(rawBody, &env); err != nil {
		return nil, fmt.Errorf("oplatform: parse envelope: %w", err)
	}
	if env.Encrypt == "" {
		return nil, fmt.Errorf("oplatform: empty Encrypt field")
	}
	if !c.crypto.VerifySignature(q.Get("msg_signature"), q.Get("timestamp"), q.Get("nonce"), env.Encrypt) {
		return nil, fmt.Errorf("oplatform: msg_signature invalid")
	}
	plain, _, err := c.crypto.Decrypt(env.Encrypt)
	if err != nil {
		return nil, fmt.Errorf("oplatform: decrypt: %w", err)
	}

	var inner componentInner
	if err := xml.Unmarshal(plain, &inner); err != nil {
		return nil, fmt.Errorf("oplatform: parse inner xml: %w", err)
	}

	notify := &ComponentNotify{
		AppID:                        inner.AppID,
		CreateTime:                   inner.CreateTime,
		InfoType:                     inner.InfoType,
		ComponentVerifyTicket:        inner.ComponentVerifyTicket,
		AuthorizerAppID:              inner.AuthorizerAppID,
		AuthorizationCode:            inner.AuthorizationCode,
		AuthorizationCodeExpiredTime: inner.AuthorizationCodeExpiredTime,
		PreAuthCode:                  inner.PreAuthCode,
		Raw:                          plain,
	}

	// 自动写 verify_ticket
	if notify.InfoType == "component_verify_ticket" && notify.ComponentVerifyTicket != "" {
		if err := c.store.SetVerifyTicket(context.Background(), notify.ComponentVerifyTicket); err != nil {
			return nil, fmt.Errorf("oplatform: store set verify ticket: %w", err)
		}
	}

	return notify, nil
}
```

- [ ] **Step 12.4: Run tests**

Run: `go test ./oplatform/...`
Expected: all tests PASS (including the 5 new ParseNotify tests).

- [ ] **Step 12.5: Commit**

```bash
git add oplatform/notify.go oplatform/notify_test.go
git commit -m "feat(oplatform): add ParseNotify for component callback events

Parses the encrypted envelope, verifies signature, decrypts via
wxcrypto, and populates ComponentNotify for four InfoType branches
(component_verify_ticket / authorized / updateauthorized /
unauthorized). component_verify_ticket is auto-persisted to Store so
callers do not need to know that Weixin pushes it out-of-band.

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

## Task 13: oplatform QR Login (sub-project 6)

**Goal:** Implement `QRLoginClient` with `AuthorizeURL`, `Code2Token`, `RefreshToken`, `UserInfo`, `Auth`.

**Files:**
- Create: `oplatform/qrlogin.go`
- Create: `oplatform/qrlogin_test.go`

- [ ] **Step 13.1: Write failing test `qrlogin_test.go`**

```go
package oplatform

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

func newQRTestClient(baseURL string) *QRLoginClient {
	return NewQRLoginClient("wxqr", "qrsecret",
		WithQRLoginHTTP(utils.NewHTTP(baseURL, utils.WithTimeout(time.Second*3))))
}

func TestQRLoginClient_AuthorizeURL(t *testing.T) {
	q := NewQRLoginClient("wxqr", "sec")
	u := q.AuthorizeURL("https://example.com/cb", "snsapi_login", "state1")
	if !strings.Contains(u, "appid=wxqr") {
		t.Errorf("appid: %s", u)
	}
	if !strings.Contains(u, "scope=snsapi_login") {
		t.Errorf("scope: %s", u)
	}
	if !strings.Contains(u, "state=state1") {
		t.Errorf("state: %s", u)
	}
	if !strings.Contains(u, "redirect_uri=https%3A%2F%2Fexample.com%2Fcb") {
		t.Errorf("redirect_uri: %s", u)
	}
	if !strings.HasSuffix(u, "#wechat_redirect") {
		t.Errorf("missing fragment: %s", u)
	}
}

func TestQRLoginClient_Code2Token(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/sns/oauth2/access_token") {
			t.Errorf("path: %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("appid") != "wxqr" || q.Get("secret") != "qrsecret" || q.Get("code") != "CODE" || q.Get("grant_type") != "authorization_code" {
			t.Errorf("unexpected query: %v", q)
		}
		_, _ = w.Write([]byte(`{"access_token":"A","expires_in":7200,"refresh_token":"R","openid":"O","scope":"snsapi_login","unionid":"U"}`))
	}))
	defer srv.Close()
	q := newQRTestClient(srv.URL)

	tok, err := q.Code2Token(context.Background(), "CODE")
	if err != nil {
		t.Fatal(err)
	}
	if tok.AccessToken != "A" || tok.OpenID != "O" || tok.UnionID != "U" {
		t.Errorf("unexpected: %+v", tok)
	}
}

func TestQRLoginClient_RefreshToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/sns/oauth2/refresh_token") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"access_token":"A2","expires_in":7200,"refresh_token":"R2","openid":"O","scope":"snsapi_login"}`))
	}))
	defer srv.Close()
	q := newQRTestClient(srv.URL)

	tok, err := q.RefreshToken(context.Background(), "RX")
	if err != nil {
		t.Fatal(err)
	}
	if tok.AccessToken != "A2" || tok.RefreshToken != "R2" {
		t.Errorf("unexpected: %+v", tok)
	}
}

func TestQRLoginClient_UserInfo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/sns/userinfo") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"openid":"O","nickname":"N","sex":1,"country":"CN","unionid":"U"}`))
	}))
	defer srv.Close()
	q := newQRTestClient(srv.URL)

	info, err := q.UserInfo(context.Background(), "TOK", "O")
	if err != nil {
		t.Fatal(err)
	}
	if info.Nickname != "N" || info.Country != "CN" {
		t.Errorf("unexpected: %+v", info)
	}
}

func TestQRLoginClient_Auth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/sns/auth") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()
	q := newQRTestClient(srv.URL)

	if err := q.Auth(context.Background(), "TOK", "O"); err != nil {
		t.Fatal(err)
	}
}

func TestQRLoginClient_Code2Token_Errcode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":40029,"errmsg":"invalid code"}`))
	}))
	defer srv.Close()
	q := newQRTestClient(srv.URL)
	if _, err := q.Code2Token(context.Background(), "BAD"); err == nil {
		t.Error("expected error")
	}
}
```

- [ ] **Step 13.2: Run failing test**

Run: `go test ./oplatform/ -run QRLogin`
Expected: undefined `QRLoginClient`, `NewQRLoginClient`, `WithQRLoginHTTP`.

- [ ] **Step 13.3: Create `oplatform/qrlogin.go`**

```go
package oplatform

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

// QRLoginClient 提供微信开放平台"网站应用微信登录"能力 (snsapi_login)。
// 与 component Client 完全独立：无 Store 依赖、无 token 缓存，每次换 token
// 都直接调微信接口。
type QRLoginClient struct {
	appID     string
	appSecret string
	http      *utils.HTTP
}

type QRLoginOption func(*QRLoginClient)

// WithQRLoginHTTP 注入自定义 HTTP（测试常用）。
func WithQRLoginHTTP(h *utils.HTTP) QRLoginOption {
	return func(q *QRLoginClient) {
		if h != nil {
			q.http = h
		}
	}
}

// NewQRLoginClient 构造一个 QR Login 客户端。
func NewQRLoginClient(appID, appSecret string, opts ...QRLoginOption) *QRLoginClient {
	q := &QRLoginClient{
		appID:     appID,
		appSecret: appSecret,
		http:      utils.NewHTTP("https://api.weixin.qq.com", utils.WithTimeout(time.Second*30)),
	}
	for _, o := range opts {
		o(q)
	}
	return q
}

// AuthorizeURL 构造开放平台扫码登录跳转 URL。
//
//	scope - snsapi_login / snsapi_base / snsapi_userinfo
//	state - CSRF 防护 token
func (q *QRLoginClient) AuthorizeURL(redirectURI, scope, state string) string {
	v := url.Values{
		"appid":         {q.appID},
		"redirect_uri":  {redirectURI},
		"response_type": {"code"},
		"scope":         {scope},
		"state":         {state},
	}
	return "https://open.weixin.qq.com/connect/qrconnect?" + v.Encode() + "#wechat_redirect"
}

// Code2Token 用扫码登录 code 换取 access_token。
func (q *QRLoginClient) Code2Token(ctx context.Context, code string) (*QRLoginToken, error) {
	ctx = touchContext(ctx)
	if code == "" {
		return nil, fmt.Errorf("oplatform: code is required")
	}
	v := url.Values{
		"appid":      {q.appID},
		"secret":     {q.appSecret},
		"code":       {code},
		"grant_type": {"authorization_code"},
	}
	out := &QRLoginToken{}
	if err := q.http.Get(ctx, "/sns/oauth2/access_token", v, out); err != nil {
		return nil, fmt.Errorf("oplatform: qrlogin code2token: %w", err)
	}
	if err := checkWeixinErr(out.ErrCode, out.ErrMsg); err != nil {
		return nil, err
	}
	return out, nil
}

// RefreshToken 用 refresh_token 换取新的 access_token。
func (q *QRLoginClient) RefreshToken(ctx context.Context, refreshToken string) (*QRLoginToken, error) {
	ctx = touchContext(ctx)
	if refreshToken == "" {
		return nil, fmt.Errorf("oplatform: refresh_token is required")
	}
	v := url.Values{
		"appid":         {q.appID},
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
	}
	out := &QRLoginToken{}
	if err := q.http.Get(ctx, "/sns/oauth2/refresh_token", v, out); err != nil {
		return nil, fmt.Errorf("oplatform: qrlogin refresh: %w", err)
	}
	if err := checkWeixinErr(out.ErrCode, out.ErrMsg); err != nil {
		return nil, err
	}
	return out, nil
}

// UserInfo 拉取用户信息。仅 snsapi_userinfo scope 可用。
func (q *QRLoginClient) UserInfo(ctx context.Context, accessToken, openID string) (*QRLoginUserInfo, error) {
	ctx = touchContext(ctx)
	v := url.Values{
		"access_token": {accessToken},
		"openid":       {openID},
		"lang":         {"zh_CN"},
	}
	out := &QRLoginUserInfo{}
	if err := q.http.Get(ctx, "/sns/userinfo", v, out); err != nil {
		return nil, fmt.Errorf("oplatform: qrlogin userinfo: %w", err)
	}
	if err := checkWeixinErr(out.ErrCode, out.ErrMsg); err != nil {
		return nil, err
	}
	return out, nil
}

// Auth 检查 access_token 是否有效。
func (q *QRLoginClient) Auth(ctx context.Context, accessToken, openID string) error {
	ctx = touchContext(ctx)
	v := url.Values{"access_token": {accessToken}, "openid": {openID}}
	out := &qrloginAuthResp{}
	if err := q.http.Get(ctx, "/sns/auth", v, out); err != nil {
		return fmt.Errorf("oplatform: qrlogin auth: %w", err)
	}
	return checkWeixinErr(out.ErrCode, out.ErrMsg)
}
```

- [ ] **Step 13.4: Run tests**

Run: `go test ./oplatform/...`
Expected: all tests PASS including 6 new QR login tests.

- [ ] **Step 13.5: Full repo test with race detector**

Run: `go test -race ./...`
Expected: PASS.

- [ ] **Step 13.6: Commit**

```bash
git add oplatform/qrlogin.go oplatform/qrlogin_test.go
git commit -m "feat(oplatform): add QR Login client (open-platform snsapi_login)

Five-method client for WeChat Open Platform website QR login:
AuthorizeURL, Code2Token, RefreshToken, UserInfo, Auth. Independent
of component flow — no Store, no token cache, no crypto.

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

## Task 14: oplatform example — callback server + QR login demo

**Goal:** Ship a compilable-only (not runnable in CI) example showing how to wire everything together. Must `go build`.

**Files:**
- Create: `oplatform/example/main.go`

- [ ] **Step 14.1: Create `oplatform/example/main.go`**

```go
// Example: running a WeChat Open Platform third-party component callback
// server plus a QR login handler. This file must compile; it is not
// exercised in CI. Replace ComponentAppID / AppSecret / Token / AESKey
// with real values before running.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	mini_program "github.com/godrealms/go-wechat-sdk/mini-program"
	"github.com/godrealms/go-wechat-sdk/offiaccount"
	"github.com/godrealms/go-wechat-sdk/oplatform"
)

func main() {
	op, err := oplatform.NewClient(oplatform.Config{
		ComponentAppID:     "wxcompXXXX",
		ComponentAppSecret: "componentsecretXXXX",
		Token:              "callbacktoken",
		EncodingAESKey:     "0123456789ABCDEF0123456789ABCDEF0123456789A", // 43 chars
	})
	if err != nil {
		log.Fatal(err)
	}

	// 1) Component callback endpoint — receives verify_ticket pushes
	//    and authorization events. SDK auto-persists verify_ticket
	//    into the Store.
	http.HandleFunc("/oplatform/callback", func(w http.ResponseWriter, r *http.Request) {
		notify, err := op.ParseNotify(r, nil)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		switch notify.InfoType {
		case "component_verify_ticket":
			// already persisted by the SDK
			log.Printf("verify_ticket refreshed")
		case "authorized":
			if _, err := op.QueryAuth(r.Context(), notify.AuthorizationCode); err != nil {
				log.Printf("QueryAuth failed: %v", err)
			}
		case "updateauthorized":
			log.Printf("authorizer %s updated", notify.AuthorizerAppID)
		case "unauthorized":
			_ = op.Store().DeleteAuthorizer(r.Context(), notify.AuthorizerAppID)
		}
		// Weixin expects plain "success"
		_, _ = w.Write([]byte("success"))
	})

	// 2) Delegated offiaccount call — token flows through oplatform
	http.HandleFunc("/call/menu/{appid}", func(w http.ResponseWriter, r *http.Request) {
		appid := r.PathValue("appid")
		auth := op.Authorizer(appid)
		off := auth.OffiaccountClient()
		tok, err := off.AccessTokenE(r.Context())
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		fmt.Fprintf(w, "authorized token: %s (for %s)", tok, appid)
	})

	// 3) Delegated mini-program call
	http.HandleFunc("/mp/token/{appid}", func(w http.ResponseWriter, r *http.Request) {
		appid := r.PathValue("appid")
		auth := op.Authorizer(appid)
		mp, err := auth.MiniProgramClient()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		tok, err := mp.AccessToken(r.Context())
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		fmt.Fprintf(w, "mp token: %s", tok)
	})

	// 4) QR login flow (open-platform website login)
	qr := oplatform.NewQRLoginClient("wxqrXXXX", "qrsecretXXXX")
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		u := qr.AuthorizeURL("https://example.com/login/cb", "snsapi_login", "csrf_token")
		http.Redirect(w, r, u, http.StatusFound)
	})
	http.HandleFunc("/login/cb", func(w http.ResponseWriter, r *http.Request) {
		tok, err := qr.Code2Token(r.Context(), r.URL.Query().Get("code"))
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		info, err := qr.UserInfo(r.Context(), tok.AccessToken, tok.OpenID)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		fmt.Fprintf(w, "welcome %s (%s)", info.Nickname, info.UnionID)
	})

	// Periodic token warm-up (every 1h is typical)
	go func() {
		if err := op.RefreshAll(context.Background()); err != nil {
			log.Printf("RefreshAll: %v", err)
		}
	}()

	// Silence unused-import warning for mini_program/offiaccount if reader
	// strips this file of the HTTP handlers above.
	_ = mini_program.Config{}
	_ = offiaccount.Config{}

	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

- [ ] **Step 14.2: Verify it compiles**

Run: `go build ./oplatform/example/...`
Expected: clean build.

- [ ] **Step 14.3: Run full test sweep once more**

Run: `go build ./... && go test ./...`
Expected: all green.

- [ ] **Step 14.4: Commit**

```bash
git add oplatform/example/main.go
git commit -m "docs(oplatform): add compilable example (callback + qrlogin + delegated calls)

Shows the typical wiring for a third-party platform service:
verify_ticket callback, authorized-event handler, delegated
offiaccount/mini-program calls via AuthorizerClient, and a
website QR login flow. Compile-only; not exercised in CI.

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

## Task 15: Update README + docs/README status tables

**Goal:** Flip `oplatform` from 🚧 to ✅ with a clear note on what is in-scope this round.

**Files:**
- Modify: `README.md`
- Modify: `docs/README.md`

- [ ] **Step 15.1: Update `README.md`**

Find this line:

```
- 🌐 开放平台 (`oplatform`) 🚧：目录占位，尚未实现
```

Replace with:

```
- 🌐 开放平台 (`oplatform`) ✅：第三方平台授权底座（verify_ticket 回调、component_access_token、pre_auth_code、authorization_code 换 authorizer token、授权事件解析）+ authorizer 代调用框架（通过 TokenSource 注入到 offiaccount / mini-program，token 自动来自开放平台）+ 网站应用扫码登录 (snsapi_login)。代小程序开发/运营管理、快速注册等子模块后续实现
```

- [ ] **Step 15.2: Update `docs/README.md`**

Find this line in the status table:

```
| `oplatform` — 开放平台 | 🚧 | 尚未实现 |
```

Replace with:

```
| `oplatform` — 开放平台 | ✅ | 第三方平台授权底座 + authorizer 代调用框架 + 网站扫码登录；代小程序发布/快速注册等待实现 |
```

- [ ] **Step 15.3: Commit**

```bash
git add README.md docs/README.md
git commit -m "docs: flip oplatform to ✅ with in-scope note

Covers sub-projects 1 (component auth), 2 (authorizer delegation via
TokenSource) and 6 (QR login). Remaining oplatform sub-projects
(代小程序 dev/ops, quick registration) tracked for future rounds.

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

## Task 16: Final verification sweep

**Goal:** One last confirmation that nothing is broken.

- [ ] **Step 16.1: Go vet**

Run: `go vet ./...`
Expected: no output.

- [ ] **Step 16.2: Race-enabled full test suite**

Run: `go test -race ./...`
Expected: all PASS.

- [ ] **Step 16.3: Build everything**

Run: `go build ./...`
Expected: clean.

- [ ] **Step 16.4: Verify no dangling references**

Run:
```bash
grep -rn "tokenFromCrypto" .
grep -rn "NewMsgCrypto" offiaccount/ oplatform/ utils/wxcrypto/
```

Expected:
- `tokenFromCrypto` — no matches anywhere (deleted in Task 2).
- `NewMsgCrypto` — only in `offiaccount/crypto.go` (shim) and possibly test files; should NOT appear in `utils/wxcrypto/`.

- [ ] **Step 16.5: Git log sanity check**

Run: `git log --oneline -20`
Expected: commits from Task 1 through Task 15 in order, each describing one atomic change.

No additional commit at this step — verification only. If any step fails, return to the relevant task to diagnose and fix.

---

## Coverage Map (self-review)

| Spec section | Task |
|---|---|
| §2.1 file layout | Tasks 5, 6, 7, 8, 9, 10, 11, 12, 13, 14 |
| §2.2 basic rules (NewClient no I/O, shared utils.HTTP, ctx everywhere) | Tasks 7, 10, 13 |
| §3 Store contract + MemoryStore | Task 6 |
| §4.1 Config & Client | Task 7 |
| §4.2 ParseNotify + auto-persist verify_ticket | Task 12 |
| §4.3 Component token lazy + RefreshComponentToken | Task 7 |
| §4.4 PreAuthCode + AuthorizeURL / MobileAuthorizeURL | Task 8 |
| §4.5 QueryAuth with Store write-through | Task 9 |
| §4.6 GetAuthorizerInfo / Option / List | Task 9 |
| §5.2 TokenSource interfaces in offiaccount & mini_program | Tasks 3, 4 |
| §5.3 offiaccount WithTokenSource + AccessTokenE delegation | Task 3 |
| §5.4 AuthorizerClient + OffiaccountClient / MiniProgramClient | Tasks 10, 11 |
| §5.5 Refresh semantics + ErrAuthorizerRevoked | Task 10 |
| §5.5 RefreshAll | Task 11 |
| §5.6 per-appid sync.Map + componentMu | Tasks 7, 10 |
| §6 QR Login 5 methods | Task 13 |
| §7 utils/wxcrypto extraction + offiaccount shim | Tasks 1, 2 |
| §8 errors (WeixinError, ErrNotFound, ErrAuthorizerRevoked, ErrVerifyTicketMissing) | Task 5 |
| §9 test matrix | Tasks 1, 2, 3, 4, 6, 7, 8, 9, 10, 11, 12, 13 |
| §10 compatibility (no offiaccount / mini-program API breaks) | Tasks 2, 3, 4 + Task 16 verification |
| §11 delivery list | Tasks 1-15 |
| §12 non-goals | Not implemented — verified by scope |
| §13 open questions | None |

All spec sections covered. No placeholders remain. All types (`Config`, `Client`, `Store`, `AuthorizerTokens`, `AuthorizerClient`, `QRLoginClient`, `ComponentNotify`, etc.) are defined before first use. All method names stay consistent across tasks:

- `ComponentAccessToken` / `RefreshComponentToken` (Task 7)
- `PreAuthCode` / `AuthorizeURL` / `MobileAuthorizeURL` (Task 8)
- `QueryAuth` / `GetAuthorizerInfo` / `GetAuthorizerOption` / `SetAuthorizerOption` / `GetAuthorizerList` (Task 9)
- `AuthorizerClient.AccessToken` / `AuthorizerClient.Refresh` / `AuthorizerClient.OffiaccountClient` / `AuthorizerClient.MiniProgramClient` (Tasks 10, 11)
- `Client.Authorizer(appID)` / `Client.RefreshAll` (Tasks 10, 11)
- `Client.ParseNotify` (Task 12)
- `QRLoginClient` + `NewQRLoginClient` + `WithQRLoginHTTP` (Task 13)
