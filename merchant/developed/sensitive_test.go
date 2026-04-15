package pay

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"strings"
	"testing"
)

func TestEncryptSensitiveField_RoundTrip(t *testing.T) {
	priv, cert := newTestKeyAndCert(t)

	ciphertext, err := EncryptSensitiveField(cert, "张三")
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}
	if ciphertext == "" {
		t.Fatal("empty ciphertext")
	}

	raw, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		t.Fatalf("base64 decode: %v", err)
	}
	// 用对应的私钥解密，验证 SDK 走的确实是 RSA-OAEP(SHA256)。
	plain, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, priv, raw, nil)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if string(plain) != "张三" {
		t.Errorf("round trip mismatch: %q", string(plain))
	}
}

func TestEncryptSensitiveField_NilCert(t *testing.T) {
	if _, err := EncryptSensitiveField(nil, "x"); err == nil {
		t.Fatal("expected error for nil cert")
	}
}

func TestEncryptSensitiveField_NonRSACert(t *testing.T) {
	// 伪造一个 PublicKey 非 *rsa.PublicKey 的证书：直接清空原证书的 PublicKey 字段
	_, cert := newTestKeyAndCert(t)
	cert.PublicKey = "not-rsa"
	_, err := EncryptSensitiveField(cert, "x")
	if err == nil || !strings.Contains(err.Error(), "not RSA") {
		t.Fatalf("expected non-RSA error, got %v", err)
	}
}

func TestClient_PlatformCertForEncrypt_FromCache(t *testing.T) {
	client, _, srv := newClientWithFakeServer(t)
	defer srv.Close()

	// newClientWithFakeServer 已经调用 AddPlatformCertificate，缓存里有一张。
	cert, serial, err := client.PlatformCertForEncrypt(context.Background())
	if err != nil {
		t.Fatalf("PlatformCertForEncrypt failed: %v", err)
	}
	if cert == nil {
		t.Fatal("cert is nil")
	}
	if serial == "" {
		t.Fatal("serial is empty")
	}
}
