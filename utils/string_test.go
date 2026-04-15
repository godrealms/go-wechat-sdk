package utils

import (
	"strings"
	"testing"
)

func TestRandomString_LengthAndPrefix(t *testing.T) {
	s := RandomString(10, "pre_")
	if !strings.HasPrefix(s, "pre_") {
		t.Errorf("missing prefix: %s", s)
	}
	if len(s)-len("pre_") != 10 {
		t.Errorf("unexpected length: %d", len(s)-len("pre_"))
	}
}

func TestRandomString_DefaultLength(t *testing.T) {
	s := RandomString(0, "")
	if len(s) != 6 {
		t.Errorf("expect default length 6, got %d", len(s))
	}
}

func TestRandomString_NotPredictable(t *testing.T) {
	seen := map[string]struct{}{}
	for i := 0; i < 100; i++ {
		s := RandomString(16, "")
		if _, ok := seen[s]; ok {
			t.Fatalf("duplicate random string in 100 trials: %s", s)
		}
		seen[s] = struct{}{}
	}
}

func TestGenerateNonceString(t *testing.T) {
	s := GenerateNonceString(32)
	if len(s) != 32 {
		t.Errorf("length mismatch: %d", len(s))
	}
	const allowed = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	for _, r := range s {
		if !strings.ContainsRune(allowed, r) {
			t.Errorf("unexpected char %q in %q", r, s)
		}
	}
}

func TestGenerateNonceString_DefaultLength(t *testing.T) {
	if got := len(GenerateNonceString(0)); got != 32 {
		t.Errorf("expect default length 32, got %d", got)
	}
	if got := len(GenerateNonceString(-1)); got != 32 {
		t.Errorf("expect default length 32 for negative, got %d", got)
	}
}

func TestGenerateNonceString_NotPredictable(t *testing.T) {
	seen := map[string]struct{}{}
	for i := 0; i < 200; i++ {
		s := GenerateNonceString(32)
		if _, ok := seen[s]; ok {
			t.Fatalf("duplicate nonce in 200 trials: %s", s)
		}
		seen[s] = struct{}{}
	}
}

// TestGenerateHashBasedString_DeprecatedAlias verifies the deprecated alias
// still works AND produces values with the same charset/length contract as
// GenerateNonceString — so any future accidental divergence (e.g. someone
// inlining a stale implementation into the alias) would be caught here.
func TestGenerateHashBasedString_DeprecatedAlias(t *testing.T) {
	const allowed = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	s := GenerateHashBasedString(32)
	if len(s) != 32 {
		t.Errorf("length mismatch: %d", len(s))
	}
	for _, r := range s {
		if !strings.ContainsRune(allowed, r) {
			t.Errorf("alias produced unexpected char %q in %q", r, s)
		}
	}
	// Default-length contract must also match.
	if got := len(GenerateHashBasedString(0)); got != 32 {
		t.Errorf("alias default length: got %d, want 32", got)
	}
}

func TestShuffleString(t *testing.T) {
	in := "abcdefghij"
	out := ShuffleString(in)
	if len(out) != len(in) {
		t.Errorf("length mismatch")
	}
	for _, c := range in {
		if !strings.ContainsRune(out, c) {
			t.Errorf("missing rune: %c", c)
		}
	}
}
