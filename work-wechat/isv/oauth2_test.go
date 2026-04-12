package isv

import (
	"net/url"
	"strings"
	"testing"
)

func TestOAuth2URL_Default(t *testing.T) {
	cfg := testConfig()
	c, err := NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	got := c.OAuth2URL("https://app.example.com/cb?x=1&y=2", "STATE1")

	// Split off the fragment before parsing.
	if !strings.HasSuffix(got, "#wechat_redirect") {
		t.Errorf("missing fragment: %q", got)
	}
	base := strings.TrimSuffix(got, "#wechat_redirect")

	u, err := url.Parse(base)
	if err != nil {
		t.Fatal(err)
	}
	if u.Host != "open.weixin.qq.com" {
		t.Errorf("host: %q", u.Host)
	}
	if u.Path != "/connect/oauth2/authorize" {
		t.Errorf("path: %q", u.Path)
	}
	q := u.Query()
	if q.Get("appid") != cfg.SuiteID {
		t.Errorf("appid: %q", q.Get("appid"))
	}
	if q.Get("redirect_uri") != "https://app.example.com/cb?x=1&y=2" {
		t.Errorf("redirect_uri: %q", q.Get("redirect_uri"))
	}
	if q.Get("response_type") != "code" {
		t.Errorf("response_type: %q", q.Get("response_type"))
	}
	if q.Get("scope") != "snsapi_privateinfo" {
		t.Errorf("scope: %q", q.Get("scope"))
	}
	if q.Get("state") != "STATE1" {
		t.Errorf("state: %q", q.Get("state"))
	}
	if q.Get("agentid") != "" {
		t.Errorf("agentid should be absent by default, got %q", q.Get("agentid"))
	}
}

func TestOAuth2URL_WithOptions(t *testing.T) {
	cfg := testConfig()
	c, err := NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	got := c.OAuth2URL(
		"https://app.example.com/cb",
		"STATE2",
		WithOAuth2Scope("snsapi_base"),
		WithOAuth2AgentID(1000001),
	)
	base := strings.TrimSuffix(got, "#wechat_redirect")
	u, err := url.Parse(base)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	if q.Get("scope") != "snsapi_base" {
		t.Errorf("scope: %q", q.Get("scope"))
	}
	if q.Get("agentid") != "1000001" {
		t.Errorf("agentid: %q", q.Get("agentid"))
	}
}
