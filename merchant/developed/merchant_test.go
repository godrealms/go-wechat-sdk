package developed

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/godrealms/go-wechat-sdk/merchant/developed/types"
	"github.com/godrealms/go-wechat-sdk/utils"
)

func generateTestKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}
	return key
}

func newTestMerchantClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	key := generateTestKey(t)
	h := utils.NewHTTP(srv.URL, utils.WithTimeout(3*time.Second))
	return NewWechatClient().
		WithAppid("wx_test_appid").
		WithMchid("1234567890").
		WithCertificateNumber("CERT_SERIAL_NO").
		WithAPIv3Key("abcdefgh12345678abcdefgh12345678").
		WithPrivateKey(key).
		WithHttp(h)
}

func testOrder() *types.Transactions {
	return &types.Transactions{
		Appid:       "wx_test_appid",
		Mchid:       "1234567890",
		Description: "Test product",
		OutTradeNo:  "ORD20240101001",
		NotifyUrl:   "https://example.com/notify",
		Amount:      &types.Amount{Total: 100, Currency: "CNY"},
	}
}

func captureServer(t *testing.T, reply string) (*httptest.Server, *[]byte) {
	t.Helper()
	var captured []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(reply))
	}))
	return srv, &captured
}

// --- TransactionsApp tests ---

func TestTransactionsApp_Success(t *testing.T) {
	srv, _ := captureServer(t, `{"prepay_id":"wx20240101prepayid"}`)
	defer srv.Close()

	c := newTestMerchantClient(t, srv)
	result, err := c.TransactionsApp(testOrder())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PrepayId != "wx20240101prepayid" {
		t.Errorf("expected prepay_id wx20240101prepayid, got %q", result.PrepayId)
	}
}

func TestTransactionsApp_SetsAuthorizationHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"prepay_id":"pid"}`))
	}))
	defer srv.Close()

	c := newTestMerchantClient(t, srv)
	_, err := c.TransactionsApp(testOrder())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(gotAuth, "WECHATPAY2-SHA256-RSA2048") {
		t.Errorf("expected WECHATPAY2-SHA256-RSA2048 Authorization header, got: %q", gotAuth)
	}
}

func TestTransactionsApp_SendsCorrectJSON(t *testing.T) {
	srv, captured := captureServer(t, `{"prepay_id":"pid"}`)
	defer srv.Close()

	c := newTestMerchantClient(t, srv)
	_, err := c.TransactionsApp(testOrder())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var body map[string]interface{}
	if err := json.Unmarshal(*captured, &body); err != nil {
		t.Fatalf("server received invalid JSON: %v, body: %s", err, string(*captured))
	}
	if body["out_trade_no"] != "ORD20240101001" {
		t.Errorf("expected out_trade_no ORD20240101001, got: %v", body["out_trade_no"])
	}
}

func TestTransactionsApp_NetworkError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	u := srv.URL
	srv.Close()

	key := generateTestKey(t)
	h := utils.NewHTTP(u, utils.WithTimeout(time.Second))
	c := NewWechatClient().WithPrivateKey(key).WithHttp(h)
	_, err := c.TransactionsApp(testOrder())
	if err == nil {
		t.Error("expected network error")
	}
}

// --- TransactionsJsapi tests ---

func TestTransactionsJsapi_Success(t *testing.T) {
	srv, _ := captureServer(t, `{"prepay_id":"jsapi_prepay_id_001"}`)
	defer srv.Close()

	c := newTestMerchantClient(t, srv)
	result, err := c.TransactionsJsapi(testOrder())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PrepayId != "jsapi_prepay_id_001" {
		t.Errorf("expected jsapi prepay_id, got %q", result.PrepayId)
	}
}

func TestTransactionsJsapi_SetsAuthorizationHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"prepay_id":"jsapi_pid"}`))
	}))
	defer srv.Close()

	c := newTestMerchantClient(t, srv)
	_, err := c.TransactionsJsapi(testOrder())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(gotAuth, "WECHATPAY2-SHA256-RSA2048") {
		t.Errorf("expected WECHATPAY2-SHA256-RSA2048 Authorization header, got: %q", gotAuth)
	}
}

// --- TransactionsNative tests ---

func TestTransactionsNative_Success(t *testing.T) {
	srv, _ := captureServer(t, `{"code_url":"weixin://wxpay/bizpayurl?pr=test"}`)
	defer srv.Close()

	c := newTestMerchantClient(t, srv)
	result, err := c.TransactionsNative(testOrder())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.CodeUrl != "weixin://wxpay/bizpayurl?pr=test" {
		t.Errorf("expected code_url, got %q", result.CodeUrl)
	}
}

