package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"strings"
	"testing"
)

// makeGCMPayload builds a fresh AEAD_AES_256_GCM payload for the given inputs.
// Returns (base64Ciphertext, nonce, associatedData, key).
func makeGCMPayload(t *testing.T, plaintext, associatedData string) (string, string, string, string) {
	t.Helper()
	key := make([]byte, 32) // AES-256
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatal(err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		t.Fatal(err)
	}
	ct := gcm.Seal(nil, nonce, []byte(plaintext), []byte(associatedData))
	return base64.StdEncoding.EncodeToString(ct), string(nonce), associatedData, string(key)
}

func TestDecryptAEADAES256GCM_RoundTrip(t *testing.T) {
	plaintext := `{"transaction_id":"4200001234"}`
	ad := "transaction"
	ct, nonce, ad2, key := makeGCMPayload(t, plaintext, ad)

	got, err := DecryptAEADAES256GCM(key, nonce, ad2, ct)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != plaintext {
		t.Errorf("got %q want %q", got, plaintext)
	}
}

func TestDecryptAEADAES256GCM_RejectsBadKey(t *testing.T) {
	ct, nonce, ad, _ := makeGCMPayload(t, "secret", "ad")
	wrongKey := strings.Repeat("x", 32)
	_, err := DecryptAEADAES256GCM(wrongKey, nonce, ad, ct)
	if err == nil {
		t.Fatal("expected error with wrong key")
	}
	if !strings.Contains(err.Error(), "gcm open") {
		t.Errorf("error should be from gcm open, got: %v", err)
	}
}

func TestDecryptAEADAES256GCM_RejectsBadAssociatedData(t *testing.T) {
	ct, nonce, _, key := makeGCMPayload(t, "secret", "original_ad")
	_, err := DecryptAEADAES256GCM(key, nonce, "tampered_ad", ct)
	if err == nil {
		t.Fatal("expected error with mismatched associated_data")
	}
}

func TestDecryptAEADAES256GCM_InvalidBase64(t *testing.T) {
	_, err := DecryptAEADAES256GCM(strings.Repeat("k", 32), "nonce123", "ad", "not_base64!!")
	if err == nil {
		t.Fatal("expected base64 decode error")
	}
	if !strings.Contains(err.Error(), "decode ciphertext") {
		t.Errorf("error should reference base64 decode, got: %v", err)
	}
}

func TestDecryptAEADAES256GCM_EmptyPlaintextRejected(t *testing.T) {
	// Encrypt an empty string. The decrypt should reject it (WeChat never
	// sends empty resources, so empty plaintext is a sentinel for "something
	// weird happened").
	ct, nonce, ad, key := makeGCMPayload(t, "", "ad")
	_, err := DecryptAEADAES256GCM(key, nonce, ad, ct)
	if err == nil {
		t.Fatal("expected error for empty plaintext")
	}
	if !strings.Contains(err.Error(), "empty plaintext") {
		t.Errorf("error should mention empty plaintext, got: %v", err)
	}
}
