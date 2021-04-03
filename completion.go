package gonsole

import (
	"github.com/jessevdk/go-flags"
	"github.com/maxlandon/readline"
)

// CompletionFunc - A function that yields one or more completion groups.
// The prefix parameter should be used with a simple 'if strings.IsPrefix()' condition.
// Please see the project wiki for documentation on how to write more elaborated engines:
// For example, do NOT use or modify the `pref string` return paramater if you don't explicitely need to.
type CompletionFunc func(prefix string) (pref string, comps []*readline.CompletionGroup)

// AddArgumentCompletion - Given a registered command, add one or more groups of completion items
// (with any display style/options) to one of the command's arguments.
// It is VERY IMPORTANT to pass the case-sensitive name of the argument, as declared in the command struct.
// The type of the underlying argument does not matter, and gonsole will correctly yield suggestions based
// on wheteher list are required, are these arguments optional, etc.
// The context is needed in order to bind these completions to the good command,
// because several contexts migh have some being identically named.
func (c *Console) AddArgumentCompletion(cmd *flags.Command, context, arg string, comps CompletionFunc) {
	if cmd == nil {
		return
	}
	return
}

// AddOptionCompletion - Given a registered command and an option LONG name, add one or
// more groups of completion items to this option's arguments.
// It is VERY IMPORTANT to pass the case-sensitive name of the option, as declared in the command struct.
// The type of the underlying argument does not matter, and gonsole will correctly yield suggestions based
// on wheteher list are required, are these arguments optional, etc.
// The context is needed in order to bind these completions to the good command,
// because several contexts migh have some being identically named.
func (c *Console) AddOptionCompletion(cmd *flags.Command, context, option string, comps CompletionFunc) {
	if cmd == nil {
		return
	}

	return
}
