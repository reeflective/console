package console

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestHighlightCacheInvalidation(t *testing.T) {
	c := New("test")
	menu := c.ActiveMenu()
	menu.SetCommands(func() *cobra.Command {
		root := &cobra.Command{Use: "root"}
		root.AddCommand(&cobra.Command{Use: "net", Run: func(*cobra.Command, []string) {}})
		return root
	})
	menu.resetPreRun()

	in := []rune("net")
	first := c.highlightSyntax(in)

	cached := c.hlCache.Load()
	if cached == nil || cached.input != "net" {
		t.Fatalf("expected cache populated for %q, got %+v", "net", cached)
	}
	if cached.output != first {
		t.Fatalf("cached output %q != returned %q", cached.output, first)
	}

	// Same input is served from cache and yields the same result.
	if second := c.highlightSyntax(in); second != first {
		t.Fatalf("second highlight %q != first %q", second, first)
	}

	// Regenerating the command tree invalidates the cache.
	menu.resetPreRun()
	if c.hlCache.Load() != nil {
		t.Fatal("expected highlight cache cleared after resetPreRun")
	}
}
