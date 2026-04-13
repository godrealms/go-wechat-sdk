package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
)

func generateKeyPEM(t *testing.T, pkcs1 bool) string {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	var blockBytes []byte
	var blockType string
	if pkcs1 {
		blockBytes = x509.MarshalPKCS1PrivateKey(priv)
		blockType = "RSA PRIVATE KEY"
	} else {
		blockBytes, err = x509.MarshalPKCS8PrivateKey(priv)
		if err != nil {
			t.Fatal(err)
		}
		blockType = "PRIVATE KEY"
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: blockType, Bytes: blockBytes}))
}

func TestLoadPrivateKey_PKCS8(t *testing.T) {
	keyPEM := generateKeyPEM(t, false)
	if _, err := LoadPrivateKey(keyPEM); err != nil {
		t.Fatalf("load PKCS8 failed: %v", err)
	}
}

func TestLoadPrivateKey_PKCS1(t *testing.T) {
	keyPEM := generateKeyPEM(t, true)
	if _, err := LoadPrivateKey(keyPEM); err != nil {
		t.Fatalf("load PKCS1 failed: %v", err)
	}
}

func TestLoadPrivateKey_BadInput(t *testing.T) {
	if _, err := LoadPrivateKey("garbage"); err == nil {
		t.Error("expected error for garbage input")
	}
	if _, err := LoadPrivateKey(string(pem.EncodeToMemory(&pem.Block{
		Type: "CERTIFICATE", Bytes: []byte{0},
	}))); err == nil {
		t.Error("expected error for wrong PEM type")
	}
}
