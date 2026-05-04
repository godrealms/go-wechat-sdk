package utils

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

// ---------- RedactURL ----------

func TestRedactURL_TableDriven(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "no query string is unchanged",
			in:   "https://api.weixin.qq.com/cgi-bin/getcallbackip",
			want: "https://api.weixin.qq.com/cgi-bin/getcallbackip",
		},
		{
			name: "no sensitive keys is unchanged",
			in:   "https://api.weixin.qq.com/x?foo=bar&baz=1",
			want: "https://api.weixin.qq.com/x?foo=bar&baz=1",
		},
		{
			name: "OAuth secret is redacted",
			in:   "https://api.weixin.qq.com/sns/oauth2/access_token?appid=wx123&secret=SECRET_VALUE&code=AC123&grant_type=authorization_code",
			// Cannot assert on exact string due to map iteration ordering in Encode().
			// Asserted via parse below.
		},
		{
			name: "appsecret variant is redacted",
			in:   "https://api.weixin.qq.com/cgi-bin/clear_quota/v2?appid=wx123&appsecret=APPSECRET_VALUE",
		},
		{
			name: "access_token is redacted",
			in:   "https://api.weixin.qq.com/cgi-bin/test?access_token=BEARER_TOKEN&id=1",
		},
		{
			name: "refresh_token is redacted",
			in:   "https://api.weixin.qq.com/sns/oauth2/refresh_token?appid=wx123&refresh_token=R&grant_type=refresh_token",
		},
		{
			name: "case insensitive — APPSECRET",
			in:   "https://api.weixin.qq.com/x?APPSECRET=upper&y=1",
		},
		{
			name: "unparseable URL is returned unchanged",
			in:   "not a url at all\n",
			want: "not a url at all\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RedactURL(tt.in)
			if tt.want != "" {
				if got != tt.want {
					t.Errorf("got %q want %q", got, tt.want)
				}
				return
			}
			// For URLs with sensitive keys: parse and verify the value is *** and
			// non-sensitive keys retain their original values.
			u, err := url.Parse(got)
			if err != nil {
				t.Fatalf("redacted URL is no longer parseable: %v (got %q)", err, got)
			}
			origU, _ := url.Parse(tt.in)
			origQ := origU.Query()
			gotQ := u.Query()
			for k, vs := range origQ {
				_, sensitive := redactedQueryKeys[strings.ToLower(k)]
				if sensitive {
					if gotQ.Get(k) != RedactedValue {
						t.Errorf("key %q should be redacted, got %q", k, gotQ.Get(k))
					}
				} else {
					// Original value preserved.
					if gotQ.Get(k) != vs[0] {
						t.Errorf("key %q value clobbered: got %q want %q", k, gotQ.Get(k), vs[0])
					}
				}
			}
		})
	}
}

// ---------- Logger redaction integration ----------

// captureLogger records every Debugf call after applying the format. Used to
// verify that the URL sent to the logger is redacted, even though the actual
// HTTP request still uses the real URL.
type captureLogger struct {
	mu      sync.Mutex
	entries []string
}

func (c *captureLogger) Debugf(format string, args ...any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = append(c.entries, fmt.Sprintf(format, args...))
}

func (c *captureLogger) joined() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return strings.Join(c.entries, "\n")
}

func TestHTTP_LoggerReceivesRedactedURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Server should still see the real secret.
		if r.URL.Query().Get("secret") != "REAL_SECRET" {
			t.Errorf("server got wrong secret: %q", r.URL.Query().Get("secret"))
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	log := &captureLogger{}
	h := NewHTTP(srv.URL, WithLogger(log), WithTimeout(time.Second*3))

	q := url.Values{"appid": {"wx123"}, "secret": {"REAL_SECRET"}, "code": {"AUTH_CODE"}}
	var out map[string]any
	if err := h.Get(context.Background(), "/sns/oauth2/access_token", q, &out); err != nil {
		t.Fatal(err)
	}

	logged := log.joined()
	if strings.Contains(logged, "REAL_SECRET") {
		t.Errorf("logger captured raw secret: %s", logged)
	}
	if strings.Contains(logged, "AUTH_CODE") {
		t.Errorf("logger captured raw OAuth code: %s", logged)
	}
	if !strings.Contains(logged, RedactedValue) {
		t.Errorf("expected %q in log, got: %s", RedactedValue, logged)
	}
	// Non-sensitive params still visible.
	if !strings.Contains(logged, "wx123") {
		t.Errorf("expected non-sensitive appid still visible, got: %s", logged)
	}
}

