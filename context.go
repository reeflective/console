package gonsole

import (
	"github.com/jessevdk/go-flags"
	"github.com/maxlandon/readline"
)

// Context - A context is a simple way to seggregate commands based on
// the environment to which they belong. For instance, when using a context
// specific to some host/user, or domain of activity, commands will vary.
type Context struct {
	Name string // This name is just used for retrieving usage

	// Prompt - A dedicated prompt with its own callbacks and colors
	Prompt *Prompt

	// commands - All command groups available in this context.
	// Because we need to reinstantiate blank commands at each loop,
	// the user registers yielder functions that are mapped to a given context.
	commands []*commandGroup

	// All generated commands and structured in equivalent groups.
	groups     map[string][]*flags.Command
	groupNames []string

	// Each context can have two specific history sources
	historyCtrlRName string
	historyCtrlR     readline.History
	historyCtrlEName string
	historyCtrlE     readline.History
}

// NewContext - Create a new command context, to which the user
// can attach some specific items, like history sources.
func (c *Console) NewContext(name string) (ctx *Context) {
	ctx = &Context{
		Name: name,
		Prompt: &Prompt{
			Callbacks: map[string]func() string{},
			Colors:    defaultColorCallbacks,
		},
		commands: make([]*commandGroup, 0),
		groups:   map[string][]*flags.Command{},
	}
	c.contexts[name] = ctx
	return
}

// GetContext - Given a name, return the appropriate context. Returns nil if invalid.
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

}

// SetHistoryCtrlR - Set the history source triggered with Ctrl-R
func (c *Context) SetHistoryCtrlR(name string, hist readline.History) {
	c.historyCtrlRName = name
	c.historyCtrlR = hist
}

// SetHistoryCtrlE - Set the history source triggered with Ctrl-E
func (c *Context) SetHistoryCtrlE(name string, hist readline.History) {
	c.historyCtrlEName = name
	c.historyCtrlE = hist
}
