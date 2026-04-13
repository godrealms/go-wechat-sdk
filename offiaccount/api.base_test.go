package offiaccount

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

func newBaseTestClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	h := utils.NewHTTP(srv.URL, utils.WithTimeout(3*time.Second))
	return NewClient(context.Background(), &Config{AppId: "test", AppSecret: "secret"},
		WithHTTPClient(h),
		WithTokenSource(fixedToken{"FAKE_TOKEN"}),
	)
}

func TestGetCallbackIp_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"ip_list":["101.226.103.0/25","101.226.62.0/26"]}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	ips, err := c.GetCallbackIp(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ips) != 2 {
		t.Errorf("expected 2 IPs, got %d", len(ips))
	}
}

func TestGetApiDomainIP_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"ip_list":["203.205.254.0/24","203.205.226.0/24"]}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	ips, err := c.GetApiDomainIP(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ips) != 2 {
		t.Errorf("expected 2 IPs, got %d", len(ips))
	}
}

func TestCreateQRCode_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"ticket":"gQH47joAAAAAAAAAASxodHRwOi8vd2VpeGlu","expire_seconds":604800,"url":"http://weixin.qq.com/q/kZgfwMTm72WWPkovabbI"}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	req := &CreateQRCodeRequest{
		ActionName: "QR_SCENE",
		ActionInfo: ActionInfo{Scene: Scene{SceneID: 123}},
	}
	result, err := c.CreateQRCode(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Ticket == "" {
		t.Error("expected non-empty ticket")
	}
	if result.ExpireSeconds != 604800 {
		t.Errorf("expected 604800, got %d", result.ExpireSeconds)
	}
}

func TestCreateQRCode_NetworkError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close()
	c := newBaseTestClient(t, srv)
	_, err := c.CreateQRCode(context.Background(), &CreateQRCodeRequest{})
	if err == nil {
		t.Error("expected network error")
	}
}

func TestGetQRCodeURL(t *testing.T) {
	c := &Client{}
	ticket := "gQH47joAAA+special chars&="
	got := c.GetQRCodeURL(ticket)
	if !strings.HasPrefix(got, "https://mp.weixin.qq.com/cgi-bin/showqrcode?ticket=") {
		t.Errorf("unexpected URL prefix: %s", got)
	}
	if strings.Contains(got, " ") {
		t.Error("URL must not contain raw spaces")
	}
	if strings.Contains(got, "&=") {
		t.Error("URL must have ticket encoded")
	}
}
