package gonsole

import "github.com/jessevdk/go-flags"

// commandGroup - A group of commands, which might be by any motive: common domain,
// type, etc, as long as the group name is the same. In addition, commands in the same
// group have an additional string filter key, which can be used further refine which
// commands are hidden or not.
// Please see the Console.Hide(filter string) or Console.Show(filter string)
// By default, a "" filter name will mean available no matter the filter.
type commandGroup struct {
	Name string

	commandGenerators map[string]func() *flags.Command
	commandNames      []string
	commandFilters    map[string]string

	// All generated commands are structured in equivalent groups.
	commandDone map[string]*flags.Command
}

// GetCommandGroup - Get the group for a command.
func (c *Console) GetCommandGroup(cmd *flags.Command) string {
	// Sliver commands are searched for if we are in this context
	for name, group := range c.current.groups {
		for _, c := range group {
			if c.Name == cmd.Name {
				return name
			}
		}
	}
	return ""
}
