package mini_program

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

// newValidationClient returns a Client whose HTTP layer points at a server
// that fails the test if it's ever called. Validation must reject bad input
// before any network I/O happens (so we don't leak credentials into a bad
// request or let an untrusted field reach upstream).
func newValidationClient(t *testing.T) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("validation should have short-circuited before HTTP: %s %s", r.Method, r.URL.Path)
	}))
	t.Cleanup(srv.Close)

	c, err := NewClient(Config{AppId: "wx", AppSecret: "sec"},
		WithHTTP(utils.NewHTTP(srv.URL, utils.WithTimeout(time.Second*3))))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func assertValidationErr(t *testing.T, err error, wantSubstr string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}
	if !strings.Contains(err.Error(), wantSubstr) {
		t.Errorf("error %q does not contain %q", err.Error(), wantSubstr)
	}
}

// --- phone.go ---

func TestGetPhoneNumber_EmptyCode(t *testing.T) {
	c := newValidationClient(t)
	_, err := c.GetPhoneNumber(context.Background(), "")
	assertValidationErr(t, err, "code is required")
}

// --- media.go ---

func TestUploadTempMedia_InvalidMediaType(t *testing.T) {
	c := newValidationClient(t)
	for _, bad := range []string{"", "png", "audio", "file"} {
		_, err := c.UploadTempMedia(context.Background(), bad, "x.png", strings.NewReader("data"))
		assertValidationErr(t, err, "mediaType must be one of")
	}
}

func TestUploadTempMedia_EmptyFileName(t *testing.T) {
	c := newValidationClient(t)
	_, err := c.UploadTempMedia(context.Background(), "image", "", strings.NewReader("data"))
	assertValidationErr(t, err, "fileName is required")
}

func TestUploadTempMedia_NilReader(t *testing.T) {
	c := newValidationClient(t)
	_, err := c.UploadTempMedia(context.Background(), "image", "x.png", nil)
	assertValidationErr(t, err, "fileData is required")
}

func TestGetTempMedia_EmptyMediaID(t *testing.T) {
	c := newValidationClient(t)
	_, err := c.GetTempMedia(context.Background(), "")
	assertValidationErr(t, err, "mediaID is required")
}

// --- urlscheme.go ---

func TestGenerateScheme_NilReq(t *testing.T) {
	c := newValidationClient(t)
	_, err := c.GenerateScheme(context.Background(), nil)
	assertValidationErr(t, err, "req is required")
}

func TestGenerateUrlLink_NilReq(t *testing.T) {
	c := newValidationClient(t)
	_, err := c.GenerateUrlLink(context.Background(), nil)
	assertValidationErr(t, err, "req is required")
}

func TestGenerateShortLink_NilReq(t *testing.T) {
	c := newValidationClient(t)
	_, err := c.GenerateShortLink(context.Background(), nil)
	assertValidationErr(t, err, "req is required")
}

func TestGenerateShortLink_EmptyPageURL(t *testing.T) {
	c := newValidationClient(t)
	_, err := c.GenerateShortLink(context.Background(), &GenerateShortLinkReq{})
	assertValidationErr(t, err, "PageURL is required")
}

// --- wxacode.go ---

func TestGetWxaCode_NilReq(t *testing.T) {
	c := newValidationClient(t)
	_, err := c.GetWxaCode(context.Background(), nil)
	assertValidationErr(t, err, "req is required")
}

func TestGetWxaCode_EmptyPath(t *testing.T) {
	c := newValidationClient(t)
	_, err := c.GetWxaCode(context.Background(), &GetWxaCodeReq{})
	assertValidationErr(t, err, "Path is required")
}

func TestGetWxaCodeUnlimit_NilReq(t *testing.T) {
	c := newValidationClient(t)
	_, err := c.GetWxaCodeUnlimit(context.Background(), nil)
	assertValidationErr(t, err, "req is required")
}

func TestGetWxaCodeUnlimit_EmptyScene(t *testing.T) {
	c := newValidationClient(t)
	_, err := c.GetWxaCodeUnlimit(context.Background(), &GetWxaCodeUnlimitReq{})
	assertValidationErr(t, err, "Scene is required")
}

func TestCreateQRCode_NilReq(t *testing.T) {
	c := newValidationClient(t)
	_, err := c.CreateQRCode(context.Background(), nil)
	assertValidationErr(t, err, "req is required")
}

func TestCreateQRCode_EmptyPath(t *testing.T) {
	c := newValidationClient(t)
	_, err := c.CreateQRCode(context.Background(), &CreateQRCodeReq{})
	assertValidationErr(t, err, "Path is required")
}

// --- security.go ---

func TestMsgSecCheck_NilReq(t *testing.T) {
	c := newValidationClient(t)
	_, err := c.MsgSecCheck(context.Background(), nil)
	assertValidationErr(t, err, "req is required")
}

func TestMsgSecCheck_EmptyContent(t *testing.T) {
	c := newValidationClient(t)
	_, err := c.MsgSecCheck(context.Background(), &MsgSecCheckReq{OpenID: "o"})
	assertValidationErr(t, err, "Content is required")
}

func TestMsgSecCheck_EmptyOpenID(t *testing.T) {
	c := newValidationClient(t)
	_, err := c.MsgSecCheck(context.Background(), &MsgSecCheckReq{Content: "hello"})
	assertValidationErr(t, err, "OpenID is required")
}

func TestMediaCheckAsync_NilReq(t *testing.T) {
	c := newValidationClient(t)
	_, err := c.MediaCheckAsync(context.Background(), nil)
	assertValidationErr(t, err, "req is required")
}

func TestMediaCheckAsync_EmptyMediaURL(t *testing.T) {
	c := newValidationClient(t)
	_, err := c.MediaCheckAsync(context.Background(), &MediaCheckAsyncReq{OpenID: "o"})
	assertValidationErr(t, err, "MediaURL is required")
}

func TestMediaCheckAsync_EmptyOpenID(t *testing.T) {
	c := newValidationClient(t)
	_, err := c.MediaCheckAsync(context.Background(), &MediaCheckAsyncReq{MediaURL: "https://x/y.png"})
	assertValidationErr(t, err, "OpenID is required")
}

// --- analysis.go ---

func TestGetDailySummary_EmptyDates(t *testing.T) {
	c := newValidationClient(t)
	for _, tc := range []struct{ begin, end string }{
		{"", ""}, {"20240101", ""}, {"", "20240101"},
	} {
		_, err := c.GetDailySummary(context.Background(), tc.begin, tc.end)
		assertValidationErr(t, err, "beginDate and endDate are required")
	}
}

func TestGetVisitPage_EmptyDates(t *testing.T) {
	c := newValidationClient(t)
	_, err := c.GetVisitPage(context.Background(), "", "")
	assertValidationErr(t, err, "beginDate and endDate are required")
}

func TestGetDailyVisitTrend_EmptyDates(t *testing.T) {
	c := newValidationClient(t)
	_, err := c.GetDailyVisitTrend(context.Background(), "", "")
	assertValidationErr(t, err, "beginDate and endDate are required")
}

// --- client.go ---

func TestSendSubscribeMessage_NilBody(t *testing.T) {
	c := newValidationClient(t)
	err := c.SendSubscribeMessage(context.Background(), nil)
	assertValidationErr(t, err, "body is required")
}

func TestCode2Session_EmptyJSCode(t *testing.T) {
	c := newValidationClient(t)
	_, err := c.Code2Session(context.Background(), "")
	assertValidationErr(t, err, "Code2Session: jsCode is required")
}
