package console

import (
	"errors"
	"strings"

	"github.com/reeflective/readline"
	"github.com/rsteube/carapace"
)

func (c *Console) complete(line []rune, pos int) readline.Completions {
	menu := c.menus.current()

	// Split the line as shell words, only using
	// what the right buffer (up to the cursor)
	rbuffer := line[:pos]
	args, prefix := splitArgs(rbuffer)

	// Apply some sanitizing to the last argument.
	args = sanitizeArgs(args)

	// Like in classic system shells, we need to add an empty
	// argument if the last character is a space: the args
	// returned from the previous call don't account for it.
	if strings.HasSuffix(string(rbuffer), " ") || len(args) == 0 {
		args = append(args, "")
	}

	// Prepare arguments for the carapace completer
	args = append([]string{"examples", "_carapace"}, args...)

	// Call the completer with our current command context.
	values, meta := carapace.Complete(menu.Command, args, menu.resetCommands)

	// Tranfer all completion results to our readline shell completions.
	raw := make([]readline.Completion, len(values))

	for idx, val := range values {
		value := readline.Completion{
			Value:       val.Value,
			Display:     val.Display,
			Description: val.Description,
			Style:       val.Style,
			Tag:         val.Tag,
		}
		raw[idx] = value
	}

	// Assign both completions and command/flags/args usage strings.
	comps := readline.CompleteRaw(raw)
	comps = comps.Usage(meta.Usage)

	// Suffix matchers for the completions if any.
	if meta.Nospace.String() != "" {
		comps = comps.NoSpace([]rune(meta.Nospace.String())...)
	}

	// If we have a quote/escape sequence unaccounted
	// for in our completions, add it to all of them.
	if prefix != "" {
		comps = comps.Prefix(prefix)
	}

	return comps
}

func sanitizeArgs(args []string) (sanitized []string) {
	if len(args) == 0 {
		return
	}

	sanitized = args[:len(args)-1]
	last := args[len(args)-1]

	// The last word should not comprise newlines.
	last = strings.ReplaceAll(last, "\n", "")
	sanitized = append(sanitized, last)

	return sanitized
}

func splitArgs(line []rune) (args []string, prefix string) {
	// Split the line as shellwords, return them if all went fine.
	args, remain, err := split(string(line), false)
	if err == nil {
		return
	}

	// If we had an error, it's because we have an unterminated quote/escape sequence.
	// In this case we split the remainder again, as the completer only ever considers
	// words as space-separated chains of characters.
	if errors.Is(err, unterminatedDoubleQuoteError) {
		remain = strings.Trim(remain, "\"")
		prefix = "\""
	} else if errors.Is(err, unterminatedSingleQuoteError) {
		remain = strings.Trim(remain, "'")
		prefix = "'"
	}

	args = append(args, strings.Split(remain, " ")...)

	return
}