func TestTransactionsNative_SetsAuthorizationHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"code_url":"weixin://test"}`))
	}))
	defer srv.Close()

	c := newTestMerchantClient(t, srv)
	_, err := c.TransactionsNative(testOrder())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(gotAuth, "WECHATPAY2-SHA256-RSA2048") {
		t.Errorf("expected WECHATPAY2-SHA256-RSA2048 Authorization header, got: %q", gotAuth)
	}
}

// --- TransactionsH5 tests ---

func TestTransactionsH5_Success(t *testing.T) {
	srv, _ := captureServer(t, `{"h5_url":"https://wx.tenpay.com/cgi-bin/mmpaywap?prepay_id=test"}`)
	defer srv.Close()

	c := newTestMerchantClient(t, srv)
	result, err := c.TransactionsH5(testOrder())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.H5Url != "https://wx.tenpay.com/cgi-bin/mmpaywap?prepay_id=test" {
		t.Errorf("expected h5_url, got %q", result.H5Url)
	}
}

func TestTransactionsH5_SetsAuthorizationHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"h5_url":"https://test.com"}`))
	}))
	defer srv.Close()

	c := newTestMerchantClient(t, srv)
	_, err := c.TransactionsH5(testOrder())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(gotAuth, "WECHATPAY2-SHA256-RSA2048") {
		t.Errorf("expected WECHATPAY2-SHA256-RSA2048 Authorization header, got: %q", gotAuth)
	}
}

// --- Client builder chain tests ---

func TestNewWechatClient_DefaultBaseURL(t *testing.T) {
	c := NewWechatClient()
	if c.Http == nil {
		t.Fatal("expected non-nil Http")
	}
	if !strings.Contains(c.Http.BaseURL, "api.mch.weixin.qq.com") {
		t.Errorf("expected default base URL to contain api.mch.weixin.qq.com, got %q", c.Http.BaseURL)
	}
}

func TestWechatClient_BuilderChain(t *testing.T) {
	key := generateTestKey(t)
	c := NewWechatClient().
		WithAppid("appid").
		WithMchid("mchid").
		WithCertificateNumber("cert_no").
		WithAPIv3Key("key32byteslongkey32byteslongkey!").
		WithPrivateKey(key)
	if c.Appid != "appid" {
		t.Errorf("expected appid, got %q", c.Appid)
	}
	if c.Mchid != "mchid" {
		t.Errorf("expected mchid, got %q", c.Mchid)
	}
	if c.CertificateNumber != "cert_no" {
		t.Errorf("expected cert_no, got %q", c.CertificateNumber)
	}
	if c.privateKey != key {
		t.Error("expected private key to be set")
	}
}

func TestWechatClient_WithHttp(t *testing.T) {
	h := utils.NewHTTP("https://custom.example.com")
	c := NewWechatClient().WithHttp(h)
	if c.Http != h {
		t.Error("expected Http to be the injected instance")
	}
	if c.Http.BaseURL != "https://custom.example.com" {
		t.Errorf("expected custom base URL, got %q", c.Http.BaseURL)
	}
}

// --- ModifyTransactionsApp tests ---

func TestModifyTransactionsApp_Success(t *testing.T) {
	srv, _ := captureServer(t, `{"prepay_id":"wx_modify_prepay_id"}`)
	defer srv.Close()

	c := newTestMerchantClient(t, srv)
	result, err := c.ModifyTransactionsApp(testOrder())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PrepayId != "wx_modify_prepay_id" {
		t.Errorf("expected prepay_id wx_modify_prepay_id, got %q", result.PrepayId)
	}
	if result.AppId != "wx_test_appid" {
		t.Errorf("expected AppId wx_test_appid, got %q", result.AppId)
	}
	if result.PackageValue != "Sign=WXPay" {
		t.Errorf("expected PackageValue Sign=WXPay, got %q", result.PackageValue)
	}
}

// --- ModifyTransactionsJsapi tests ---

func TestModifyTransactionsJsapi_Success(t *testing.T) {
	srv, _ := captureServer(t, `{"prepay_id":"jsapi_modify_pid"}`)
	defer srv.Close()

	c := newTestMerchantClient(t, srv)
	result, err := c.ModifyTransactionsJsapi(testOrder())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedPkg := "prepay_id=jsapi_modify_pid"
	if result.Package != expectedPkg {
		t.Errorf("expected package %q, got %q", expectedPkg, result.Package)
	}
	if result.SignType != "RSA" {
		t.Errorf("expected SignType RSA, got %q", result.SignType)
	}
}

