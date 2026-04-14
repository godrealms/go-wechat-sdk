package mini_program

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
)

// DecryptUserData 解密 wx.getUserInfo / wx.getPhoneNumber 返回的 encryptedData。
//
//	sessionKey    - 来自 Code2Session 的 session_key（base64）
//	encryptedData - 前端传回的 encryptedData（base64）
//	iv            - 前端传回的 iv（base64）
//	result        - 用于 JSON Unmarshal 的目标，传 nil 则只返回原始字节
//
// 返回解密后的 JSON 字节。
func DecryptUserData(sessionKey, encryptedData, iv string, result any) ([]byte, error) {
	key, err := base64.StdEncoding.DecodeString(sessionKey)
	if err != nil {
		return nil, fmt.Errorf("decode sessionKey: %w", err)
	}
	cipherBytes, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("decode encryptedData: %w", err)
	}
	ivBytes, err := base64.StdEncoding.DecodeString(iv)
	if err != nil {
		return nil, fmt.Errorf("decode iv: %w", err)
	}
	if len(key) != 16 {
		return nil, fmt.Errorf("mini_program: sessionKey must be 16 bytes after base64 decode, got %d", len(key))
	}
	if len(ivBytes) != 16 {
		return nil, fmt.Errorf("mini_program: iv must be 16 bytes after base64 decode, got %d", len(ivBytes))
	}
	if len(cipherBytes)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("mini_program: ciphertext not aligned to blocksize")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	plain := make([]byte, len(cipherBytes))
	cipher.NewCBCDecrypter(block, ivBytes).CryptBlocks(plain, cipherBytes)
	plain, err = pkcs7Unpad(plain, aes.BlockSize)
	if err != nil {
		return nil, err
	}
	if result != nil {
		if err := json.Unmarshal(plain, result); err != nil {
			return plain, fmt.Errorf("unmarshal plaintext: %w", err)
		}
	}
	return plain, nil
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
