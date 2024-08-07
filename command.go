package console

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	// CommandFilterKey should be used as a key to in a cobra.Annotation map.
	// The value will be used as a filter to disable commands when the console
	// calls the Filter("name") method on the console.
	// The string value will be comma-splitted, with each split being a filter.
	CommandFilterKey = "console-hidden"
)

// Commands is a simple function a root cobra command containing an arbitrary tree
// of subcommands, along with any behavior parameters normally found in cobra.
// This function is used by each menu to produce a new, blank command tree after
// each execution run, as well as each command completion invocation.
type Commands func() *cobra.Command

// SetCommands requires a function returning a tree of cobra commands to be used.
func (m *Menu) SetCommands(cmds Commands) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	m.cmds = cmds
}

// HideCommands - Commands, in addition to their menus, can be shown/hidden based
// on a filter string. For example, some commands applying to a Windows host might
// be scattered around different groups, but, having all the filter "windows".
// If "windows" is used as the argument here, all windows commands for the current
// menu are subsequently hidden, until ShowCommands("windows") is called.
func (c *Console) HideCommands(filters ...string) {
next:
	for _, filt := range filters {
		for _, filter := range c.filters {
			if filt == filter {
				continue next
			}
		}
		if filt != "" {
			c.filters = append(c.filters, filt)
		}
	}
}

// ShowCommands - Commands, in addition to their menus, can be shown/hidden based
// on a filter string. For example, some commands applying to a Windows host might
// be scattered around different groups, but, having all the filter "windows".
// Use this function if you have previously called HideCommands("filter") and want
// these commands to be available back under their respective menu.
func (c *Console) ShowCommands(filters ...string) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	updated := make([]string, 0)

	if len(filters) == 0 {
		c.filters = updated

		return
	}

next:
	for _, filt := range c.filters {
		for _, filter := range filters {
			if filt == filter {
				continue next
			}
		}
		updated = append(updated, filt)
	}

	c.filters = updated
}

// resetFlagsDefaults resets all flags to their default values.
//
// Slice flags accumulate per execution (and do not reset),
//
//	so we must reset them manually.
//
// Example:
//
//	Given cmd.Flags().StringSlice("comment", nil, "")
//	If you run a command with --comment "a" --comment "b" you will get
//	the expected [a, b] slice.
//
//	If you run a command again with no --comment flags, you will get
//	[a, b] again instead of an empty slice.
//
//	If you run the command again with --comment "c" --comment "d" flags,
//	you will get [a, b, c, d] instead of just [c, d].
func resetFlagsDefaults(target *cobra.Command) {
	target.Flags().VisitAll(func(flag *pflag.Flag) {
		flag.Changed = false
		switch value := flag.Value.(type) {
		case pflag.SliceValue:
			var res []string

			if len(flag.DefValue) > 0 && flag.DefValue != "[]" {
				res = append(res, flag.DefValue)
			}

			value.Replace(res)

		default:
			flag.Value.Set(flag.DefValue)
		}
	})
}
