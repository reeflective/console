package gonsole

import (
	"fmt"

	"github.com/jessevdk/go-flags"
	"github.com/maxlandon/readline"
)

// SetParserOptions - Set the general options that apply to the root command parser.
// Default options are:
// -h, --h options are available to all registered commands.
// Ignored option dashes are ignored and passed along the command tree.
func (c *Console) SetParserOptions(options flags.Options) {
	c.parserOpts = options
	if c.parser != nil {
		c.parser.Options = options
	}
	return
}

// Find - Given the name of the command, return its go-flags object.
// Can be used for many things: please see the go-flags documentation.
func (c *Console) Find(command string) (cmd *flags.Command) {
	return c.parser.Find(command)
}

func (c *Console) initParser() {
	c.parser = flags.NewNamedParser("", c.parserOpts)
}

var (
	commandError = fmt.Sprintf("%s[Command Error]%s ", readline.RED, readline.RESET) // CommandError - Command input error
	parserError  = fmt.Sprintf("%s[Parser Error]%s ", readline.RED, readline.RESET)  // ParserError - Failed to parse some tokens in the input
)
