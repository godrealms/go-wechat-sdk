package utils

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TokenFetcher is called by TokenCache when a fresh token is needed.
// It must return the raw token string and the server-reported expires_in
// (in seconds). The caller is responsible for issuing the HTTP request and
// mapping any errcode into a typed error before returning.
type TokenFetcher func(ctx context.Context) (token string, expiresIn int64, err error)

// TokenCache provides goroutine-safe, in-process access_token caching with
// double-checked locking and a configurable TTL safety margin.
//
// It is designed for the six standard WeChat packages (channels, mini-program,
// mini-game, mini-store, aispeech, xiaowei) that all share the same caching
// pattern: RWMutex, 60 s safety window, 60 s TTL floor.
//
// For packages with different caching strategies (oplatform, work-wechat/isv)
// or Store-backed caching, use their own implementations.
type TokenCache struct {
	pkg   string       // package name, used in error messages
	fetch TokenFetcher // called to obtain a fresh token

	mu          sync.RWMutex
	accessToken string
	expiresAt   time.Time
}

// NewTokenCache constructs a TokenCache.
//   - pkg is the package name used in error messages (e.g. "channels").
//   - fetch is called when the cached token is missing or expired.
func NewTokenCache(pkg string, fetch TokenFetcher) *TokenCache {
	return &TokenCache{pkg: pkg, fetch: fetch}
}

// Get returns a cached access_token, refreshing it when fewer than 60 seconds
// remain before expiry. Uses a read-lock fast path with a write-lock fallback
// (double-checked locking) to minimise contention under high concurrency.
func (tc *TokenCache) Get(ctx context.Context) (string, error) {
	// Fast path: read lock.
	tc.mu.RLock()
	if tc.accessToken != "" && time.Now().Before(tc.expiresAt) {
		t := tc.accessToken
		tc.mu.RUnlock()
		return t, nil
	}
	tc.mu.RUnlock()

	// Slow path: write lock + double-check.
	tc.mu.Lock()
	defer tc.mu.Unlock()
	if tc.accessToken != "" && time.Now().Before(tc.expiresAt) {
		return tc.accessToken, nil
	}

	token, expiresIn, err := tc.fetch(ctx)
	if err != nil {
		return "", err
	}
	if token == "" {
		return "", fmt.Errorf("%s: empty access_token", tc.pkg)
	}

	// Clamp TTL: subtract a 60 s safety window so we refresh before the
	// server-side token actually expires; floor at 60 s so a hostile or
	// malformed upstream cannot cause a refresh storm.
	ttl := expiresIn - 60
	if ttl < 60 {
		ttl = 60
	}
	tc.accessToken = token
	tc.expiresAt = time.Now().Add(time.Duration(ttl) * time.Second)
	return tc.accessToken, nil
}
