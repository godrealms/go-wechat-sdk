package pay

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
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

// newTestMerchantClient returns a Client wired up to a signed fake httptest.Server.
// Use the returned *fakeServer to inspect captured requests (fs.requests) and to
// override the response (fs.respond).
func newTestMerchantClient(t *testing.T) (*Client, *fakeServer, *httptest.Server) {
	t.Helper()
	return newClientWithFakeServer(t)
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

// --- TransactionsApp tests ---

func TestTransactionsApp_Success(t *testing.T) {
	c, fs, srv := newTestMerchantClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{"prepay_id":"wx20240101prepayid"}`)
	}

	result, err := c.TransactionsApp(context.Background(), testOrder())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PrepayId != "wx20240101prepayid" {
		t.Errorf("expected prepay_id wx20240101prepayid, got %q", result.PrepayId)
	}
}

func TestTransactionsApp_SetsAuthorizationHeader(t *testing.T) {
	c, fs, srv := newTestMerchantClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{"prepay_id":"pid"}`)
	}

	_, err := c.TransactionsApp(context.Background(), testOrder())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fs.requests) == 0 {
		t.Fatal("expected at least one request")
	}
	if !strings.HasPrefix(fs.requests[0].Auth, "WECHATPAY2-SHA256-RSA2048") {
		t.Errorf("expected WECHATPAY2-SHA256-RSA2048 Authorization header, got: %q", fs.requests[0].Auth)
	}
}

func TestTransactionsApp_SendsCorrectJSON(t *testing.T) {
	c, fs, srv := newTestMerchantClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{"prepay_id":"pid"}`)
	}

	_, err := c.TransactionsApp(context.Background(), testOrder())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fs.requests) == 0 {
		t.Fatal("expected at least one request")
	}
	var body map[string]interface{}
	if err := json.Unmarshal(fs.requests[0].Body, &body); err != nil {
		t.Fatalf("server received invalid JSON: %v, body: %s", err, string(fs.requests[0].Body))
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
	_, err := c.TransactionsApp(context.Background(), testOrder())
	if err == nil {
		t.Error("expected network error")
	}
}

// --- TransactionsJsapi tests ---

func TestTransactionsJsapi_Success(t *testing.T) {
	c, fs, srv := newTestMerchantClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{"prepay_id":"jsapi_prepay_id_001"}`)
	}

	result, err := c.TransactionsJsapi(context.Background(), testOrder())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PrepayId != "jsapi_prepay_id_001" {
		t.Errorf("expected jsapi prepay_id, got %q", result.PrepayId)
	}
}

func TestTransactionsJsapi_SetsAuthorizationHeader(t *testing.T) {
	c, fs, srv := newTestMerchantClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{"prepay_id":"jsapi_pid"}`)
	}

	_, err := c.TransactionsJsapi(context.Background(), testOrder())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fs.requests) == 0 {
		t.Fatal("expected at least one request")
	}
	if !strings.HasPrefix(fs.requests[0].Auth, "WECHATPAY2-SHA256-RSA2048") {
		t.Errorf("expected WECHATPAY2-SHA256-RSA2048 Authorization header, got: %q", fs.requests[0].Auth)
	}
}

// --- TransactionsNative tests ---

func TestTransactionsNative_Success(t *testing.T) {
	c, fs, srv := newTestMerchantClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{"code_url":"weixin://wxpay/bizpayurl?pr=test"}`)
	}

	result, err := c.TransactionsNative(context.Background(), testOrder())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.CodeUrl != "weixin://wxpay/bizpayurl?pr=test" {
		t.Errorf("expected code_url, got %q", result.CodeUrl)
	}
}

