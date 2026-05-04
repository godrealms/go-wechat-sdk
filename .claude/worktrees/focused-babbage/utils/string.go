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
	const maxByte = 256 - (256 % len(charset)) // 248, 消除取模偏差
	result := make([]byte, length)
	buf := make([]byte, length+(length/4)) // 额外缓冲区用于被拒绝的字节

	i := 0
	for i < length {
		if _, err := io.ReadFull(cryptoRand.Reader, buf); err != nil {
			return RandomString(length, "")
		}
		for _, b := range buf {
			if i >= length {
				break
			}
			if int(b) < maxByte {
				result[i] = charset[int(b)%len(charset)]
				i++
			}
		}
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
