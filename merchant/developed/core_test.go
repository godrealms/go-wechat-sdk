package pay

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/godrealms/go-wechat-sdk/merchant/developed/types"
	"github.com/godrealms/go-wechat-sdk/utils"
)

// helper: 生成测试用的 RSA 商户私钥与对应的证书（这里复用为"假的"平台证书）。
func newTestKeyAndCert(t *testing.T) (*rsa.PrivateKey, *x509.Certificate) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(0xABCDEF),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	return priv, cert
}

func signResponseWithUtils(t *testing.T, priv *rsa.PrivateKey, timestamp, nonce, body string) string {
	t.Helper()
	source := timestamp + "\n" + nonce + "\n" + body + "\n"
	sig, err := utils.SignSHA256WithRSA(source, priv)
	if err != nil {
		t.Fatal(err)
	}
	return sig
}

// fakeServer 返回一个签好响应签名的 httptest.Server。
type fakeServer struct {
	t          *testing.T
	priv       *rsa.PrivateKey
	cert       *x509.Certificate
	requests   []recordedRequest
	requestsMu sync.Mutex
	respond    func(r *http.Request) (status int, body []byte)
}

type recordedRequest struct {
	Method string
	Path   string
	Auth   string
	Body   []byte
}

func (f *fakeServer) handler(w http.ResponseWriter, r *http.Request) {
	bodyBytes, _ := io.ReadAll(r.Body)
	f.requestsMu.Lock()
	f.requests = append(f.requests, recordedRequest{
		Method: r.Method,
		Path:   r.URL.RequestURI(),
		Auth:   r.Header.Get("Authorization"),
		Body:   bodyBytes,
	})
	f.requestsMu.Unlock()

	status, body := 200, []byte(`{"prepay_id":"wx_prepay_123"}`)
	if f.respond != nil {
		status, body = f.respond(r)
	}

	ts := time.Now().Unix()
	nonce := "noncefortest12345678"
	sig := signResponseWithUtils(f.t, f.priv, intToString(ts), nonce, string(body))
	w.Header().Set("Wechatpay-Timestamp", intToString(ts))
	w.Header().Set("Wechatpay-Nonce", nonce)
	w.Header().Set("Wechatpay-Signature", sig)
	w.Header().Set("Wechatpay-Serial", utils.GetCertificateSerialNumber(*f.cert))
	w.WriteHeader(status)
	if status != http.StatusNoContent {
		_, _ = w.Write(body)
	}
}

