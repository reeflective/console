package console

import (
	"reflect"
	"testing"

	"github.com/spf13/cobra"
)

// buildFilterTree returns a small command tree:
//
//	root
//	└── net   (filter: "windows")
//	    └── scan   (no annotations -> inherits from parent)
//
// plus a standalone, unannotated "free" command.
func buildFilterTree() (net, scan, free *cobra.Command) {
	root := &cobra.Command{Use: "root"}
	net = &cobra.Command{Use: "net", Annotations: map[string]string{CommandFilterKey: "windows"}}
	scan = &cobra.Command{Use: "scan"}
	free = &cobra.Command{Use: "free"}

	root.AddCommand(net, free)
	net.AddCommand(scan)

	return net, scan, free
}

func TestActiveFiltersFor(t *testing.T) {
	c := New("test")
	menu := c.ActiveMenu()
	net, scan, free := buildFilterTree()

	// No filter active yet: nothing is filtered, even annotated commands.
	if got := menu.ActiveFiltersFor(net); len(got) != 0 {
		t.Fatalf("before HideCommands: ActiveFiltersFor(net) = %q, want none", got)
	}

	// Activate the "windows" filter.
	c.HideCommands("windows")

	if got := menu.ActiveFiltersFor(net); !reflect.DeepEqual(got, []string{"windows"}) {
		t.Fatalf("ActiveFiltersFor(net) = %q, want [windows]", got)
	}

	// A child with no annotations inherits its parent's active filters.
	if got := menu.ActiveFiltersFor(scan); !reflect.DeepEqual(got, []string{"windows"}) {
		t.Fatalf("ActiveFiltersFor(scan) = %q, want [windows] (inherited)", got)
	}

	// An unrelated, unannotated command is never filtered.
	if got := menu.ActiveFiltersFor(free); len(got) != 0 {
		t.Fatalf("ActiveFiltersFor(free) = %q, want none", got)
	}

	// Removing the filter restores availability.
	c.ShowCommands("windows")
	if got := menu.ActiveFiltersFor(net); len(got) != 0 {
		t.Fatalf("after ShowCommands: ActiveFiltersFor(net) = %q, want none", got)
	}
}

func TestCheckIsAvailable(t *testing.T) {
	c := New("test")
	menu := c.ActiveMenu()
	net, scan, free := buildFilterTree()

	// A nil command is always available.
	if err := menu.CheckIsAvailable(nil); err != nil {
		t.Fatalf("CheckIsAvailable(nil) = %v, want nil", err)
	}

	c.HideCommands("windows")

	if err := menu.CheckIsAvailable(net); err == nil {
		t.Fatal("CheckIsAvailable(net) = nil, want error (command is filtered)")
	}
	if err := menu.CheckIsAvailable(scan); err == nil {
		t.Fatal("CheckIsAvailable(scan) = nil, want error (inherited filter)")
	}
	if err := menu.CheckIsAvailable(free); err != nil {
		t.Fatalf("CheckIsAvailable(free) = %v, want nil (not filtered)", err)
	}
}
