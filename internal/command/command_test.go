package command

import (
	"reflect"
	"testing"

	"github.com/spf13/cobra"
)

func filtered(use string, filters string) *cobra.Command {
	return &cobra.Command{
		Use:         use,
		Annotations: map[string]string{FilterKey: filters},
	}
}

func TestActiveFilters(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	win := filtered("win", "windows")
	multi := filtered("multi", "windows,admin")
	plain := &cobra.Command{Use: "plain"}
	root.AddCommand(win, multi, plain)

	if got := ActiveFilters(win, []string{"windows"}); !reflect.DeepEqual(got, []string{"windows"}) {
		t.Fatalf("win with windows active = %v, want [windows]", got)
	}
	if got := ActiveFilters(win, []string{"linux"}); len(got) != 0 {
		t.Fatalf("win with linux active = %v, want none", got)
	}
	if got := ActiveFilters(multi, []string{"admin"}); !reflect.DeepEqual(got, []string{"admin"}) {
		t.Fatalf("multi with admin active = %v, want [admin]", got)
	}
	if got := ActiveFilters(plain, []string{"windows"}); len(got) != 0 {
		t.Fatalf("plain command = %v, want none", got)
	}
}

// A command with no matching filter inherits its parent's filtered state, so a
// hidden subtree stays hidden regardless of the child's own annotations.
func TestActiveFiltersInheritsFromParent(t *testing.T) {
	parent := filtered("parent", "windows")
	child := &cobra.Command{Use: "child"}
	parent.AddCommand(child)

	if got := ActiveFilters(child, []string{"windows"}); !reflect.DeepEqual(got, []string{"windows"}) {
		t.Fatalf("child under filtered parent = %v, want [windows]", got)
	}
}

func TestHideFiltered(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	win := filtered("win", "windows")
	lin := filtered("lin", "linux")
	root.AddCommand(win, lin)

	HideFiltered(root, []string{"windows"})

	if !win.Hidden {
		t.Fatal("windows-filtered command should be hidden")
	}
	if lin.Hidden {
		t.Fatal("linux-filtered command should stay visible with only windows active")
	}
}

func TestHideCarapace(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	internal := &cobra.Command{Use: "_carapace"}
	sub := &cobra.Command{Use: "sub"}
	nestedInternal := &cobra.Command{Use: "_carapace"}
	sub.AddCommand(nestedInternal)
	root.AddCommand(internal, sub)

	HideCarapace(root)

	if !internal.Hidden {
		t.Fatal("top-level _carapace not hidden")
	}
	if !nestedInternal.Hidden {
		t.Fatal("nested _carapace not hidden")
	}
	if sub.Hidden {
		t.Fatal("regular command must not be hidden")
	}
}

func TestResetFlagsDefaults(t *testing.T) {
	cmd := &cobra.Command{Use: "serve"}
	cmd.Flags().Bool("verbose", false, "")
	cmd.Flags().StringSlice("item", []string{"base"}, "")

	if err := cmd.Flags().Set("verbose", "true"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("item", "one"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("item", "two"); err != nil {
		t.Fatal(err)
	}

	ResetFlagsDefaults(cmd)

	if cmd.Flags().Changed("verbose") || cmd.Flags().Changed("item") {
		t.Fatal("Changed state not cleared")
	}
	if v, _ := cmd.Flags().GetBool("verbose"); v {
		t.Fatal("verbose not reset to false")
	}
	if items, _ := cmd.Flags().GetStringSlice("item"); !reflect.DeepEqual(items, []string{"base"}) {
		t.Fatalf("item slice = %v, want [base]", items)
	}
}

func TestResetFlagsDefaultsNilSafe(t *testing.T) {
	ResetFlagsDefaults(nil) // must not panic
}

func TestParseSliceDefault(t *testing.T) {
	cases := map[string][]string{
		"":        nil,
		"[]":      nil,
		"[base]":  {"base"},
		"[a,b,c]": {"a", "b", "c"},
	}

	for in, want := range cases {
		if got := parseSliceDefault(in); !reflect.DeepEqual(got, want) {
			t.Fatalf("parseSliceDefault(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestResetCompletionFlagState(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	serve := &cobra.Command{Use: "serve"}
	serve.Flags().Bool("verbose", false, "")
	root.AddCommand(serve)

	if err := serve.Flags().Set("verbose", "true"); err != nil {
		t.Fatal(err)
	}
	if err := serve.Flags().Parse([]string{"--", "positional"}); err != nil {
		t.Fatal(err)
	}
	if serve.Flags().ArgsLenAtDash() < 0 {
		t.Fatal("setup did not set ArgsLenAtDash")
	}

	// Target the "serve" subcommand from the completion words.
	ResetCompletionFlagState(root, []string{"serve"})

	if serve.Flags().Changed("verbose") {
		t.Fatal("Changed state not cleared on completion reset")
	}
	if v, _ := serve.Flags().GetBool("verbose"); v {
		t.Fatal("verbose not reset")
	}
	if got := serve.Flags().ArgsLenAtDash(); got != -1 {
		t.Fatalf("ArgsLenAtDash = %d, want -1", got)
	}
}

func TestResetCompletionFlagStateNilSafe(t *testing.T) {
	ResetCompletionFlagState(nil, nil) // must not panic
}
