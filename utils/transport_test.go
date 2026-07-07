package utils

import (
	"net/http"
	"testing"
	"time"
)

// TestNewHTTPClient_UsesSharedTunedTransport locks in the P1-1 connection-pool
// fix: every SDK client must share one transport whose per-host idle-connection
// limit is raised well above http.DefaultTransport's default of 2, otherwise
// concurrent WeChat calls degrade to a fresh TLS handshake per request.
func TestNewHTTPClient_UsesSharedTunedTransport(t *testing.T) {
	c1 := NewHTTPClient(30 * time.Second)
	c2 := NewHTTPClient(5 * time.Second)

	if c1.Timeout != 30*time.Second {
		t.Errorf("timeout not propagated: got %v, want 30s", c1.Timeout)
	}

	tr1, ok := c1.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("Transport is %T, want *http.Transport", c1.Transport)
	}

	// Both clients must share the same pooled transport instance.
	if c1.Transport != c2.Transport {
		t.Error("clients do not share a single transport; connection pooling is per-client")
	}

	if tr1.MaxIdleConnsPerHost != 64 {
		t.Errorf("MaxIdleConnsPerHost = %d, want 64 (DefaultTransport's 2 is the bottleneck this fixes)", tr1.MaxIdleConnsPerHost)
	}
	if tr1.MaxIdleConns != 100 {
		t.Errorf("MaxIdleConns = %d, want 100", tr1.MaxIdleConns)
	}
	if tr1.IdleConnTimeout != 90*time.Second {
		t.Errorf("IdleConnTimeout = %v, want 90s", tr1.IdleConnTimeout)
	}
	if !tr1.ForceAttemptHTTP2 {
		t.Error("ForceAttemptHTTP2 = false, want true")
	}

	// Cloning DefaultTransport must preserve proxy support so env proxies keep working.
	if tr1.Proxy == nil {
		t.Error("Proxy is nil; DefaultTransport's ProxyFromEnvironment was dropped")
	}
}

// TestNewHTTP_UsesSharedTransport ensures the HTTP helper (used by all product
// packages) wires the tuned transport rather than falling back to the stdlib default.
func TestNewHTTP_UsesSharedTransport(t *testing.T) {
	h := NewHTTP("https://api.weixin.qq.com")
	if h.Client.Transport != sdkTransport {
		t.Error("NewHTTP client does not use the shared sdkTransport")
	}
}
