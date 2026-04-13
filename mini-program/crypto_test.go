package mini_program

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"testing"
)

func TestDecryptUserData_Roundtrip(t *testing.T) {
	// 构造 16 字节 key 和 iv
	key := []byte("0123456789abcdef")
	iv := []byte("fedcba9876543210")
	plain := []byte(`{"openId":"o1","nickName":"Jie"}`)

	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	// PKCS#7 pad
	pad := aes.BlockSize - len(plain)%aes.BlockSize
	padded := append(plain, bytes.Repeat([]byte{byte(pad)}, pad)...)
	encrypted := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(encrypted, padded)

	var out struct {
		OpenId   string `json:"openId"`
		NickName string `json:"nickName"`
	}
	got, err := DecryptUserData(
		base64.StdEncoding.EncodeToString(key),
		base64.StdEncoding.EncodeToString(encrypted),
		base64.StdEncoding.EncodeToString(iv),
		&out,
	)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(plain) {
		t.Errorf("plaintext mismatch: %q vs %q", got, plain)
	}
	if out.OpenId != "o1" || out.NickName != "Jie" {
		t.Errorf("unmarshalled = %+v", out)
	}
}

func TestDecryptUserData_BadSessionKey(t *testing.T) {
	_, err := DecryptUserData("!!!bad!!!", "aGVsbG8=", "aGVsbG8=", nil)
	if err == nil {
		t.Error("expected error")
	}
}
