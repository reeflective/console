package console

import (
	"encoding/csv"
	"strings"

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
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.cmds = cmds
}

// HideCommands - Commands, in addition to their menus, can be shown/hidden based
// on a filter string. For example, some commands applying to a Windows host might
// be scattered around different groups, but, having all the filter "windows".
// If "windows" is used as the argument here, all windows commands for the current
// menu are subsequently hidden, until ShowCommands("windows") is called.
func (c *Console) HideCommands(filters ...string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

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
	c.mutex.Lock()
	defer c.mutex.Unlock()

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

// resetFlagsDefaults resets all flags on a command to their default values.
func resetFlagsDefaults(target *cobra.Command) {
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