// --- QueryTransactionId tests ---

func TestQueryTransactionId_Success(t *testing.T) {
	srv, _ := captureServer(t, `{"transaction_id":"4200001234","out_trade_no":"ORD001","trade_state":"SUCCESS"}`)
	defer srv.Close()

	c := newTestMerchantClient(t, srv)
	result, err := c.QueryTransactionId("4200001234")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestQueryTransactionId_SetsAuthorizationHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := newTestMerchantClient(t, srv)
	_, _ = c.QueryTransactionId("txn123")
	if !strings.HasPrefix(gotAuth, "WECHATPAY2-SHA256-RSA2048") {
		t.Errorf("expected WECHATPAY2-SHA256-RSA2048 Authorization header, got: %q", gotAuth)
	}
}

// --- QueryOutTradeNo tests ---

func TestQueryOutTradeNo_Success(t *testing.T) {
	srv, _ := captureServer(t, `{"out_trade_no":"ORD001","trade_state":"SUCCESS"}`)
	defer srv.Close()

	c := newTestMerchantClient(t, srv)
	result, err := c.QueryOutTradeNo("ORD001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestQueryOutTradeNo_SetsAuthorizationHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := newTestMerchantClient(t, srv)
	_, _ = c.QueryOutTradeNo("ORD001")
	if !strings.HasPrefix(gotAuth, "WECHATPAY2-SHA256-RSA2048") {
		t.Errorf("expected WECHATPAY2-SHA256-RSA2048 Authorization header, got: %q", gotAuth)
	}
}

// --- TransactionsClose tests ---

func TestTransactionsClose_Success(t *testing.T) {
	srv, _ := captureServer(t, ``)
	defer srv.Close()

	c := newTestMerchantClient(t, srv)
	err := c.TransactionsClose("ORD001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTransactionsClose_SetsAuthorizationHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(``))
	}))
	defer srv.Close()

	c := newTestMerchantClient(t, srv)
	_ = c.TransactionsClose("ORD001")
	if !strings.HasPrefix(gotAuth, "WECHATPAY2-SHA256-RSA2048") {
		t.Errorf("expected WECHATPAY2-SHA256-RSA2048 Authorization header, got: %q", gotAuth)
	}
}

// --- Refunds tests ---

