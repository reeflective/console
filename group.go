package gonsole

import (
	"fmt"

	"github.com/jessevdk/go-flags"
	"github.com/maxlandon/readline"
)

// commandGroup - A group of commands, which might be by any motive: common domain,
// type, etc, as long as the group name is the same. In addition, commands in the same
// group have an additional string filter key, which can be used further refine which
// commands are hidden or not.
// Please see the Console.Hide(filter string) or Console.Show(filter string)
// By default, a "" filter name will mean available no matter the filter.
type commandGroup struct {
	Name string
	cmds []*Command
}

// GetCommandGroup - Get the group for a command.
func (c *Console) GetCommandGroup(cmd *flags.Command) string {

	// Sliver commands are searched for if we are in this context
	for _, group := range c.current.groupsAltT {
		for _, c := range group.cmds {
			if c.Name == cmd.Name {
				// We don't return the name if the command is not generated
				if c.cmd != nil {
					return group.Name
				}
			}
		}
	}
	return ""
}

type Command struct {
	Name    string
	Short   string
	Long    string
	Context string
	Group   string
	Filters []string

	generator func() *flags.Command
	cmd       *flags.Command

	// subcommands
	groups []*commandGroup

	// completions
	argComps map[string]CompletionFunc
	optComps map[string]CompletionFunc
}

func (c *Command) AddCommand(name, short, long, group, filter string, context string, data interface{}) NewCommand {

	// Check if the group exists within this context, or create
	// it and attach to the specificed context.if needed
	var grp *commandGroup
	for _, g := range c.groups {
		if g.Name == group {
			grp = g
		}
	}
	if grp == nil {
		grp = &commandGroup{Name: group}
		c.groups = append(c.groups, grp)
	}

	// Store the interface data in a command spawing funtion, which acts as an instantiator.
	// We use the command's go-flags struct, as opposed to the console root parser.
	var spawner = func() *flags.Command {
		cmd, err := c.cmd.AddCommand(name, short, long, data)
		if err != nil {
			fmt.Printf("%s Command bind error:%s %s\n", readline.RED, readline.RESET, err.Error())
		}
		if cmd == nil {
			return nil
		}
		return cmd
	}

	// Make a new command struct with everything, and store it in the command tree
	command := &Command{
		Name:      name,
		Short:     short,
		Long:      long,
		Context:   context,
		Group:     group,
		Filters:   []string{filter},
		generator: spawner,
	}
	grp.cmds = append(grp.cmds, command)

	return command.AddCommand
}

func (c *Command) FindCommand(name string) (command *Command) {
	for _, group := range c.groups {
		for _, cmd := range group.cmds {
			if cmd.Name == name {
				return cmd
			}
		}
	}
	return
}
