package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
)

func TestSignSHA256WithRSA_NilPrivateKey(t *testing.T) {
	_, err := SignSHA256WithRSA("test source", nil)
	if err == nil {
		t.Fatal("expected error for nil private key, got nil")
	}
}

func TestSignSHA256WithRSA_ValidKey(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	signature, err := SignSHA256WithRSA("test source", privateKey)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if signature == "" {
		t.Fatal("expected non-empty signature, got empty string")
	}
}
