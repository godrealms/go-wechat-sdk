package pay_test

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"

	developed "github.com/godrealms/go-wechat-sdk/merchant/developed"
	"github.com/godrealms/go-wechat-sdk/utils"
)

// newClientWithFakeServer creates a test client backed by a fake HTTP server.
// The caller is responsible for calling srv.Close().
func newClientWithFakeServer(t *testing.T) (*developed.Client, string, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"code":0,"msg":"ok"}`))
	}))

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	c := developed.NewWechatClient().
		WithHttp(utils.NewHTTP(server.URL)).
		WithAppid("test-appid").
		WithMchid("test-mchid").
		WithCertificateNumber("test-cert-number").
		WithAPIv3Key("01234567890123456789012345678901").
		WithPrivateKey(key)

	return c, server.URL, server
}
