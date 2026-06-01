package line

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestHighlightCommandAlias is a regression test for a bug where a command
// invoked through one of its aliases was never highlighted: the alias branch
// used to `break` out of the loop before reaching the highlight block.
func TestHighlightCommandAlias(t *testing.T) {
	root := &cobra.Command{Use: "app"}
	root.AddCommand(&cobra.Command{Use: "deploy host", Aliases: []string{"d", "dep"}})

	tests := []struct {
		name string
		arg  string
		want bool // whether arg should be highlighted as a command
	}{
		{"canonical name", "deploy", true},
		{"first alias", "d", true},
		{"second alias", "dep", true},
		{"unknown word", "nope", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			done, _ := HighlightCommand(nil, []string{tc.arg}, root, GreenFG)

			highlighted := len(done) > 0 && strings.Contains(done[0], GreenFG)
			if highlighted != tc.want {
				t.Fatalf("arg %q: highlighted=%v, want %v (got %q)", tc.arg, highlighted, tc.want, done)
			}
		})
	}
}

func TestHighlightCommandFlags(t *testing.T) {
	args := []string{"--verbose", "target", "-x", "value"}
	done, _ := HighlightCommandFlags(nil, args, BrightWhiteFG)

	if len(done) != len(args) {
		t.Fatalf("HighlightCommandFlags returned %d words, want %d: %q", len(done), len(args), done)
	}

	tests := []struct {
		idx           int
		shouldHighlit bool
		raw           string
	}{
		{0, true, "--verbose"},
		{1, false, "target"},
		{2, true, "-x"},
		{3, false, "value"},
	}

	for _, tc := range tests {
		got := done[tc.idx]
		colored := strings.Contains(got, BrightWhiteFG)
		if colored != tc.shouldHighlit {
			t.Errorf("word %q: highlighted=%v, want %v (got %q)", tc.raw, colored, tc.shouldHighlit, got)
		}
		if !strings.Contains(got, tc.raw) {
			t.Errorf("word %d: %q does not contain original %q", tc.idx, got, tc.raw)
		}
	}
}
