package isv

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Fake 43-char EncodingAESKey for tests.
const testAESKey = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQ"

func testConfig() Config {
	return Config{
		SuiteID:        "suite1",
		SuiteSecret:    "secret1",
		Token:          "TOKEN",
		EncodingAESKey: testAESKey,
	}
}

func timePlus(sec int) time.Time {
	return time.Now().Add(time.Duration(sec) * time.Second)
}

func TestNewClient_RequiredFields(t *testing.T) {
	cases := []struct {
		name string
		mut  func(*Config)
	}{
		{"suite id missing", func(c *Config) { c.SuiteID = "" }},
		{"suite secret missing", func(c *Config) { c.SuiteSecret = "" }},
		{"token missing", func(c *Config) { c.Token = "" }},
		{"aes key missing", func(c *Config) { c.EncodingAESKey = "" }},
		{"aes key wrong length", func(c *Config) { c.EncodingAESKey = "tooShort" }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cfg := testConfig()
			c.mut(&cfg)
			if _, err := NewClient(cfg); err == nil {
				t.Fatal("want error")
			}
		})
	}
}

func TestNewClient_ProviderPartial(t *testing.T) {
	cfg := testConfig()
	cfg.ProviderCorpID = "wx1"
	// missing ProviderSecret → error
	if _, err := NewClient(cfg); err == nil {
		t.Fatal("want error for partial provider config")
	}
	cfg.ProviderSecret = "psecret"
	if _, err := NewClient(cfg); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestNewClient_Options(t *testing.T) {
	customStore := NewMemoryStore()
	customHTTP := &http.Client{}
	c, err := NewClient(testConfig(),
		WithStore(customStore),
		WithHTTPClient(customHTTP),
		WithBaseURL("https://example.test"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if c.store != customStore {
		t.Error("store not applied")
	}
	if c.http != customHTTP {
		t.Error("http client not applied")
	}
	if c.baseURL != "https://example.test" {
		t.Error("base url not applied")
	}
}

func TestNewClient_Defaults(t *testing.T) {
	c, err := NewClient(testConfig())
	if err != nil {
		t.Fatal(err)
	}
	if c.store == nil {
		t.Error("default store missing")
	}
	if c.http == nil {
		t.Error("default http client missing")
	}
	if c.baseURL != "https://qyapi.weixin.qq.com" {
		t.Errorf("default baseURL wrong: %q", c.baseURL)
	}
}

func TestDoPost_WeixinError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"errcode": 42001,
			"errmsg":  "access_token expired",
		})
	}))
	defer srv.Close()

	c, err := NewClient(testConfig(), WithBaseURL(srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	// pre-seed a suite token so doPost has something to inject
	_ = c.store.PutSuiteToken(context.Background(), "suite1", "TOK", timePlus(3600))

	var out struct{}
	err = c.doPost(context.Background(), "/cgi-bin/whatever", map[string]string{"x": "y"}, &out)
	var we *WeixinError
	if !errors.As(err, &we) || we.ErrCode != 42001 {
		t.Fatalf("want *WeixinError 42001, got %v", err)
	}
}

func TestDoPost_AttachesTokenQuery(t *testing.T) {
	var gotRawQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRawQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()

	c, err := NewClient(testConfig(), WithBaseURL(srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	_ = c.store.PutSuiteToken(context.Background(), "suite1", "TOK42", timePlus(3600))

	var out struct{}
	if err := c.doPost(context.Background(), "/cgi-bin/test", map[string]string{}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotRawQuery, "suite_access_token=TOK42") {
		t.Fatalf("query missing token: %q", gotRawQuery)
	}
}
