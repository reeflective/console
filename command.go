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

	// Check if the group exists within this context, create if needed
	var grp *commandGroup
	for _, g := range groups {
		if g.Name == group {
			grp = g
		}
	}
	if grp == nil {
		grp = &commandGroup{
			Name:     group,
			commands: map[string][]registerCommand{},
		}
		groups = append(groups, grp)
	}

	// Store the interface data in a command spawing funtion, which acts as an instantiator.
	var spawner = func(name, short, long string, data interface{}) error {
		cmd, err := c.parser.AddCommand(name, short, long, data)
		if err != nil {
			fmt.Printf("%s Command bind error:%s %s\n", readline.RED, readline.RESET, err.Error())
		}
		if cmd == nil {
			return nil
		}
		return nil
	}

	// Add the command to the list of spawners, mapped to a filter.
	// This function will be called at each readline execution loop, for binding the command.
	grp.commands[filter] = append(grp.commands[filter], spawner)
}

// CommandParser - Returns the root command parser of the console.
// Maybe used to find an modify some commands, or add completions to them, etc.
// NOTE: The parser's AddCommand() should not be used to register commands, because
// they will lack a certain quantity of wrapping code.
func (c *Console) CommandParser() (parser *flags.Parser) {
	return c.parser
}

// commandGroup - A group of commands, which might be by any motive: common domain,
// type, etc, as long as the group name is the same. In addition, commands in the same
// group have an additional string filter key, which can be used further refine which
// commands are hidden or not.
// Please see the Console.Hide(filter string) or Console.Show(filter string)
// By default, a "" filter name will mean available no matter the filter.
type commandGroup struct {
	Name     string
	commands map[string][]registerCommand
}

// registerCommand - The command registration functions used to instantiate commands.
type registerCommand func(name, short, long string, data interface{}) error
