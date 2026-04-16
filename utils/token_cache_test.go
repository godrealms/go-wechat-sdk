package utils

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestTokenCache_CachesToken(t *testing.T) {
	var calls int
	tc := NewTokenCache("test", func(ctx context.Context) (string, int64, error) {
		calls++
		return "TOK", 7200, nil
	})
	for i := 0; i < 5; i++ {
		tok, err := tc.Get(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if tok != "TOK" {
			t.Errorf("got %q, want TOK", tok)
		}
	}
	if calls != 1 {
		t.Errorf("expected 1 fetch, got %d", calls)
	}
}

func TestTokenCache_FetchError(t *testing.T) {
	want := errors.New("boom")
	tc := NewTokenCache("test", func(ctx context.Context) (string, int64, error) {
		return "", 0, want
	})
	_, err := tc.Get(context.Background())
	if !errors.Is(err, want) {
		t.Errorf("got %v, want %v", err, want)
	}
}

func TestTokenCache_EmptyToken(t *testing.T) {
	tc := NewTokenCache("mypkg", func(ctx context.Context) (string, int64, error) {
		return "", 7200, nil
	})
	_, err := tc.Get(context.Background())
	if err == nil {
		t.Fatal("expected error for empty token")
	}
	if got := err.Error(); got != "mypkg: empty access_token" {
		t.Errorf("got %q", got)
	}
}

func TestTokenCache_TTLClamp_ZeroExpiresIn(t *testing.T) {
	var calls int
	tc := NewTokenCache("test", func(ctx context.Context) (string, int64, error) {
		calls++
		return "TOK", 0, nil // hostile: zero expires_in
	})
	// First call fetches.
	if _, err := tc.Get(context.Background()); err != nil {
		t.Fatal(err)
	}
	// Second call should still be cached (TTL floored to 60s).
	if _, err := tc.Get(context.Background()); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Errorf("expected 1 fetch with TTL clamp, got %d", calls)
	}
}

func TestTokenCache_TTLClamp_SmallExpiresIn(t *testing.T) {
	var calls int
	tc := NewTokenCache("test", func(ctx context.Context) (string, int64, error) {
		calls++
		return "TOK", 30, nil // tiny: 30s
	})
	if _, err := tc.Get(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := tc.Get(context.Background()); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Errorf("expected 1 fetch with TTL clamp, got %d", calls)
	}
}

func TestTokenCache_Concurrent(t *testing.T) {
	var fetchCount atomic.Int64
	tc := NewTokenCache("test", func(ctx context.Context) (string, int64, error) {
		fetchCount.Add(1)
		time.Sleep(10 * time.Millisecond) // simulate network latency
		return "TOK", 7200, nil
	})

	var wg sync.WaitGroup
	errs := make(chan error, 50)
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tok, err := tc.Get(context.Background())
			if err != nil {
				errs <- err
				return
			}
			if tok != "TOK" {
				errs <- errors.New("unexpected token: " + tok)
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
	// With double-checked locking, at most a small number of goroutines will
	// call fetch (those that acquired the write lock before the first fetch
	// completed). Practically this should be 1, but we allow a small margin.
	if n := fetchCount.Load(); n > 3 {
		t.Errorf("expected at most 3 fetches under contention, got %d", n)
	}
}

func TestTokenCache_RefreshAfterExpiry(t *testing.T) {
	var calls int
	tc := NewTokenCache("test", func(ctx context.Context) (string, int64, error) {
		calls++
		return "TOK", 7200, nil
	})
	if _, err := tc.Get(context.Background()); err != nil {
		t.Fatal(err)
	}
	// Simulate token expiry by directly manipulating expiresAt.
	tc.mu.Lock()
	tc.expiresAt = time.Now().Add(-1 * time.Second)
	tc.mu.Unlock()

	tok, err := tc.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if tok != "TOK" {
		t.Errorf("got %q, want TOK", tok)
	}
	if calls != 2 {
		t.Errorf("expected 2 fetches after expiry, got %d", calls)
	}
}
