package gonsole

import (
	"fmt"

	"github.com/jessevdk/go-flags"
	"github.com/maxlandon/readline"
)

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
// Explanations:
// The function itself accepts a function because the console will need to re-instantiate new,blank command instances
// at each execution loop (so that option/argument values are correctly reset).
// NOTE: The 'data' interface{} parameter needs to be a struct passed by value, not a pointer.
func (c *Console) AddCommand(name, short, long, group, filter string, context string, data interface{}) {

	// Check if the context exists, create it if needed
	var groups []*commandGroup
	ctx, exist := c.contexts[context]
	if exist {
		groups = ctx.commands
	} else {
		c.NewContext(context)
		groups = c.GetContext(context).commands
	}

	// The context needs to keep track now of this command,
	// because maps reorder everything. Lists are used to solve this.
	ctx.groupNames = append(ctx.groupNames, group)

	// Check if the group exists within this context, or create
	// it and attach to the specificed context.if needed
	var grp *commandGroup
	for _, g := range groups {
		if g.Name == group {
			grp = g
		}
	}
	if grp == nil {
		grp = &commandGroup{
			Name:              group,
			commandGenerators: map[string]func() *flags.Command{},
			commandDone:       map[string]*flags.Command{},
		}
		ctx.commands = append(groups, grp)
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

		// The context keeps a reference to this newly generated command.
		ctx.groups[group] = append(ctx.groups[group], cmd)

		return cmd
	}

	// Add the command to the list of spawners, mapped to a filter.
	// This function will be called at each readline execution loop, for binding the command.
	grp.commandGenerators[name] = spawner
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

	for _, group := range c.current.groupsAlt {
		groupNames = append(groupNames, group.Name)

		for _, cmd := range group.commandDone {
			groups[group.Name] = append(groups[group.Name], cmd)
		}
	}
	return
}

// CommandParser - Returns the root command parser of the console.
// Maybe used to find an modify some commands, or add completions to them, etc.
// NOTE: The parser's AddCommand() should not be used to register commands, because
// they will lack a certain quantity of wrapping code.
func (c *Console) CommandParser() (parser *flags.Parser) {
	return c.parser
}

// bindCommands - At every readline loop, we reinstantiate and bind new instances for
// each command. We do not generate those that are filtered with an active filter,
// so that users of the go-flags parser don't have to perform filtering.
func (c *Console) bindCommands() {

	// First, reinstantiate the console command parser
	c.initParser()

	// For each command group in the current context
	for _, groupName := range c.current.groupNames {
		group := c.current.groupsAlt[groupName]

		// erase all references to the currently generated && bound commands.
		group.commandDone = map[string]*flags.Command{}

		// For each command in this group, no matter the filters
		for _, cmdName := range group.commandNames {

			// Find the function that will generate a new instance
			commandGenerate := group.commandGenerators[cmdName]

			// Call the generator function for this command:
			// a new instance will be bound to the parser.
			command := commandGenerate()
			group.commandDone[command.Name] = command

			// If there is an active filter on this command, we mark it hidden.
			cmdFilter, exists := group.commandFilters[cmdName]
			if exists && c.filters[cmdFilter] == true && command != nil {
				command.Hidden = true
			}
		}
	}
}
