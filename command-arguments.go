package gonsole

import (
	"strings"

	"github.com/jessevdk/go-flags"

	"github.com/maxlandon/readline"
)

// CompleteCommandArguments - Completes all values for arguments to a command.
// Arguments here are different from command options (--option).
// Many categories, from multiple sources in multiple contexts
func (c *CommandCompleter) completeCommandArguments(gcmd *Command, cmd *flags.Command, arg string, lastWord string) (prefix string, completions []*readline.CompletionGroup) {

	// the prefix is the last word, by default
	prefix = lastWord
	found := argumentByName(cmd, arg)

	// Check if the argument name has a user-defined completion generator
	for argName, completer := range gcmd.argComps {
		if strings.Contains(found.Name, argName) {

			// Call this generator and add to the completions
			pref, comps := completer(lastWord)
			prefix = pref
			completions = append(completions, comps...)
		}
	}

	return
}
