package offiaccount

import (
	"path/filepath"
	"runtime"
	"testing"
)

// TestFilepathBaseHandlesAllSeparators is a sanity check on the platform's
// filepath.Base behaviour. It documents (rather than enforces via SDK code)
// that on Windows both "/" and "\" are treated as separators, while on Unix
// only "/" is. The H10 fix replaced strings.Split(p, "/") with filepath.Base
// in three offiaccount upload-by-path methods so Windows callers no longer
// receive the full path string verbatim.
func TestFilepathBaseHandlesAllSeparators(t *testing.T) {
	if got := filepath.Base("/var/tmp/photo.jpg"); got != "photo.jpg" {
		t.Errorf("unix path: got %q want photo.jpg", got)
	}
	if runtime.GOOS == "windows" {
		if got := filepath.Base(`C:\Users\foo\bar.jpg`); got != "bar.jpg" {
			t.Errorf("windows path: got %q want bar.jpg", got)
		}
	}
}