func intToString(n int64) string {
	// 避免引入额外格式化包
	if n == 0 {
		return "0"
	}
	negative := n < 0
	if negative {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if negative {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func newClientWithFakeServer(t *testing.T) (*Client, *fakeServer, *httptest.Server) {
	priv, cert := newTestKeyAndCert(t)
	fs := &fakeServer{t: t, priv: priv, cert: cert}
	srv := httptest.NewServer(http.HandlerFunc(fs.handler))

	httpClient := utils.NewHTTP(srv.URL)
	client, err := NewClient(Config{
		Appid:             "wxtest",
		Mchid:             "1900000001",
		CertificateNumber: "TEST_SERIAL",
		APIv3Key:          "0123456789012345678901234567890_",
		PrivateKey:        priv, // 用同一对密钥扮演商户与平台
		HTTP:              httpClient,
	})
	if err != nil {
		t.Fatal(err)
	}
	client.AddPlatformCertificate(cert) // 注入平台证书，跳过实际证书拉取
	return client, fs, srv
}

func TestClient_NewClient_RequiresFields(t *testing.T) {
	_, err := NewClient(Config{})
	if err == nil {
		t.Fatal("expected error for empty config")
	}
}

func TestNewClient_RejectsShortAPIv3Key(t *testing.T) {
	priv, _ := newTestKeyAndCert(t)
	_, err := NewClient(Config{
		Appid:             "wxtest",
		Mchid:             "1900000001",
		CertificateNumber: "TEST",
		APIv3Key:          "tooshort",
		PrivateKey:        priv,
	})
	if err == nil || !strings.Contains(err.Error(), "APIv3Key must be 32 bytes") {
		t.Fatalf("expected length validation error, got %v", err)
	}
}

func TestClient_TransactionsJsapi_SignsAndVerifies(t *testing.T) {
	client, fs, srv := newClientWithFakeServer(t)
	defer srv.Close()

	resp, err := client.TransactionsJsapi(context.Background(), &types.Transactions{
		Appid:      "wxtest",
		Mchid:      "1900000001",
		OutTradeNo: "OUT123",
	})
	if err != nil {
		t.Fatalf("TransactionsJsapi failed: %v", err)
	}
	if resp.PrepayId != "wx_prepay_123" {
		t.Errorf("unexpected prepay_id: %s", resp.PrepayId)
	}
	if len(fs.requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(fs.requests))
	}
	if !strings.Contains(fs.requests[0].Auth, `WECHATPAY2-SHA256-RSA2048`) {
		t.Errorf("missing WeChat Pay auth scheme: %s", fs.requests[0].Auth)
	}
	if !strings.Contains(fs.requests[0].Auth, `mchid="1900000001"`) {
		t.Errorf("missing mchid in auth: %s", fs.requests[0].Auth)
	}
}

func TestClient_TransactionsClose_SendsCorrectBody(t *testing.T) {
	client, fs, srv := newClientWithFakeServer(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) { return http.StatusNoContent, nil }

	if err := client.TransactionsClose(context.Background(), "OUT123"); err != nil {
		t.Fatalf("TransactionsClose failed: %v", err)
	}
	if len(fs.requests) != 1 {
		t.Fatalf("expected 1 request")
	}
	body := string(fs.requests[0].Body)
	// 之前的 bug：body 会变成 "{\"mchid\":\"...\"}" 这种被双重 marshal 的字符串。
	// 修复后必须是真正的 JSON 对象。
	if body != `{"mchid":"1900000001"}` {
		t.Errorf("unexpected body: %s", body)
	}
	if fs.requests[0].Path != "/v3/pay/transactions/out-trade-no/OUT123/close" {
		t.Errorf("unexpected path: %s", fs.requests[0].Path)
	}
}

func TestClient_QueryTransactionId_PropagatesError(t *testing.T) {
	client, fs, srv := newClientWithFakeServer(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return http.StatusBadRequest, []byte(`{"code":"PARAM_ERROR","message":"bad"}`)
	}

	_, err := client.QueryTransactionId(context.Background(), "12345")
	if err == nil {
		t.Fatal("expected error to be propagated")
	}
}

func TestClient_QueryTransactionId_QueryHasMchid(t *testing.T) {
	client, fs, srv := newClientWithFakeServer(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) { return 200, []byte(`{"out_trade_no":"X"}`) }

	if _, err := client.QueryTransactionId(context.Background(), "12345"); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if !strings.Contains(fs.requests[0].Path, "mchid=1900000001") {
		t.Errorf("missing mchid in query: %s", fs.requests[0].Path)
	}
}

func TestClient_DoV3_RejectsTamperedResponse(t *testing.T) {
	priv, cert := newTestKeyAndCert(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := []byte(`{"prepay_id":"x"}`)
		ts := time.Now().Unix()
		// 故意签错的内容
		sig := signResponseWithUtils(t, priv, intToString(ts), "noncefortest12345678", "different body")
		w.Header().Set("Wechatpay-Timestamp", intToString(ts))
		w.Header().Set("Wechatpay-Nonce", "noncefortest12345678")
		w.Header().Set("Wechatpay-Signature", sig)
		w.Header().Set("Wechatpay-Serial", utils.GetCertificateSerialNumber(*cert))
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	client, err := NewClient(Config{
		Appid:             "wxtest",
		Mchid:             "1900000001",
		CertificateNumber: "TEST",
		APIv3Key:          "0123456789012345678901234567890_",
		PrivateKey:        priv,
		HTTP:              utils.NewHTTP(srv.URL),
	})
	if err != nil {
		t.Fatal(err)
	}
	client.AddPlatformCertificate(cert)

	_, err = client.TransactionsJsapi(context.Background(), &types.Transactions{Appid: "wxtest", Mchid: "1900000001"})
	if err == nil || !strings.Contains(err.Error(), "verify response signature") {
		t.Fatalf("expected signature verification failure, got: %v", err)
	}
}

func TestClient_DoV3_RejectsResponseWithMissingSignatureHeaders(t *testing.T) {
	priv, cert := newTestKeyAndCert(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Deliberately do NOT set any Wechatpay-* headers.
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"prepay_id":"x"}`))
	}))
	defer srv.Close()

	client, err := NewClient(Config{
		Appid:             "wxtest",
		Mchid:             "1900000001",
		CertificateNumber: "TEST",
		APIv3Key:          "0123456789012345678901234567890_",
		PrivateKey:        priv,
		HTTP:              utils.NewHTTP(srv.URL),
	})
	if err != nil {
		t.Fatal(err)
	}
	client.AddPlatformCertificate(cert)

	_, err = client.TransactionsJsapi(context.Background(), &types.Transactions{Appid: "wxtest", Mchid: "1900000001"})
	if err == nil {
		t.Fatal("expected error when response has no signature headers")
	}
	if !strings.Contains(err.Error(), "missing wechatpay signature header") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestClient_DoV3_RejectsStaleWechatpayTimestamp(t *testing.T) {
	priv, cert := newTestKeyAndCert(t)
	staleTs := time.Now().Unix() - 3600 // 1 hour ago
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := []byte(`{"prepay_id":"x"}`)
		nonce := "noncefortest12345678"
		sig := signResponseWithUtils(t, priv, intToString(staleTs), nonce, string(body))
		w.Header().Set("Wechatpay-Timestamp", intToString(staleTs))
		w.Header().Set("Wechatpay-Nonce", nonce)
		w.Header().Set("Wechatpay-Signature", sig)
		w.Header().Set("Wechatpay-Serial", utils.GetCertificateSerialNumber(*cert))
		w.WriteHeader(200)
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	client, err := NewClient(Config{
		Appid:             "wxtest",
		Mchid:             "1900000001",
		CertificateNumber: "TEST",
		APIv3Key:          "0123456789012345678901234567890_",
		PrivateKey:        priv,
		HTTP:              utils.NewHTTP(srv.URL),
	})
	if err != nil {
		t.Fatal(err)
	}
	client.AddPlatformCertificate(cert)

	_, err = client.TransactionsJsapi(context.Background(), &types.Transactions{Appid: "wxtest", Mchid: "1900000001"})
	if err == nil {
		t.Fatal("expected error when wechatpay timestamp is stale")
	}
	if !strings.Contains(err.Error(), "timestamp out of window") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestClient_PEMRoundtrip(t *testing.T) {
	priv, _ := newTestKeyAndCert(t)
	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})
	loaded, err := utils.LoadPrivateKey(string(pemBytes))
	if err != nil {
		t.Fatal(err)
	}
	if loaded.N.Cmp(priv.N) != 0 {
		t.Error("loaded key mismatches")
	}
}

func TestClient_DoV3_ParsesV3ErrorEnvelope(t *testing.T) {
	client, fs, srv := newClientWithFakeServer(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return http.StatusBadRequest, []byte(`{"code":"PARAM_ERROR","message":"appid invalid"}`)
	}

	_, err := client.TransactionsJsapi(context.Background(), &types.Transactions{
		Appid: "wxtest", Mchid: "1900000001",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	var v3 *V3Error
	if !errors.As(err, &v3) {
		t.Fatalf("expected *V3Error, got %T: %v", err, err)
	}
	if v3.Code != "PARAM_ERROR" {
		t.Errorf("unexpected code: %s", v3.Code)
	}
	if v3.HTTPStatus != http.StatusBadRequest {
		t.Errorf("unexpected status: %d", v3.HTTPStatus)
	}
}

// The fallback branch in doV3 must preserve the underlying *utils.HTTPError
// when the response body is not a parseable v3 envelope. Three shapes:
// empty body, valid JSON with no code field, and outright garbage.
func TestClient_DoV3_FallsBackWhenNotV3Envelope(t *testing.T) {
	cases := []struct {
		name string
		body []byte
	}{
		{"empty body", nil},
		{"json without code field", []byte(`{"message":"oops"}`)},
		{"not json at all", []byte(`<html>502 Bad Gateway</html>`)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client, fs, srv := newClientWithFakeServer(t)
			defer srv.Close()
			fs.respond = func(r *http.Request) (int, []byte) {
				return http.StatusBadGateway, tc.body
			}

			_, err := client.TransactionsJsapi(context.Background(), &types.Transactions{
				Appid: "wxtest", Mchid: "1900000001",
			})
			if err == nil {
				t.Fatal("expected error")
			}
			var v3 *V3Error
			if errors.As(err, &v3) {
				t.Errorf("expected NOT to unwrap into *V3Error, but got: %+v", v3)
			}
			var httpErr *utils.HTTPError
			if !errors.As(err, &httpErr) {
				t.Errorf("expected *utils.HTTPError fallback, got %T: %v", err, err)
			}
		})
	}
}