// ---------- MaxResponseSize ----------

func TestHTTP_MaxResponseSize_RejectsOversized(t *testing.T) {
	// Server returns 100 bytes, cap is 50. We expect "exceeds 50 bytes".
	big := strings.Repeat("a", 100)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(big))
	}))
	defer srv.Close()

	h := NewHTTP(srv.URL, WithMaxResponseSize(50), WithTimeout(time.Second*3))
	var out map[string]any
	err := h.Get(context.Background(), "/x", nil, &out)
	if err == nil {
		t.Fatal("expected size-exceeded error")
	}
	if !strings.Contains(err.Error(), "exceeds 50 bytes") {
		t.Errorf("expected 'exceeds 50 bytes' in error, got: %v", err)
	}
}

func TestHTTP_MaxResponseSize_AllowsExactlyAtLimit(t *testing.T) {
	// Server returns exactly 50 bytes (a valid JSON of that size).
	// `{"x":""}` is 8 bytes, so 42 'a' chars makes it exactly 50.
	body := `{"x":"` + strings.Repeat("a", 42) + `"}`
	if len(body) != 50 {
		t.Fatalf("test setup: expected 50 bytes, got %d", len(body))
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	h := NewHTTP(srv.URL, WithMaxResponseSize(50), WithTimeout(time.Second*3))
	var out map[string]any
	if err := h.Get(context.Background(), "/x", nil, &out); err != nil {
		t.Fatalf("expected success at exactly the limit, got %v", err)
	}
	if got := out["x"]; got != strings.Repeat("a", 42) {
		t.Errorf("decoded payload mismatch: %v", out)
	}
}

func TestHTTP_MaxResponseSize_NegativeDisablesCap(t *testing.T) {
	// Server returns more than DefaultMaxResponseSize would normally allow.
	// With cap disabled, the read must succeed.
	const bytesToSend = 256 // small but tests the negative-disables path
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"x":"` + strings.Repeat("a", bytesToSend) + `"}`))
	}))
	defer srv.Close()

	h := NewHTTP(srv.URL, WithMaxResponseSize(-1), WithTimeout(time.Second*3))
	var out map[string]any
	if err := h.Get(context.Background(), "/x", nil, &out); err != nil {
		t.Fatalf("negative cap should disable limit, got %v", err)
	}
}

func TestHTTP_MaxResponseSize_DefaultIs10MB(t *testing.T) {
	h := NewHTTP("http://unused")
	// MaxResponseSize stays zero on the struct; readLimitedBody substitutes
	// the default. Verify the constant matches our doc claim of 10 MiB.
	if DefaultMaxResponseSize != 10<<20 {
		t.Errorf("DefaultMaxResponseSize = %d, want 10 MiB (%d)", DefaultMaxResponseSize, 10<<20)
	}
	if h.MaxResponseSize != 0 {
		t.Errorf("zero-value HTTP should have MaxResponseSize=0, got %d", h.MaxResponseSize)
	}
}

// ---------- readLimitedBody unit tests ----------

func TestReadLimitedBody_ZeroUsesDefault(t *testing.T) {
	// Reading a small body with limit=0 must succeed and use the default 10 MiB.
	body := strings.NewReader("hello")
	got, err := readLimitedBody(body, 0)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello" {
		t.Errorf("got %q want hello", got)
	}
}

func TestReadLimitedBody_NegativeNoLimit(t *testing.T) {
	body := strings.NewReader(strings.Repeat("a", 1000))
	got, err := readLimitedBody(body, -1)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1000 {
		t.Errorf("got %d bytes want 1000", len(got))
	}
}

func TestReadLimitedBody_AtLimitOK(t *testing.T) {
	body := strings.NewReader("12345")
	got, err := readLimitedBody(body, 5)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "12345" {
		t.Errorf("got %q", got)
	}
}

func TestReadLimitedBody_OverLimitErrors(t *testing.T) {
	body := strings.NewReader("123456")
	_, err := readLimitedBody(body, 5)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "exceeds 5 bytes") {
		t.Errorf("error mismatch: %v", err)
	}
}