func TestTransactionsNative_SetsAuthorizationHeader(t *testing.T) {
	c, fs, srv := newTestMerchantClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{"code_url":"weixin://test"}`)
	}

	_, err := c.TransactionsNative(context.Background(), testOrder())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fs.requests) == 0 {
		t.Fatal("expected at least one request")
	}
	if !strings.HasPrefix(fs.requests[0].Auth, "WECHATPAY2-SHA256-RSA2048") {
		t.Errorf("expected WECHATPAY2-SHA256-RSA2048 Authorization header, got: %q", fs.requests[0].Auth)
	}
}

// --- TransactionsH5 tests ---

func TestTransactionsH5_Success(t *testing.T) {
	c, fs, srv := newTestMerchantClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{"h5_url":"https://wx.tenpay.com/cgi-bin/mmpaywap?prepay_id=test"}`)
	}

	result, err := c.TransactionsH5(context.Background(), testOrder())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.H5Url != "https://wx.tenpay.com/cgi-bin/mmpaywap?prepay_id=test" {
		t.Errorf("expected h5_url, got %q", result.H5Url)
	}
}

func TestTransactionsH5_SetsAuthorizationHeader(t *testing.T) {
	c, fs, srv := newTestMerchantClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{"h5_url":"https://test.com"}`)
	}

	_, err := c.TransactionsH5(context.Background(), testOrder())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fs.requests) == 0 {
		t.Fatal("expected at least one request")
	}
	if !strings.HasPrefix(fs.requests[0].Auth, "WECHATPAY2-SHA256-RSA2048") {
		t.Errorf("expected WECHATPAY2-SHA256-RSA2048 Authorization header, got: %q", fs.requests[0].Auth)
	}
}

// --- Client builder chain tests ---

func TestNewWechatClient_DefaultBaseURL(t *testing.T) {
	c := NewWechatClient()
	if c.HTTP() == nil {
		t.Fatal("expected non-nil HTTP client")
	}
	if !strings.Contains(c.HTTP().BaseURL, "api.mch.weixin.qq.com") {
		t.Errorf("expected default base URL to contain api.mch.weixin.qq.com, got %q", c.HTTP().BaseURL)
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
	if c.Appid() != "appid" {
		t.Errorf("expected appid, got %q", c.Appid())
	}
	if c.Mchid() != "mchid" {
		t.Errorf("expected mchid, got %q", c.Mchid())
	}
	if c.CertificateNumber() != "cert_no" {
		t.Errorf("expected cert_no, got %q", c.CertificateNumber())
	}
	if c.PrivateKeyVal() != key {
		t.Error("expected private key to be set")
	}
}

func TestWechatClient_WithHttp(t *testing.T) {
	h := utils.NewHTTP("https://custom.example.com")
	c := NewWechatClient().WithHttp(h)
	if c.HTTP() != h {
		t.Error("expected HTTP() to be the injected instance")
	}
	if c.HTTP().BaseURL != "https://custom.example.com" {
		t.Errorf("expected custom base URL, got %q", c.HTTP().BaseURL)
	}
}

// --- ModifyTransactionsApp tests ---

func TestModifyTransactionsApp_Success(t *testing.T) {
	c, fs, srv := newTestMerchantClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{"prepay_id":"wx_modify_prepay_id"}`)
	}

	result, err := c.ModifyTransactionsApp(context.Background(), testOrder())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PrepayId != "wx_modify_prepay_id" {
		t.Errorf("expected prepay_id wx_modify_prepay_id, got %q", result.PrepayId)
	}
	// ModifyTransactionsApp copies the client's configured appid into the response.
	if result.AppId != c.Appid() {
		t.Errorf("expected AppId %q, got %q", c.Appid(), result.AppId)
	}
	if result.PackageValue != "Sign=WXPay" {
		t.Errorf("expected PackageValue Sign=WXPay, got %q", result.PackageValue)
	}
}

// --- ModifyTransactionsJsapi tests ---

