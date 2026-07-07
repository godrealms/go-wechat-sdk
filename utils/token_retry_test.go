package utils_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/godrealms/go-wechat-sdk/utils"
)

func TestIsTokenExpired(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"plain error", errors.New("boom"), false},
		{"40001 invalid", &fakeAPIError{code: 40001}, true},
		{"40014 malformed", &fakeAPIError{code: 40014}, true},
		{"42001 expired", &fakeAPIError{code: 42001}, true},
		{"42007 ticket expired", &fakeAPIError{code: 42007}, true},
		{"40003 other business error", &fakeAPIError{code: 40003}, false},
		{"zero errcode", &fakeAPIError{code: 0}, false},
		{"wrapped 40001", fmt.Errorf("channels: /x: %w", &fakeAPIError{code: 40001}), true},
		{"wrapped non-token", fmt.Errorf("channels: /x: %w", &fakeAPIError{code: 40003}), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.IsTokenExpired(tt.err); got != tt.want {
				t.Errorf("IsTokenExpired(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

// countingInvalidator records how many times Invalidate was called.
type countingInvalidator struct{ n int }

func (c *countingInvalidator) Invalidate() { c.n++ }

func TestDoWithTokenRetry(t *testing.T) {
	tokenErr := &fakeAPIError{code: 40001, msg: "invalid credential"}
	otherErr := errors.New("network down")

	t.Run("success on first try: no retry, no invalidate", func(t *testing.T) {
		inv := &countingInvalidator{}
		calls := 0
		err := utils.DoWithTokenRetry(inv, func() error {
			calls++
			return nil
		})
		if err != nil {
			t.Errorf("err = %v, want nil", err)
		}
		if calls != 1 {
			t.Errorf("fn called %d times, want 1", calls)
		}
		if inv.n != 0 {
			t.Errorf("Invalidate called %d times, want 0", inv.n)
		}
	})

	t.Run("token-expired then success: retry once, invalidate once", func(t *testing.T) {
		inv := &countingInvalidator{}
		calls := 0
		err := utils.DoWithTokenRetry(inv, func() error {
			calls++
			if calls == 1 {
				return tokenErr
			}
			return nil
		})
		if err != nil {
			t.Errorf("err = %v, want nil after successful retry", err)
		}
		if calls != 2 {
			t.Errorf("fn called %d times, want 2", calls)
		}
		if inv.n != 1 {
			t.Errorf("Invalidate called %d times, want 1", inv.n)
		}
	})

	t.Run("token-expired twice: at most one retry, second error propagates", func(t *testing.T) {
		inv := &countingInvalidator{}
		calls := 0
		err := utils.DoWithTokenRetry(inv, func() error {
			calls++
			return tokenErr
		})
		if !errors.Is(err, tokenErr) {
			t.Errorf("err = %v, want the token error", err)
		}
		if calls != 2 {
			t.Errorf("fn called %d times, want 2 (no infinite retry)", calls)
		}
		if inv.n != 1 {
			t.Errorf("Invalidate called %d times, want 1", inv.n)
		}
	})

	t.Run("non-token error: no retry, no invalidate", func(t *testing.T) {
		inv := &countingInvalidator{}
		calls := 0
		err := utils.DoWithTokenRetry(inv, func() error {
			calls++
			return otherErr
		})
		if !errors.Is(err, otherErr) {
			t.Errorf("err = %v, want the original error", err)
		}
		if calls != 1 {
			t.Errorf("fn called %d times, want 1", calls)
		}
		if inv.n != 0 {
			t.Errorf("Invalidate called %d times, want 0", inv.n)
		}
	})

	t.Run("nil invalidator: no retry even on token error", func(t *testing.T) {
		calls := 0
		err := utils.DoWithTokenRetry(nil, func() error {
			calls++
			return tokenErr
		})
		if !errors.Is(err, tokenErr) {
			t.Errorf("err = %v, want the token error", err)
		}
		if calls != 1 {
			t.Errorf("fn called %d times, want 1 (nil inv disables retry)", calls)
		}
	})
}
