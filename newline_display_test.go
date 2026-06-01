package console

import (
	"io"
	"os"
	"testing"
)

// captureStdout redirects os.Stdout for the duration of fn and returns what was
// written. The display functions print via fmt.Println, which targets
// os.Stdout directly.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}

	os.Stdout = w
	fn()
	os.Stdout = orig

	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	_ = r.Close()

	return string(out)
}

func TestDisplayNewlineMatrix(t *testing.T) {
	// A newline is printed iff: enabled && (whenEmpty || input is non-empty).
	cases := []struct {
		name        string
		enabled     bool
		whenEmpty   bool
		input       string
		wantNewline bool
	}{
		{"disabled/empty", false, false, "", false},
		{"disabled/nonempty", false, false, "cmd", false},
		{"disabled/whenEmpty/nonempty", false, true, "cmd", false},

		{"enabled/nonempty", true, false, "cmd", true},
		{"enabled/empty", true, false, "", false},
		{"enabled/spaces-are-empty", true, false, "   \t ", false},

		{"enabled/whenEmpty/empty", true, true, "", true},
		{"enabled/whenEmpty/nonempty", true, true, "cmd", true},
		{"enabled/whenEmpty/spaces", true, true, "   \t ", true},
	}

	for _, tc := range cases {
		want := ""
		if tc.wantNewline {
			want = "\n"
		}

		t.Run("pre/"+tc.name, func(t *testing.T) {
			c := New("test")
			c.NewlineBefore = tc.enabled
			c.NewlineWhenEmpty = tc.whenEmpty

			got := captureStdout(t, func() { c.displayPreRun(tc.input) })
			if got != want {
				t.Fatalf("displayPreRun(%q) printed %q, want %q", tc.input, got, want)
			}
		})

		t.Run("post/"+tc.name, func(t *testing.T) {
			c := New("test")
			c.NewlineAfter = tc.enabled
			c.NewlineWhenEmpty = tc.whenEmpty

			got := captureStdout(t, func() { c.displayPostRun(tc.input) })
			if got != want {
				t.Fatalf("displayPostRun(%q) printed %q, want %q", tc.input, got, want)
			}
		})
	}
}

// TestDisplayNewlineMenuOverride checks that a per-menu newline override is
// honored by the display path even when the console default differs.
func TestDisplayNewlineMenuOverride(t *testing.T) {
	c := New("test")
	c.NewlineAfter = false // console default: off
	c.ActiveMenu().SetNewlineAfter(true)

	if got := captureStdout(t, func() { c.displayPostRun("cmd") }); got != "\n" {
		t.Fatalf("menu override on: displayPostRun printed %q, want %q", got, "\n")
	}

	// And the inverse: console on, menu override off.
	c.NewlineAfter = true
	c.ActiveMenu().SetNewlineAfter(false)

	if got := captureStdout(t, func() { c.displayPostRun("cmd") }); got != "" {
		t.Fatalf("menu override off: displayPostRun printed %q, want %q", got, "")
	}
}
