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

func TestGenerateHashBasedString(t *testing.T) {
	s := GenerateHashBasedString(32)
	if len(s) != 32 {
		t.Errorf("length mismatch: %d", len(s))
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
