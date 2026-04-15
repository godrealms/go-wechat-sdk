package wxcrypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"crypto/subtle"
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

// errOpaque is the single error returned for ANY post-AES-decryption failure
// in this package. We deliberately do not distinguish padding errors, length
// errors, or appid mismatch — that distinction is a padding-oracle leak (audit C4).
var errOpaque = errors.New("wxcrypto: decrypt failed")

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
		return nil, "", errOpaque
	}
	block, err := aes.NewCipher(m.aesKey)
	if err != nil {
		return nil, "", errOpaque
	}
	if len(cipherText)%aes.BlockSize != 0 {
		return nil, "", errOpaque
	}
	buf := make([]byte, len(cipherText))
	mode := cipher.NewCBCDecrypter(block, m.iv)
	mode.CryptBlocks(buf, cipherText)

	buf, err = pkcs7Unpad(buf, aes.BlockSize)
	if err != nil {
		return nil, "", errOpaque
	}
	if len(buf) < 20 {
		return nil, "", errOpaque
	}
	buf = buf[16:]
	msgLen := binary.BigEndian.Uint32(buf[:4])
	if int(msgLen)+4 > len(buf) {
		return nil, "", errOpaque
	}
	msg := buf[4 : 4+msgLen]
	gotAppid := string(buf[4+msgLen:])
	if m.appid != "" && gotAppid != m.appid {
		return nil, "", errOpaque
	}
	return msg, gotAppid, nil
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

// pkcs7Unpad removes PKCS#7 padding using crypto/subtle so that the work
// done is independent of whether the padding is valid. Audit C4.
func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	n := len(data)
	if n == 0 || n%blockSize != 0 {
		return nil, errOpaque
	}
	pad := int(data[n-1])
	// pad must be in [1, blockSize].
	valid := subtle.ConstantTimeLessOrEq(1, pad)
	valid &= subtle.ConstantTimeLessOrEq(pad, blockSize)
	// Check that the last `pad` bytes all equal `pad`. We always inspect the
	// last `blockSize` bytes; positions outside the padding region get a
	// pass via the (NOT inPad) branch so any byte there is ignored.
	for i := 0; i < blockSize; i++ {
		b := data[n-1-i]
		// inPad == 1 iff (i+1) <= pad, i.e. this byte is in the padding region.
		inPad := subtle.ConstantTimeLessOrEq(i+1, pad)
		match := subtle.ConstantTimeByteEq(b, byte(pad))
		// valid &= (NOT inPad) | match — equivalent to: inPad implies match
		valid &= 1 ^ (inPad & (1 ^ match))
	}
	if valid != 1 {
		return nil, errOpaque
	}
	return data[:n-pad], nil
}