func TestModifyTransactionsJsapi_Success(t *testing.T) {
	c, fs, srv := newTestMerchantClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{"prepay_id":"jsapi_modify_pid"}`)
	}

	result, err := c.ModifyTransactionsJsapi(context.Background(), testOrder())
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
	c, fs, srv := newTestMerchantClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{"transaction_id":"4200001234","out_trade_no":"ORD001","trade_state":"SUCCESS"}`)
	}

	result, err := c.QueryTransactionId(context.Background(), "4200001234")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestQueryTransactionId_SetsAuthorizationHeader(t *testing.T) {
	c, fs, srv := newTestMerchantClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{}`)
	}

	_, _ = c.QueryTransactionId(context.Background(), "txn123")
	if len(fs.requests) == 0 {
		t.Fatal("expected at least one request")
	}
	if !strings.HasPrefix(fs.requests[0].Auth, "WECHATPAY2-SHA256-RSA2048") {
		t.Errorf("expected WECHATPAY2-SHA256-RSA2048 Authorization header, got: %q", fs.requests[0].Auth)
	}
}

// --- QueryOutTradeNo tests ---

func TestQueryOutTradeNo_Success(t *testing.T) {
	c, fs, srv := newTestMerchantClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{"out_trade_no":"ORD001","trade_state":"SUCCESS"}`)
	}

	result, err := c.QueryOutTradeNo(context.Background(), "ORD001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestQueryOutTradeNo_SetsAuthorizationHeader(t *testing.T) {
	c, fs, srv := newTestMerchantClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{}`)
	}

	_, _ = c.QueryOutTradeNo(context.Background(), "ORD001")
	if len(fs.requests) == 0 {
		t.Fatal("expected at least one request")
	}
	if !strings.HasPrefix(fs.requests[0].Auth, "WECHATPAY2-SHA256-RSA2048") {
		t.Errorf("expected WECHATPAY2-SHA256-RSA2048 Authorization header, got: %q", fs.requests[0].Auth)
	}
}

// --- TransactionsClose tests ---

func TestTransactionsClose_Success(t *testing.T) {
	c, fs, srv := newTestMerchantClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return http.StatusNoContent, nil
	}

	err := c.TransactionsClose(context.Background(), "ORD001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTransactionsClose_SetsAuthorizationHeader(t *testing.T) {
	c, fs, srv := newTestMerchantClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return http.StatusNoContent, nil
	}

	_ = c.TransactionsClose(context.Background(), "ORD001")
	if len(fs.requests) == 0 {
		t.Fatal("expected at least one request")
	}
	if !strings.HasPrefix(fs.requests[0].Auth, "WECHATPAY2-SHA256-RSA2048") {
		t.Errorf("expected WECHATPAY2-SHA256-RSA2048 Authorization header, got: %q", fs.requests[0].Auth)
	}
}

// --- Refunds tests ---

func TestRefunds_Success(t *testing.T) {
	c, fs, srv := newTestMerchantClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{"refund_id":"REF001","out_refund_no":"REFNO001","status":"PROCESSING"}`)
	}

	refund := &types.Refunds{
		OutTradeNo:  "ORD001",
		OutRefundNo: "REFNO001",
		Amount:      &types.Amount{Total: 100, Currency: "CNY"},
	}
	result, err := c.Refunds(context.Background(), refund)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestRefunds_SetsAuthorizationHeader(t *testing.T) {
	c, fs, srv := newTestMerchantClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{"refund_id":"r1"}`)
	}

	refund := &types.Refunds{
		OutTradeNo:  "ORD001",
		OutRefundNo: "REFNO001",
		Amount:      &types.Amount{Total: 100, Currency: "CNY"},
	}
	_, _ = c.Refunds(context.Background(), refund)
	if len(fs.requests) == 0 {
		t.Fatal("expected at least one request")
	}
	if !strings.HasPrefix(fs.requests[0].Auth, "WECHATPAY2-SHA256-RSA2048") {
		t.Errorf("expected WECHATPAY2-SHA256-RSA2048 Authorization header, got: %q", fs.requests[0].Auth)
	}
}

// --- QueryRefunds tests ---

func TestQueryRefunds_Success(t *testing.T) {
	c, fs, srv := newTestMerchantClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{"refund_id":"REF001","out_refund_no":"REFNO001","status":"SUCCESS"}`)
	}

	result, err := c.QueryRefunds(context.Background(), "REFNO001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestQueryRefunds_SetsAuthorizationHeader(t *testing.T) {
	c, fs, srv := newTestMerchantClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{}`)
	}

	_, _ = c.QueryRefunds(context.Background(), "REFNO001")
	if len(fs.requests) == 0 {
		t.Fatal("expected at least one request")
	}
	if !strings.HasPrefix(fs.requests[0].Auth, "WECHATPAY2-SHA256-RSA2048") {
		t.Errorf("expected WECHATPAY2-SHA256-RSA2048 Authorization header, got: %q", fs.requests[0].Auth)
	}
}

// --- ApplyAbnormalRefund tests ---

func TestApplyAbnormalRefund_Success(t *testing.T) {
	c, fs, srv := newTestMerchantClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{"refund_id":"REF002","status":"PROCESSING"}`)
	}

	abnormal := &types.AbnormalRefund{
		OutRefundNo: "REFNO002",
	}
	result, err := c.ApplyAbnormalRefund(context.Background(), "REF001", abnormal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

// --- TradeBill tests ---

func TestTradeBill_Success(t *testing.T) {
	c, fs, srv := newTestMerchantClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{"download_url":"https://api.mch.weixin.qq.com/v3/billdownload/file","hash_value":"abc"}`)
	}

	quest := &types.TradeBillQuest{BillDate: "2024-01-01"}
	result, err := c.TradeBill(context.Background(), quest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestTradeBill_SetsAuthorizationHeader(t *testing.T) {
	c, fs, srv := newTestMerchantClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{}`)
	}

	_, _ = c.TradeBill(context.Background(), &types.TradeBillQuest{BillDate: "2024-01-01"})
	if len(fs.requests) == 0 {
		t.Fatal("expected at least one request")
	}
	if !strings.HasPrefix(fs.requests[0].Auth, "WECHATPAY2-SHA256-RSA2048") {
		t.Errorf("expected WECHATPAY2-SHA256-RSA2048 Authorization header, got: %q", fs.requests[0].Auth)
	}
}

