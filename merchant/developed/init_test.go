package developed_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	developed "github.com/godrealms/go-wechat-sdk/merchant/developed"
	"github.com/godrealms/go-wechat-sdk/utils"
)

// newClientWithFakeServer creates a test client with a fake HTTP server.
// It returns the client, the server URL, and the server itself.
// The caller is responsible for calling srv.Close().
func newClientWithFakeServer(t *testing.T) (*developed.Client, string, *httptest.Server) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Default response - return success
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{
			"code": 0,
			"msg":  "ok",
		}
		_ = json.NewEncoder(w).Encode(response)
	}))

	client := &developed.Client{
		Appid:             "test-appid",
		Mchid:             "test-mchid",
		CertificateNumber: "test-cert-number",
		APIv3Key:          "test-api-v3-key",
		Http:              utils.NewHTTP(server.URL),
	}

	return client, server.URL, server
}
