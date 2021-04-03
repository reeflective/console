package gonsole

import (
	"fmt"

	"github.com/jessevdk/go-flags"
	"github.com/maxlandon/readline"
)

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

func (c *Command) AddCommand(name, short, long, group, filter string, context string, data interface{}) *Command {

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

	return command
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

// AddCommand - Add a command to gonsole. This command is registered within the go-flags parser, and  return so that you
// can further refine its settings, or pass it to other gonsole functions (ex: for registering argument/option completions)
// Parameters:
// @name    - Then name of the command as typed in the input line
// @short   - A short description, used in completions and hints
// @long    - A long description, appended to the -h, --help message.
// @group   - Name of a group of command. It gives commands a structure in completions and help.
// @filter  - An optional, other filter under which a user may hide/show this command. See Console.ShowFilter()
// @context - The command is available in the given context. Available in all if empty. See gonsole.Context
// @data    - The actual command struct PASSED BY VALUE. See wiki/docs on how to declare commands via struct fields.
//
// Return values:
// @NewCommand - A function identical to this AddCommand, for registering subcommands to this command.
// NOTE: The 'data' interface{} parameter needs to be a struct passed by value, not a pointer.
func (c *Console) AddCommand(name, short, long, group, filter string, context string, data interface{}) *Command {

	// Check if the context exists, create it if needed
	var groups []*commandGroup
	ctx, exist := c.contexts[context]
	if exist {
		groups = ctx.groupsAltT
	} else {
		c.NewContext(context)
		groups = c.GetContext(context).groupsAltT
	}

	// Check if the group exists within this context, or create
	// it and attach to the specificed context.if needed
	var grp *commandGroup
	for _, g := range groups {
		if g.Name == group {
			grp = g
		}
	}
	if grp == nil {
		grp = &commandGroup{Name: group}
		ctx.groupsAltT = append(groups, grp)
	}

	// Store the interface data in a command spawing funtion, which acts as an instantiator.
	var spawner = func() *flags.Command {
		cmd, err := c.parser.AddCommand(name, short, long, data)
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

	// Return the function allowing to register subcommands to this command
	return command
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

// GetCommands - Callers of this are for example the TabCompleter, which needs to call
// this regularly in order to have a list of commands belonging to the current context.
func (c *Console) GetCommands() (groups map[string][]*flags.Command, groupNames []string) {

	groups = map[string][]*flags.Command{}

	for _, group := range c.current.groupsAltT {
		groupNames = append(groupNames, group.Name)

		for _, cmd := range group.cmds {
			groups[group.Name] = append(groups[group.Name], cmd.cmd)
		}
	}
	return
}

func (c *Console) FindCommand(name string) (command *Command) {
	for _, group := range c.current.groupsAltT {
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
func (c *Console) bindCommandsAlt() {

	// First, reinstantiate the console command parser
	c.initParser()

	for _, group := range c.current.groupsAltT {

		// For each command in the group, yield a flags.Command
		for _, cmd := range group.cmds {

			// The generator function has been contextually adapted, to either
			// bind to the root parser, or to the parent command of this one.
			cmd.cmd = cmd.generator()

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
