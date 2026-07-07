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

func TestCompleteResetsFlagDefaults(t *testing.T) {
	c := New("test")
	root := &cobra.Command{Use: "root"}
	cmd := &cobra.Command{Use: "serve"}
	cmd.Flags().Bool("verbose", false, "")
	root.AddCommand(cmd)
	c.activeMenu().Command = root

	if err := cmd.Flags().Set("verbose", "true"); err != nil {
		t.Fatal(err)
	}

	_ = c.complete([]rune("serve "), len("serve "))

	flag := cmd.Flags().Lookup("verbose")
	if flag == nil {
		t.Fatal("missing verbose flag")
	}
	if flag.Changed {
		t.Fatal("completion did not clear flag Changed state")
	}
	if flag.Value.String() != "false" {
		t.Fatalf("flag value = %q, want false", flag.Value.String())
	}
}

func TestCompleteResetsArgsLenAtDash(t *testing.T) {
	c := New("test")
	root := &cobra.Command{Use: "root"}
	cmd := &cobra.Command{Use: "serve"}
	cmd.Flags().Bool("verbose", false, "")
	root.AddCommand(cmd)
	c.activeMenu().Command = root

	if err := cmd.Flags().Parse([]string{"--", "positional"}); err != nil {
		t.Fatal(err)
	}
	if got := cmd.Flags().ArgsLenAtDash(); got < 0 {
		t.Fatalf("test setup did not set ArgsLenAtDash: %d", got)
	}

	_ = c.complete([]rune("serve "), len("serve "))

	if got := cmd.Flags().ArgsLenAtDash(); got != -1 {
		t.Fatalf("ArgsLenAtDash = %d, want -1", got)
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
