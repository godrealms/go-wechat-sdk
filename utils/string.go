package utils

import (
	cryptoRand "crypto/rand"
	"encoding/binary"
	"io"
	"math/rand"
	"strings"
)

// RandomString 生成长度为 length 的随机字符串并附加前缀。
// 内部使用 crypto/rand 以保证不可预测性，可直接用于 nonceStr。
func RandomString(length int, prefix string) string {
	if length <= 0 {
		length = 6
	}
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghjkmnopqrstuvwxyz23456789"

	var sb strings.Builder
	sb.Grow(length)

	randomBytes := make([]byte, length)
	if _, err := io.ReadFull(cryptoRand.Reader, randomBytes); err != nil {
		// 极端情况下退回到 math/rand，但仍然遵守长度约束
		fallback := newFallbackRand()
		for i := 0; i < length; i++ {
			sb.WriteByte(charset[fallback.Intn(len(charset))])
		}
		return prefix + sb.String()
	}

	for _, b := range randomBytes {
		sb.WriteByte(charset[int(b)%len(charset)])
	}
	return prefix + sb.String()
}

// GenerateNonceString 使用 crypto/rand 生成长度为 length 的随机字母数字串，
// 适合用作微信支付 nonce_str / 请求 ID 等场景。length <= 0 时退化为 32。
//
// 与 RandomString 的区别：本函数使用包含 0/O、1/I 等易混淆字符的更宽 charset
// (A-Za-z0-9, 共 62 个字符)，结果长度严格等于 length，不附加前缀。
func GenerateNonceString(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	if length <= 0 {
		length = 32
	}
	result := make([]byte, length)
	randomBytes := make([]byte, length)

	if _, err := io.ReadFull(cryptoRand.Reader, randomBytes); err != nil {
		return RandomString(length, "")
	}

	for i, b := range randomBytes {
		result[i] = charset[b%byte(len(charset))]
	}
	return string(result)
}

// GenerateHashBasedString 是 GenerateNonceString 的旧名。该名字对实际行为
// (随机字母数字串而非任何形式的"哈希")存在误导，将在未来版本移除。
//
// Deprecated: 请改用 GenerateNonceString。
func GenerateHashBasedString(length int) string {
	return GenerateNonceString(length)
}

// ShuffleString 使用 crypto/rand 打乱字符串
func ShuffleString(s string) string {
	runes := []rune(s)
	n := len(runes)
	if n < 2 {
		return s
	}

	for i := n - 1; i > 0; i-- {
		j := cryptoIntn(i + 1)
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// cryptoIntn 使用 crypto/rand 生成 [0, n) 范围内的随机整数
func cryptoIntn(n int) int {
	if n <= 0 {
		return 0
	}
	var buf [8]byte
	if _, err := io.ReadFull(cryptoRand.Reader, buf[:]); err != nil {
		return newFallbackRand().Intn(n)
	}
	return int(binary.BigEndian.Uint64(buf[:]) % uint64(n))
}

// newFallbackRand 仅在 crypto/rand 失败时作为兜底使用。
// Go 1.20+ math/rand 全局已自动播种，所以这里可以放心使用 rand.Intn 风格的兜底。
func newFallbackRand() *rand.Rand {
	return rand.New(rand.NewSource(rand.Int63()))
}
