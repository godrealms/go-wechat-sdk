package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	pay "github.com/godrealms/go-wechat-sdk/merchant/developed"
	"github.com/godrealms/go-wechat-sdk/utils"
)

// newTestKeyAndCert 生成一份用于测试的 RSA 私钥 + 自签证书。同一对密钥在测试里
// 既扮演商户签名密钥、也扮演微信支付平台证书公钥，这样就能用同一个 priv 既给
// 请求签名、又给响应伪造签名，避免真实拉取 /v3/certificates。
func newTestKeyAndCert(t *testing.T) (*rsa.PrivateKey, *x509.Certificate) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(0xC0FFEE),
		Subject:      pkix.Name{CommonName: "service-test"},
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

type recordedRequest struct {
	Method string
	Path   string
	Body   []byte
	Header http.Header
}

// fakeServer 是一个能签响应签名的 httptest.Server，服务 service 包测试使用。
// 它复用 core_test.go 的结构（未导出，所以重新实现一份）。
type fakeServer struct {
	t       *testing.T
	priv    *rsa.PrivateKey
	cert    *x509.Certificate
	mu      sync.Mutex
	history []recordedRequest
	// respond 让每个用例自定义返回的 (statusCode, body)。
	respond func(r *http.Request) (int, []byte)
}

func (f *fakeServer) handler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	f.mu.Lock()
	f.history = append(f.history, recordedRequest{
		Method: r.Method,
		Path:   r.URL.RequestURI(),
		Body:   body,
		Header: r.Header.Clone(),
	})
	f.mu.Unlock()

	status, respBody := 200, []byte(`{}`)
	if f.respond != nil {
		status, respBody = f.respond(r)
	}

	ts := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := "noncefortest12345678"
	source := fmt.Sprintf("%s\n%s\n%s\n", ts, nonce, string(respBody))
	sig, err := utils.SignSHA256WithRSA(source, f.priv)
	if err != nil {
		f.t.Fatalf("sign response: %v", err)
	}
	w.Header().Set("Wechatpay-Timestamp", ts)
	w.Header().Set("Wechatpay-Nonce", nonce)
	w.Header().Set("Wechatpay-Signature", sig)
	w.Header().Set("Wechatpay-Serial", utils.GetCertificateSerialNumber(*f.cert))
	w.WriteHeader(status)
	if status != http.StatusNoContent {
		_, _ = w.Write(respBody)
	}
}

func (f *fakeServer) lastRequest(t *testing.T) recordedRequest {
	t.Helper()
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.history) == 0 {
		t.Fatal("no requests recorded")
	}
	return f.history[len(f.history)-1]
}

// newFakeClient 构造一个指向假服务器的 service.Client。
func newFakeClient(t *testing.T) (*Client, *fakeServer, *httptest.Server) {
	t.Helper()
	priv, cert := newTestKeyAndCert(t)
	fs := &fakeServer{t: t, priv: priv, cert: cert}
	srv := httptest.NewServer(http.HandlerFunc(fs.handler))

	httpClient := utils.NewHTTP(srv.URL)
	inner, err := pay.NewClient(pay.Config{
		Appid:             "wx_sp_appid",
		Mchid:             "1900000001",
		CertificateNumber: "TEST_SERIAL",
		APIv3Key:          "0123456789012345678901234567890_",
		PrivateKey:        priv,
		HTTP:              httpClient,
	})
	if err != nil {
		t.Fatal(err)
	}
	inner.AddPlatformCertificate(cert)

	c := &Client{
		inner:    inner,
		subMchid: "1900000002",
		subAppid: "wx_sub_appid",
	}
	return c, fs, srv
}
