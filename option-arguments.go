package gonsole

import (
	"strings"

	"github.com/jessevdk/go-flags"

	"github.com/maxlandon/readline"
)

// completeOptionArguments - Completes all values for arguments to a command. Arguments here are different from command options (--option).
// Many categories, from multiple sources in multiple contexts
func (c *CommandCompleter) completeOptionArguments(gcmd *Command, cmd *flags.Command, opt *flags.Option, lastWord string) (prefix string, completions []*readline.CompletionGroup) {

	// By default the last word is the prefix
	prefix = lastWord

	// First of all: some options, no matter their contexts and subject, have default values.
	// When we have such an option, we don't bother analyzing context, we just build completions and return.
	if len(opt.Choices) > 0 {
		var comp = &readline.CompletionGroup{
			Name:        opt.ValueName, // Value names are specified in struct metadata fields
			DisplayType: readline.TabDisplayGrid,
		}
		for _, choice := range opt.Choices {
			if strings.HasPrefix(choice, lastWord) {
				comp.Suggestions = append(comp.Suggestions, choice)
			}
		}
		completions = append(completions, comp)
		return
	}

	// Check if the option name has a user-defined completion generator
	for optName, completer := range gcmd.optComps {
		if strings.Contains(opt.Field().Name, optName) {

			// Call this generator and add to the completions
			pref, comps := completer(lastWord)
			prefix = pref
			completions = append(completions, comps...)
		}
	}

	return
}
