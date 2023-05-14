package console

import (
	"strings"

	"github.com/spf13/cobra"
)

var (
	seqFgGreen  = "\x1b[32m"
	seqFgYellow = "\x1b[33m"
	seqFgReset  = "\x1b[39m"
)

// highlightSyntax - Entrypoint to all input syntax highlighting in the Wiregost console.
func (c *Console) highlightSyntax(input []rune) (line string) {
	// Split the line as shellwords
	args, unprocessed, err := split(string(input), true)
	if err != nil {
		args = append(args, unprocessed)
	}

	highlighted := make([]string, 0)   // List of processed words, append to
	remain := args                     // List of words to process, draw from
	trimmed := trimSpacesMatch(remain) // Match stuff against trimmed words

	// Highlight the root command when found.
	cmd, _, _ := c.activeMenu().Find(trimmed)
	if cmd != nil {
		highlighted, remain = c.highlightCommand(highlighted, args, cmd)
	}

	// Done with everything, add remainind, non-processed words
	highlighted = append(highlighted, remain...)

	// Join all words.
	line = strings.Join(highlighted, "")

	return line
}

func (c *Console) highlightCommand(done, args []string, cmd *cobra.Command) ([]string, []string) {
	highlighted := make([]string, 0)
	rest := make([]string, 0)

	if len(args) == 0 {
		return done, args
	}

	// The first word is the command, highlight it.
	if rootcmd, _, _ := c.activeMenu().Find(args[1:]); rootcmd != nil {
		highlighted = append(highlighted, seqFgGreen+args[0]+seqFgReset)
		rest = args[1:]
	}

	return append(done, highlighted...), rest
}
