package gonsole

import (
	"github.com/jessevdk/go-flags"
	"github.com/maxlandon/readline"
)

// Context - A context is a simple way to seggregate commands based on
// the environment to which they belong. For instance, when using a context
// specific to some host/user, or domain of activity, commands will vary.
type Context struct {
	Name   string  // This name is just used for retrieving usage
	Prompt *Prompt // A dedicated prompt with its own callbacks and colors

	// UnknownCommandHandler - The user can specify a function that will
	// be executed if the error raised by the application parser is a
	// ErrUnknownCommand error. This might be used for executing the
	// input line directly via a system shell, or any os.Exec mean...
	UnknownCommandHandler func(args []string) error

	// Each context has its own command parser, which executes dispatched commands
	parser *flags.Parser

	// Command - The context embeds a command so that users
	// can more explicitly register commands to a given context.
	cmd *Command

	// Each context can have two specific history sources
	historyCtrlRName string
	historyCtrlR     readline.History
	historyAltRName  string
	historyAltR      readline.History

	// expansionComps - A list of completion generators that are triggered when
	// the given string is detected (anywhere, even in other completions) in the input line.
	expansionComps map[rune]CompletionFunc
}

func newContext() *Context {
	ctx := &Context{
		Prompt: &Prompt{
			Callbacks: map[string]func() string{},
			Colors:    defaultColorCallbacks,
		},
		cmd:            newCommand(),
		expansionComps: map[rune]CompletionFunc{},
	}
	return ctx
}

// initParser - Called each time the readline loops, before rebinding all command instances.
func (c *Context) initParser(opts flags.Options) {
	c.parser = flags.NewNamedParser(c.Name, opts)
}

// Commands - Returns the list of child gonsole.Commands for this command. You can set
// anything to them, these changes will persist for the lifetime of the application,
// or until you deregister this command or one of its childs.
func (c *Context) Commands() (cmds []*Command) {
	return c.cmd.Commands()
}

// CommandGroups - Returns the command's child commands, structured in their respective groups.
// Commands having been assigned no specific group are the group named "".
func (c *Context) CommandGroups() (grps []*commandGroup) {
	return c.cmd.groups
}

// OptionGroups - Returns all groups of options that are bound to this command. These
// groups (and their options) are available for use even in the command's child commands.
func (c *Context) OptionGroups() (grps []*optionGroup) {
	return c.cmd.opts
}

// AddGlobalOptions - Add global options for this context command parser. Will appear in all commands.
func (c *Context) AddGlobalOptions(shortDescription, longDescription string, data func() interface{}) {
	c.cmd.AddGlobalOptions(shortDescription, longDescription, data)
}

// AddCommand - Add a command to this context. This command will be available when this context is active.
func (c *Context) AddCommand(name, short, long, group string, filters []string, data func() interface{}) *Command {
	return c.cmd.AddCommand(name, short, long, group, filters, data)
}

// SetHistoryCtrlR - Set the history source triggered with Ctrl-R
func (c *Context) SetHistoryCtrlR(name string, hist readline.History) {
	c.historyCtrlRName = name
	c.historyCtrlR = hist
}

// SetHistoryAltR - Set the history source triggered with Alt-r
func (c *Context) SetHistoryAltR(name string, hist readline.History) {
	c.historyAltRName = name
	c.historyAltR = hist
}

// NewContext - Create a new command context, to which the user
// can attach some specific items, like history sources.
func (c *Console) NewContext(name string) (ctx *Context) {
	ctx = newContext()
	ctx.Name = name
	c.contexts[name] = ctx
	return
}

// GetContext - Given a name, return the appropriate context.
// If the context does not exists, it returns nil
func (c *Console) GetContext(name string) (ctx *Context) {
	if context, exists := c.contexts[name]; exists {
		return context
	}
	return
}

// SwitchContext - Given a name, the console switches its command context:
// The next time the console rebinds all of its commands, it will only bind those
// that belong to this new context. If the context is invalid, i.e that no commands
// are bound to this context name, the current context is kept.
func (c *Console) SwitchContext(context string) {
	for _, ctx := range c.contexts {
		if ctx.Name == context {
			c.current = ctx
		}
	}
}

// CurrentContext - Return the current console context. Because the Context
// is just a reference, any modifications to this context will persist.
func (c *Console) CurrentContext() *Context {
	return c.current
}
