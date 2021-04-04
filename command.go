package gonsole

import (
	"fmt"

	"github.com/jessevdk/go-flags"
	"github.com/maxlandon/readline"
)

// Command - A struct storing basic command info, functions used for command
// instantiation, completion generation, and any number of subcommand groups.
type Command struct {
	Name             string
	ShortDescription string
	LongDescription  string
	Group            string
	Filters          []string

	Data      func() interface{}
	generator func(cParser commandParser) *flags.Command
	cmd       *flags.Command

	// global opts generator
	opts []*optionGroup

	// subcommands
	groups []*commandGroup

	// completions
	argComps map[string]CompletionFunc
	optComps map[string]CompletionFunc
}

func newCommand() *Command {
	c := &Command{
		argComps: map[string]CompletionFunc{},
		optComps: map[string]CompletionFunc{},
	}
	return c
}

// AddCommand - Add a command to the given command (the console Contexts embed a command for this matter). If you are
// calling this function directly like gonsole.Console.AddCommand(), be aware that this will bind the command to the
// default context named "". If you don't intend to use multiple contexts this is fine, but if you do, you should
// create and name each of your contexts, and add commands to them, like Console.NewContext("name").AddCommand("", "", ...)
func (c *Command) AddCommand(name, short, long, group, filter string, data func() interface{}) *Command {

	if data == nil {
		return nil
	}

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
	var spawner = func(cmdParser commandParser) *flags.Command {
		cmd, err := cmdParser.AddCommand(name, short, long, data())
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
		Name:             name,
		ShortDescription: short,
		LongDescription:  long,
		Group:            group,
		Filters:          []string{filter},
		generator:        spawner,
	}
	grp.cmds = append(grp.cmds, command)

	return command
}

// Add - Same as AddCommand("", "", ...), but passing a populated Command struct.
func (c *Command) Add(cmd *Command) *Command {
	return c.AddCommand(cmd.Name, cmd.ShortDescription, cmd.LongDescription, cmd.Group, cmd.Filters[0], cmd.Data)
}

// AddCommandT - Add a command to the default console context, named "". Please check gonsole.CurrentContext().AddCommand(),
// if you intend to use multiple contexts, for more detailed explanations
func (c *Console) AddCommandT(name, short, long, group, filter string, context string, data func() interface{}) *Command {
	return c.current.cmd.AddCommand(name, short, long, group, filter, data)
}

// Add - Same as AddCommand("", "", ...), but passing a populated Command struct.
func (c *Console) Add(cmd *Command) *Command {
	return c.current.AddCommand(cmd.Name, cmd.ShortDescription, cmd.LongDescription, cmd.Group, cmd.Filters[0], cmd.Data)
}

// HideCommands - Commands, in addition to their contexts, can be shown/hidden based
// on a filter string. For example, some commands applying to a Windows host might
// be scattered around different groups, but, having all the filter "windows".
// If "windows" is used as the argument here, all windows commands for the current
// context are subsquently hidden, until ShowCommands("windows") is called.
func (c *Console) HideCommands(filter string) {
}

// ShowCommands - Commands, in addition to their contexts, can be shown/hidden based
// on a filter string. For example, some commands applying to a Windows host might
// be scattered around different groups, but, having all the filter "windows".
// Use this function if you have previously called HideCommands("filter") and want
// these commands to be available back under their respective context.
func (c *Console) ShowCommands(filter string) {

}

// FindCommand - Find a subcommand of this command, given the command name.
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

// GetCommands - Callers of this are for example the TabCompleter, which needs to call
// this regularly in order to have a list of commands belonging to the current context.
func (c *Console) GetCommands() (groups map[string][]*flags.Command, groupNames []string) {

	groups = map[string][]*flags.Command{}

	for _, group := range c.current.cmd.groups {
		groupNames = append(groupNames, group.Name)

		for _, cmd := range group.cmds {
			groups[group.Name] = append(groups[group.Name], cmd.cmd)
		}
	}
	return
}

// FindCommand - Find a command among the root ones in the application, for the current context.
func (c *Console) FindCommand(name string) (command *Command) {
	for _, group := range c.current.cmd.groups {
		for _, cmd := range group.cmds {
			if cmd.Name == name {
				return cmd
			}
		}

	}
	return
}

// bindCommands - At every readline loop, we reinstantiate and bind new instances for
// each command. We do not generate those that are filtered with an active filter,
// so that users of the go-flags parser don't have to perform filtering.
func (c *Console) bindCommands() {
	cc := c.current

	// First, reset the parser for the current context.
	cc.initParser(c.parserOpts)

	// Generate all global options if there are some.
	for _, opt := range cc.cmd.opts {
		cc.parser.AddGroup(opt.short, opt.long, opt.generator())
	}

	// For each (root) command group in this context.
	for _, group := range cc.cmd.groups {

		// For each command in the group, yield a flags.Command
		for _, cmd := range group.cmds {

			// The generator function has been contextually adapted, to either
			// bind to the root parser, or to the parent command of this one.
			cmd.cmd = cmd.generator(cc.parser)

			// Bind any subcommands of this cmd
			for _, subgroup := range cmd.groups {
				c.bindCommandGroup(cmd, subgroup)
			}

			// If there is an active filter on this command, we mark it hidden.
			for _, filter := range c.filters {
				for _, filt := range cmd.Filters {
					if filt == filter && cmd.cmd != nil {
						cmd.cmd.Hidden = true
					}
				}
			}
		}
	}
}
