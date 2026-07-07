package console

import (
	"strings"
	"testing"

	"github.com/reeflective/readline"
	"github.com/spf13/cobra"
)

func TestCompleteHidesCarapaceCommand(t *testing.T) {
	c := New("test")
	root := &cobra.Command{Use: "root"}
	internal := &cobra.Command{Use: "_carapace"}
	root.AddCommand(internal, &cobra.Command{Use: "visible"})
	c.activeMenu().Command = root

	comps := c.complete(nil, 0)

	if !internal.Hidden {
		t.Fatal("_carapace command was not hidden")
	}

	for _, value := range completionValues(comps) {
		if strings.TrimSpace(value) == "_carapace" {
			t.Fatalf("completion values include internal command: %v", completionValues(comps))
		}
	}
}

func completionValues(comps readline.Completions) []string {
	var values []string

	comps.EachValue(func(comp readline.Completion) readline.Completion {
		values = append(values, comp.Value)
		return comp
	})

	return values
}
