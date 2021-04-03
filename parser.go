package gonsole

import (
	"fmt"

	"github.com/jessevdk/go-flags"
	"github.com/maxlandon/readline"
)

func (c *Console) initParser() {
	c.parser = flags.NewNamedParser("", c.parserOpts)
}

var (
	commandError = fmt.Sprintf("%s[Command Error]%s ", readline.RED, readline.RESET)
	parserError  = fmt.Sprintf("%s[Parser Error]%s ", readline.RED, readline.RESET)
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

// CommandParser - Returns the root command parser of the console.
// Maybe used to find an modify some commands, or add completions to them, etc.
// NOTE: The parser's AddCommand() should not be used to register commands, because
// they will lack a certain quantity of wrapping code.
func (c *Console) CommandParser() (parser *flags.Parser) {
	return c.parser
}

// Find - Given the name of the command, return its go-flags object.
// Can be used for many things: please see the go-flags documentation.
func (c *Console) Find(command string) (cmd *flags.Command) {
	return c.parser.Find(command)
}
