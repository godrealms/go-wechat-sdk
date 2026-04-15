package mini_program

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
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

// TestDecryptUserData_OpaqueErrors verifies that every pre-plaintext failure
// path returns the single ErrDecrypt sentinel — no oracle. Audit C4.
func TestDecryptUserData_OpaqueErrors(t *testing.T) {
	key := []byte("0123456789abcdef")
	iv := []byte("fedcba9876543210")
	plain := []byte(`{"openId":"o1"}`)

	block, _ := aes.NewCipher(key)
	pad := aes.BlockSize - len(plain)%aes.BlockSize
	padded := append(plain, bytes.Repeat([]byte{byte(pad)}, pad)...)
	encrypted := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(encrypted, padded)

	keyB64 := base64.StdEncoding.EncodeToString(key)
	ivB64 := base64.StdEncoding.EncodeToString(iv)
	encB64 := base64.StdEncoding.EncodeToString(encrypted)

	// Tampered ciphertext (flip one byte in the last block → bad padding).
	tamperedBytes := append([]byte(nil), encrypted...)
	tamperedBytes[len(tamperedBytes)-1] ^= 0xff
	tamperedB64 := base64.StdEncoding.EncodeToString(tamperedBytes)

	cases := []struct {
		name                       string
		sessionKey, encrypted, iv  string
	}{
		{"bad base64 sessionKey", "!!!bad!!!", encB64, ivB64},
		{"bad base64 encryptedData", keyB64, "!!!bad!!!", ivB64},
		{"bad base64 iv", keyB64, encB64, "!!!bad!!!"},
		{"wrong sessionKey length", base64.StdEncoding.EncodeToString([]byte("short")), encB64, ivB64},
		{"wrong iv length", keyB64, encB64, base64.StdEncoding.EncodeToString([]byte("short"))},
		{"ciphertext not block-aligned", keyB64, base64.StdEncoding.EncodeToString([]byte("abc")), ivB64},
		{"empty ciphertext", keyB64, "", ivB64},
		{"tampered ciphertext / bad padding", keyB64, tamperedB64, ivB64},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := DecryptUserData(tc.sessionKey, tc.encrypted, tc.iv, nil)
			if !errors.Is(err, ErrDecrypt) {
				t.Fatalf("expected ErrDecrypt, got %v", err)
			}
		})
	}
}

// TestDecryptUserData_UnmarshalErrorDistinct ensures JSON shape errors stay a
// separate error class (they are not a crypto-oracle risk).
func TestDecryptUserData_UnmarshalErrorDistinct(t *testing.T) {
	key := []byte("0123456789abcdef")
	iv := []byte("fedcba9876543210")
	plain := []byte(`not-json-at-all`)

	block, _ := aes.NewCipher(key)
	pad := aes.BlockSize - len(plain)%aes.BlockSize
	padded := append(plain, bytes.Repeat([]byte{byte(pad)}, pad)...)
	encrypted := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(encrypted, padded)

	var out struct{ X string }
	_, err := DecryptUserData(
		base64.StdEncoding.EncodeToString(key),
		base64.StdEncoding.EncodeToString(encrypted),
		base64.StdEncoding.EncodeToString(iv),
		&out,
	)
	if err == nil {
		t.Fatal("expected unmarshal error")
	}
	if errors.Is(err, ErrDecrypt) {
		t.Fatal("unmarshal error must not be conflated with ErrDecrypt")
	}
}
