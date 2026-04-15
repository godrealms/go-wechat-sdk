package mini_program

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
)

// ErrDecrypt is the single error returned for ANY failure up to and including
// AES-CBC decrypt + PKCS#7 unpad in DecryptUserData. We deliberately do not
// distinguish base64 errors, key-length errors, IV-length errors, block-size
// errors, or padding errors: any such distinction is a padding-oracle leak
// (audit C4). JSON unmarshal errors after a successful decrypt are a separate,
// non-secret class and are returned wrapped.
var ErrDecrypt = errors.New("mini_program: decrypt failed")

// DecryptUserData 解密 wx.getUserInfo / wx.getPhoneNumber 返回的 encryptedData。
//
//	sessionKey    - 来自 Code2Session 的 session_key（base64）
//	encryptedData - 前端传回的 encryptedData（base64）
//	iv            - 前端传回的 iv（base64）
//	result        - 用于 JSON Unmarshal 的目标，传 nil 则只返回原始字节
//
// 返回解密后的 JSON 字节。任何解密失败都会返回同一个 ErrDecrypt，不会区分
// "sessionKey 长度错"、"padding 非法"等具体原因，以避免 padding oracle。
func DecryptUserData(sessionKey, encryptedData, iv string, result any) ([]byte, error) {
	key, err := base64.StdEncoding.DecodeString(sessionKey)
	if err != nil {
		return nil, ErrDecrypt
	}
	cipherBytes, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, ErrDecrypt
	}
	ivBytes, err := base64.StdEncoding.DecodeString(iv)
	if err != nil {
		return nil, ErrDecrypt
	}
	if len(key) != 16 || len(ivBytes) != 16 {
		return nil, ErrDecrypt
	}
	if len(cipherBytes) == 0 || len(cipherBytes)%aes.BlockSize != 0 {
		return nil, ErrDecrypt
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, ErrDecrypt
	}
	plain := make([]byte, len(cipherBytes))
	cipher.NewCBCDecrypter(block, ivBytes).CryptBlocks(plain, cipherBytes)
	plain, ok := pkcs7Unpad(plain, aes.BlockSize)
	if !ok {
		return nil, ErrDecrypt
	}
	if result != nil {
		if err := json.Unmarshal(plain, result); err != nil {
			return plain, fmt.Errorf("mini_program: unmarshal plaintext: %w", err)
		}
	}
	return plain, nil
}

// pkcs7Unpad removes PKCS#7 padding in constant time. It returns (data, true)
// on success and (nil, false) on any malformed padding. The work done is
// independent of whether the padding is valid, so an attacker timing this
// function cannot learn anything about the plaintext. Audit C4.
func pkcs7Unpad(data []byte, blockSize int) ([]byte, bool) {
	n := len(data)
	if n == 0 || n%blockSize != 0 {
		return nil, false
	}
	pad := int(data[n-1])
	valid := subtle.ConstantTimeLessOrEq(1, pad)
	valid &= subtle.ConstantTimeLessOrEq(pad, blockSize)
	for i := 0; i < blockSize; i++ {
		b := data[n-1-i]
		inPad := subtle.ConstantTimeLessOrEq(i+1, pad)
		match := subtle.ConstantTimeByteEq(b, byte(pad))
		// inPad implies match
		valid &= 1 ^ (inPad & (1 ^ match))
	}
	if valid != 1 {
		return nil, false
	}
	return data[:n-pad], true
}
