package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
)

func TestSignAndVerifySHA256WithRSA(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	source := "POST\n/v3/pay/transactions/jsapi\n1700000000\nnonce\n{\"a\":1}\n"

	sig, err := SignSHA256WithRSA(source, priv)
	if err != nil {
		t.Fatalf("sign failed: %v", err)
	}
	if sig == "" {
		t.Fatal("empty signature")
	}

	if err := VerifySHA256WithRSA(source, sig, &priv.PublicKey); err != nil {
		t.Errorf("verify failed: %v", err)
	}

	// 篡改后必须验证失败
	if err := VerifySHA256WithRSA(source+"x", sig, &priv.PublicKey); err == nil {
		t.Error("verify should have failed for tampered source")
	}
}

func TestSignSHA256WithRSARejectsNilKey(t *testing.T) {
	if _, err := SignSHA256WithRSA("x", nil); err == nil {
		t.Error("expected error for nil private key")
	}
}

func TestVerifySHA256WithRSARejectsNilKey(t *testing.T) {
	if err := VerifySHA256WithRSA("x", "y", nil); err == nil {
		t.Error("expected error for nil public key")
	}
}
