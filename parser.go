package gonsole

import (
	"fmt"

	"github.com/jessevdk/go-flags"
	"github.com/maxlandon/readline"
)

var (
	commandError = fmt.Sprintf("%s[Command Error]%s ", readline.RED, readline.RESET)
	parserError  = fmt.Sprintf("%s[Parser Error]%s ", readline.RED, readline.RESET)

	info     = fmt.Sprintf("%s[-]%s ", readline.BLUE, readline.RESET)   // Info - All normal messages
	warn     = fmt.Sprintf("%s[!]%s ", readline.YELLOW, readline.RESET) // Warn - Errors in parameters, notifiable events in modules/sessions
	errorStr = fmt.Sprintf("%s[!]%s ", readline.RED, readline.RESET)    // Error - Error in commands, filters, modules and implants.
	Success  = fmt.Sprintf("%s[*]%s ", readline.GREEN, readline.RESET)  // Success - Success events

	infof   = fmt.Sprintf("%s[-] ", readline.BLUE)   // Infof - formatted
	warnf   = fmt.Sprintf("%s[!] ", readline.YELLOW) // Warnf - formatted
	errorf  = fmt.Sprintf("%s[!] ", readline.RED)    // Errorf - formatted
	sucessf = fmt.Sprintf("%s[*] ", readline.GREEN)  // Sucessf - formatted
)

// commandParser - Both flags.Command and flags.Parser can add commands in the same way,
// we need to be able to call the appropriate target no matter the level of command nesting.
type commandParser interface {
	AddCommand(name, short, long string, data interface{}) (cmd *flags.Command, err error)
}

// SetParserOptions - Set the general options that apply to the root command parser.
// Default options are:
// -h, --h options are available to all registered commands.
// Ignored option dashes are ignored and passed along the command tree.
func (c *Console) SetParserOptions(options flags.Options) {
	c.parserOpts = options
	if c.current.parser != nil {
		c.current.parser.Options = options
	}
	return
}

// CommandParser - Returns the root command parser of the console.
// Maybe used to find an modify some commands, or add completions to them, etc.
// NOTE: The parser's AddCommand() should not be used to register commands, because
// they will lack a certain quantity of wrapping code.
func (c *Console) CommandParser() (parser *flags.Parser) {
	return c.current.parser
}

// Find - Given the name of the command, return its go-flags object.
// Can be used for many things: please see the go-flags documentation.
// This will only scan the commands for the current context.
func (c *Console) Find(command string) (cmd *flags.Command) {
	return c.current.parser.Find(command)
}

// optionGroup - Used to generate global option structs, bound to commands/parsers.
type optionGroup struct {
	short     string
	long      string
	generator func() interface{}
}

// AddGlobalOptions - Global options are available in all child commands of this command
// (or all commands of the parser). The data interface is a struct declared the same way
// as you'd declare a go-flags parsable option struct.
func (c *Command) AddGlobalOptions(shortDescription, longDescription string, data func() interface{}) {
	optGroup := &optionGroup{
		short:     shortDescription,
		long:      longDescription,
		generator: data,
	}
	c.opts = append(c.opts, optGroup)
}
