package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"fmt"
)

// DecryptAEADAES256GCM decrypts a WeChat Pay v3 AEAD_AES_256_GCM payload.
// It accepts the inputs as the WeChat callback delivers them: a base64-encoded
// ciphertext, the raw nonce string, the associated_data string, and the
// merchant's APIv3 key.
//
// The plaintext returned is the JSON encoding of the decrypted resource (e.g.
// a Transaction or RefundResp). An empty plaintext is treated as a decryption
// failure rather than a successful zero-length result, since WeChat never
// sends empty resources.
//
// This is the canonical implementation used by both pay.Client (platform
// certificate refresh + ParseNotification) and types.Notify (the deprecated
// caller-driven API). Do not duplicate it elsewhere — file a refactor first.
func DecryptAEADAES256GCM(key, nonce, associatedData, ciphertext string) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("aead: decode ciphertext base64: %w", err)
	}
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, fmt.Errorf("aead: build aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("aead: build gcm: %w", err)
	}
	plain, err := gcm.Open(nil, []byte(nonce), decoded, []byte(associatedData))
	if err != nil {
		return nil, fmt.Errorf("aead: gcm open: %w", err)
	}
	if len(plain) == 0 {
		return nil, errors.New("aead: empty plaintext")
	}
	return plain, nil
}