// --- FundFlowBill tests ---

func TestFundFlowBill_Success(t *testing.T) {
	c, fs, srv := newTestMerchantClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{"download_url":"https://api.mch.weixin.qq.com/v3/billdownload/file","hash_value":"def"}`)
	}

	quest := &types.FundsBillQuest{BillDate: "2024-01-01"}
	result, err := c.FundFlowBill(context.Background(), quest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestFundFlowBill_SetsAuthorizationHeader(t *testing.T) {
	c, fs, srv := newTestMerchantClient(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{}`)
	}

	_, _ = c.FundFlowBill(context.Background(), &types.FundsBillQuest{BillDate: "2024-01-01"})
	if len(fs.requests) == 0 {
		t.Fatal("expected at least one request")
	}
	if !strings.HasPrefix(fs.requests[0].Auth, "WECHATPAY2-SHA256-RSA2048") {
		t.Errorf("expected WECHATPAY2-SHA256-RSA2048 Authorization header, got: %q", fs.requests[0].Auth)
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

func TestParseNotification_NoSigHeaders(t *testing.T) {
	c := NewWechatClient()
	req := httptest.NewRequest("POST", "/notify", strings.NewReader(`{"id":"evt001"}`))
	// When signature headers are absent, verifyResponseSignature skips verification and returns nil.
	notify, err := c.ParseNotification(context.Background(), req, nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if notify == nil {
		t.Error("expected non-nil notify")
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

func TestParseRefundNotify_NoSigHeaders(t *testing.T) {
	c := NewWechatClient()
	req := httptest.NewRequest("POST", "/notify", strings.NewReader(`{"id":"evt001"}`))
	// When signature headers are absent, verifyResponseSignature skips verification and returns nil.
	notify, _, err := c.ParseRefundNotify(context.Background(), req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if notify == nil {
		t.Error("expected non-nil notify")
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
