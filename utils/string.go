package utils

import (
	cryptoRand "crypto/rand"
	"io"
	"math/rand"
	"strings"
	"time"
)

func RandomString(length int, prefix string) string {
	if length <= 0 {
		length = 6
	}
	charset := "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghjkmnopqrstuvwxyz23456789"

	var sb strings.Builder
	sb.Grow(length)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		randomIndex := r.Intn(len(charset))
		sb.WriteByte(charset[randomIndex])
	}

	return prefix + sb.String()
}

// GenerateHashBasedString 哈稀散字符串
func GenerateHashBasedString(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	randomBytes := make([]byte, length)

	// 使用 crypto/rand 生成随机字节
	if _, err := io.ReadFull(cryptoRand.Reader, randomBytes); err != nil {
		return RandomString(length, "")
	}

	// 根据随机字节选择字符集中的字符
	for i, b := range randomBytes {
		result[i] = charset[b%byte(len(charset))]
	}

	return string(result)
}

// ShuffleString 打乱字符串
func ShuffleString(s string) string {
	// 将字符串转换为字符切片
	runes := []rune(s)
	n := len(runes)

	// 创建一个新的随机数生成器
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Fisher-Yates 洗牌算法
	for i := n - 1; i > 0; i-- {
		j := r.Intn(i + 1)                      // 生成 [0, i] 范围内的随机数
		runes[i], runes[j] = runes[j], runes[i] // 交换
	}

	return string(runes) // 转换回字符串
}
