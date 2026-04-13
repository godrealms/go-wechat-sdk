package utils

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
)

// SignSHA256WithRSA 通过私钥对字符串以 SHA256WithRSA 算法生成签名信息
func SignSHA256WithRSA(source string, privateKey *rsa.PrivateKey) (signature string, err error) {
	if privateKey == nil {
		return "", fmt.Errorf("private key should not be nil")
	}
	h := crypto.SHA256.New()
	if _, err = h.Write([]byte(source)); err != nil {
		return "", fmt.Errorf("hash write failed: %w", err)
	}
	hashed := h.Sum(nil)
	signatureByte, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signatureByte), nil
}

// VerifySHA256WithRSA 通过公钥对 base64 编码的签名做 SHA256WithRSA 验证。
// 主要用于校验微信支付平台返回的响应签名以及回调通知签名。
func VerifySHA256WithRSA(source string, signature string, publicKey *rsa.PublicKey) error {
	if publicKey == nil {
		return fmt.Errorf("public key should not be nil")
	}
	sig, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("decode signature failed: %w", err)
	}
	h := crypto.SHA256.New()
	if _, err = h.Write([]byte(source)); err != nil {
		return fmt.Errorf("hash write failed: %w", err)
	}
	return rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, h.Sum(nil), sig)
}
