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

// GenerateHashBasedString 基于 crypto/rand 生成哈希散列字符串
func GenerateHashBasedString(length int) string {
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
