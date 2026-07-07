// Package command provides pure utilities for manipulating cobra command
// trees: matching commands against console filters, hiding filtered or internal
// commands, resetting reused flag state, and locating the command targeted by a
// line of input. None of these functions depend on console state, so they can be
// tested in isolation; the root console package wraps them in its own methods.
package command

import (
	"encoding/csv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// FilterKey is the cobra annotation key whose comma-separated value marks a
// command with the filters that hide it. The console re-exports this as
// CommandFilterKey for application use.
const FilterKey = "console-hidden"

// ActiveFilters returns the console filters that cmd (or its nearest annotated
// ancestor) declares itself incompatible with. A non-empty result means the
// command is currently hidden/unavailable under the given console filters.
func ActiveFilters(cmd *cobra.Command, consoleFilters []string) []string {
	if cmd.Annotations == nil {
		if cmd.HasParent() {
			return ActiveFilters(cmd.Parent(), consoleFilters)
		}

		return nil
	}

	// Get the filters declared on the command.
	filterStr := cmd.Annotations[FilterKey]
	var filters []string

	for _, cmdFilter := range strings.Split(filterStr, ",") {
		for _, filter := range consoleFilters {
			if cmdFilter != "" && cmdFilter == filter {
				filters = append(filters, cmdFilter)
			}
		}
	}

	if len(filters) > 0 || !cmd.HasParent() {
		return filters
	}

	// Any parent that is hidden makes its whole subtree hidden also.
	return ActiveFilters(cmd.Parent(), consoleFilters)
}

// HideFiltered hides every subcommand of root that matches an active console
// filter, so it is not shown in help strings or offered as a completion.
// Commands already hidden are left untouched.
func HideFiltered(root *cobra.Command, consoleFilters []string) {
	for _, cmd := range root.Commands() {
		// Don't override commands if they are already hidden.
		if cmd.Hidden {
			continue
		}

		if filters := ActiveFilters(cmd, consoleFilters); len(filters) > 0 {
			cmd.Hidden = true
		}
	}
}

// HideCarapace recursively hides carapace's internal _carapace completion
// command so it is never offered as a normal user command.
func HideCarapace(root *cobra.Command) {
	if root == nil {
		return
	}

	for _, cmd := range root.Commands() {
		if cmd.Name() == "_carapace" {
			cmd.Hidden = true
			continue
		}

		HideCarapace(cmd)
	}
}

// ResetFlagsDefaults resets every flag on target back to its registered default
// value and clears its Changed state. Console reuses cobra command trees across
// completions and executions; when the application supplies a command tree
// directly (no generator), flag state parsed by an earlier run would otherwise
// leak into later ones.
func ResetFlagsDefaults(target *cobra.Command) {
	if target == nil {
		return
	}

	target.Flags().VisitAll(func(flag *pflag.Flag) {
		flag.Changed = false

		switch value := flag.Value.(type) {
		case pflag.SliceValue:
			_ = value.Replace(parseSliceDefault(flag.DefValue))
		default:
			_ = flag.Value.Set(flag.DefValue)
		}
	})
}

// parseSliceDefault turns a pflag slice flag's DefValue string representation
// (e.g. "[a,b]") back into the individual default elements.
func parseSliceDefault(defValue string) []string {
	if defValue == "" || defValue == "[]" {
		return nil
	}
	if strings.HasPrefix(defValue, "[") && strings.HasSuffix(defValue, "]") {
		defValue = defValue[1 : len(defValue)-1]
	}
	if defValue == "" {
		return nil
	}

	values, err := csv.NewReader(strings.NewReader(defValue)).Read()
	if err != nil {
		return []string{defValue}
	}

	return values
}

// ResetCompletionFlagState clears flag state left over from a previous
// completion or execution on a reused command tree, before carapace parses the
// current input. It restores the target command's flag defaults (shared with
// the execution path) and resets ArgsLenAtDash along the command's lineage.
func ResetCompletionFlagState(root *cobra.Command, args []string) {
	if root == nil {
		return
	}

	target := findCompletionTarget(root, args)

	// Force cobra to merge persistent/inherited flags into the full flag set
	// so ResetFlagsDefaults sees them all.
	_ = target.LocalFlags()

	ResetFlagsDefaults(target)
	resetArgsLenAtDash(target)
}

// resetArgsLenAtDash clears the "-- seen at index" bookkeeping on the target
// command and every parent, which a previous parse may have left set.
func resetArgsLenAtDash(target *cobra.Command) {
	for cmd := target; cmd != nil; cmd = cmd.Parent() {
		resetFlagSetArgsLenAtDash(cmd.Flags(), cmd.DisplayName())
		resetFlagSetArgsLenAtDash(cmd.PersistentFlags(), cmd.DisplayName())
	}
}

func resetFlagSetArgsLenAtDash(fs *pflag.FlagSet, name string) {
	if fs == nil {
		return
	}

	// FlagSet.Init resets argsLenAtDash to -1 without discarding registered
	// flags; it is the only exported way to clear that internal state.
	fs.Init(name, pflag.ContinueOnError)
}

// findCompletionTarget walks the command tree following the positional words in
// args, stopping at the first flag or "--", to locate the command being completed.
func findCompletionTarget(root *cobra.Command, args []string) *cobra.Command {
	cmd := root
	for _, arg := range args {
		if arg == "--" || strings.HasPrefix(arg, "-") {
			break
		}

		next := findSubcommand(cmd, arg)
		if next == nil {
			break
		}
		cmd = next
	}

	return cmd
}

func findSubcommand(cmd *cobra.Command, name string) *cobra.Command {
	if cmd == nil {
		return nil
	}

	for _, sub := range cmd.Commands() {
		if sub.Name() == name || sub.HasAlias(name) {
			return sub
		}
	}

	return nil
}
