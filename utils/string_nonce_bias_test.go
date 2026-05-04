package utils

import (
	"strings"
	"testing"
)

// TestGenerateNonceString_Length verifies the function returns the requested
// length (or 32 when length<=0), regardless of how many rejection-sampled
// bytes it had to discard.
func TestGenerateNonceString_Length(t *testing.T) {
	lengths := []int{0, 1, 10, 32, 64, 128, 1024}
	for _, l := range lengths {
		got := GenerateNonceString(l)
		want := l
		if want <= 0 {
			want = 32
		}
		if len(got) != want {
			t.Errorf("GenerateNonceString(%d) length = %d, want %d", l, len(got), want)
		}
	}
}

// TestGenerateNonceString_Charset verifies every byte is in the expected
// 62-character A-Za-z0-9 alphabet.
func TestGenerateNonceString_Charset(t *testing.T) {
	const allowed = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	for i := 0; i < 100; i++ {
		s := GenerateNonceString(64)
		for _, c := range s {
			if !strings.ContainsRune(allowed, c) {
				t.Fatalf("unexpected character %q in %q", c, s)
			}
		}
	}
}

// TestGenerateNonceString_ChiSquaredApprox runs a coarse uniformity check on
// the output distribution. Without rejection sampling, charset[0:8] would be
// drawn ~25% more often than charset[8:62] (because 256 mod 62 = 8).
//
// We sample N=20000*62 = 1.24M characters, expected count per character is
// 20000. With rejection sampling, the per-character distribution should have
// no detectable systematic offset between the first 8 and the remaining 54.
//
// We assert: max(count[0:8])/min(count[8:62]) is close to 1.0 (specifically,
// less than 1.10). With the buggy old implementation the ratio approached
// 1.25, so this catches a regression at any reasonable confidence.
func TestGenerateNonceString_ChiSquaredApprox(t *testing.T) {
	const allowed = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	const samples = 20000 // per-character expected count
	totalChars := samples * len(allowed)
	out := GenerateNonceString(totalChars)

	counts := make(map[byte]int)
	for i := 0; i < len(out); i++ {
		counts[out[i]]++
	}

	// First 8 bytes of charset are the ones that would be biased high
	// in the buggy implementation.
	maxFirst8 := 0
	for i := 0; i < 8; i++ {
		c := counts[allowed[i]]
		if c > maxFirst8 {
			maxFirst8 = c
		}
	}
	minRest := samples * 2 // initialise to something unreachable
	for i := 8; i < len(allowed); i++ {
		c := counts[allowed[i]]
		if c < minRest {
			minRest = c
		}
	}

	if minRest == 0 {
		t.Fatalf("some character never appeared (test broken or RNG broken)")
	}
	ratio := float64(maxFirst8) / float64(minRest)
	// Threshold: 1.10. Buggy old code would land ~1.25; rejection-sampled
	// code lands within statistical noise of 1.0.
	if ratio > 1.10 {
		t.Errorf("max(count[0:8])/min(count[8:62]) = %.3f, expected near 1.0 (rejection sampling failure?)", ratio)
	}
}