func TestRefunds_Success(t *testing.T) {
	srv, _ := captureServer(t, `{"refund_id":"REF001","out_refund_no":"REFNO001","status":"PROCESSING"}`)
	defer srv.Close()

	c := newTestMerchantClient(t, srv)
	refund := &types.Refunds{
		OutTradeNo:  "ORD001",
		OutRefundNo: "REFNO001",
		Amount:      &types.Amount{Total: 100, Currency: "CNY"},
	}
	result, err := c.Refunds(refund)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestRefunds_SetsAuthorizationHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"refund_id":"r1"}`))
	}))
	defer srv.Close()

	c := newTestMerchantClient(t, srv)
	refund := &types.Refunds{
		OutTradeNo:  "ORD001",
		OutRefundNo: "REFNO001",
		Amount:      &types.Amount{Total: 100, Currency: "CNY"},
	}
	_, _ = c.Refunds(refund)
	if !strings.HasPrefix(gotAuth, "WECHATPAY2-SHA256-RSA2048") {
		t.Errorf("expected WECHATPAY2-SHA256-RSA2048 Authorization header, got: %q", gotAuth)
	}
}

// --- QueryRefunds tests ---

func TestQueryRefunds_Success(t *testing.T) {
	srv, _ := captureServer(t, `{"refund_id":"REF001","out_refund_no":"REFNO001","status":"SUCCESS"}`)
	defer srv.Close()

	c := newTestMerchantClient(t, srv)
	result, err := c.QueryRefunds("REFNO001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestQueryRefunds_SetsAuthorizationHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := newTestMerchantClient(t, srv)
	_, _ = c.QueryRefunds("REFNO001")
	if !strings.HasPrefix(gotAuth, "WECHATPAY2-SHA256-RSA2048") {
		t.Errorf("expected WECHATPAY2-SHA256-RSA2048 Authorization header, got: %q", gotAuth)
	}
}

// --- ApplyAbnormalRefund tests ---

func TestApplyAbnormalRefund_Success(t *testing.T) {
	srv, _ := captureServer(t, `{"refund_id":"REF002","status":"PROCESSING"}`)
	defer srv.Close()

	c := newTestMerchantClient(t, srv)
	abnormal := &types.AbnormalRefund{
		OutRefundNo: "REFNO002",
	}
	result, err := c.ApplyAbnormalRefund("REF001", abnormal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

// --- TradeBill tests ---

func TestTradeBill_Success(t *testing.T) {
	srv, _ := captureServer(t, `{"download_url":"https://api.mch.weixin.qq.com/v3/billdownload/file","hash_value":"abc"}`)
	defer srv.Close()

	c := newTestMerchantClient(t, srv)
	quest := &types.TradeBillQuest{BillDate: "2024-01-01"}
	result, err := c.TradeBill(quest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestTradeBill_SetsAuthorizationHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := newTestMerchantClient(t, srv)
	_, _ = c.TradeBill(&types.TradeBillQuest{BillDate: "2024-01-01"})
	if !strings.HasPrefix(gotAuth, "WECHATPAY2-SHA256-RSA2048") {
		t.Errorf("expected WECHATPAY2-SHA256-RSA2048 Authorization header, got: %q", gotAuth)
	}
}

// --- FundFlowBill tests ---

func TestFundFlowBill_Success(t *testing.T) {
	srv, _ := captureServer(t, `{"download_url":"https://api.mch.weixin.qq.com/v3/billdownload/file","hash_value":"def"}`)
	defer srv.Close()

	c := newTestMerchantClient(t, srv)
	quest := &types.FundsBillQuest{BillDate: "2024-01-01"}
	result, err := c.FundFlowBill(quest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestFundFlowBill_SetsAuthorizationHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := newTestMerchantClient(t, srv)
	_, _ = c.FundFlowBill(&types.FundsBillQuest{BillDate: "2024-01-01"})
	if !strings.HasPrefix(gotAuth, "WECHATPAY2-SHA256-RSA2048") {
		t.Errorf("expected WECHATPAY2-SHA256-RSA2048 Authorization header, got: %q", gotAuth)
	}
}

// --- WithCertificate and WithPublicKey builder tests ---

func TestWechatClient_WithCertificateAndPublicKey(t *testing.T) {
	c := NewWechatClient().
		WithPublicKey(nil)
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

// --- ParseNotification / notify tests ---

func TestParseNotification_NilRequest(t *testing.T) {
	c := NewWechatClient()
	_, err := c.ParseNotification(nil, nil, nil)
	if err == nil {
		t.Error("expected error for nil request")
	}
}

func TestParseNotification_NilBody(t *testing.T) {
	c := NewWechatClient()
	req := httptest.NewRequest("POST", "/notify", nil)
	req.Body = nil
	_, err := c.ParseNotification(nil, req, nil)
	if err == nil {
		t.Error("expected error for nil body")
	}
}

func TestParseNotification_SignatureError(t *testing.T) {
	c := NewWechatClient()
	req := httptest.NewRequest("POST", "/notify", strings.NewReader(`{"id":"evt001"}`))
	_, err := c.ParseNotification(nil, req, nil)
	// verifyResponseSignature is a stub that always errors, so we expect an error
	if err == nil {
		t.Error("expected error due to signature verification stub")
	}
}

func TestAckNotification_WritesSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	AckNotification(w)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "SUCCESS") {
		t.Errorf("expected SUCCESS in body, got: %s", body)
	}
}

func TestFailNotification_WritesError(t *testing.T) {
	w := httptest.NewRecorder()
	FailNotification(w, "bad request")
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "FAIL") {
		t.Errorf("expected FAIL in body, got: %s", body)
	}
}

func TestFailNotification_EmptyMessage(t *testing.T) {
	w := httptest.NewRecorder()
	FailNotification(w, "")
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "FAILED") {
		t.Errorf("expected FAILED in body, got: %s", body)
	}
}

// --- ParseRefundNotify test ---

func TestParseRefundNotify_SignatureError(t *testing.T) {
	c := NewWechatClient()
	req := httptest.NewRequest("POST", "/notify", strings.NewReader(`{"id":"evt001"}`))
	_, _, err := c.ParseRefundNotify(nil, req)
	// verifyResponseSignature is a stub that always errors
	if err == nil {
		t.Error("expected error due to signature verification stub")
	}
}

// --- CreateTransferBatch test (uses postV3 stub) ---

func TestCreateTransferBatch_StubError(t *testing.T) {
	c := NewWechatClient()
	_, err := c.CreateTransferBatch(context.Background(), map[string]any{"test": "val"})
	if err == nil {
		t.Error("expected error from CreateTransferBatch (no private key/mchid)")
	}
}
