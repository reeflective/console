package gonsole

import (
	"strings"

	"github.com/jessevdk/go-flags"

	"github.com/maxlandon/readline"
)

// hintCompleter - Entrypoint to all hints in the Wiregost console
func (c *CommandCompleter) hintCompleter(line []rune, pos int) (hint []rune) {

	// Format and sanitize input
	// @args     => All items of the input line
	// @last     => The last word detected in input line as []rune
	// @lastWord => The last word detected in input as string
	args, last, lastWord := formatInput(line)

	// Detect base command automatically
	var command = c.detectedCommand(args)

	// Menu hints (command line is empty, or nothing recognized)
	if noCommandOrEmpty(args, last, command) {
		hint = menuHint(args, last)
	}

	// Check environment variables
	if c.envVarAsked(args, lastWord) {
		return c.envVarHint(args, last)
	}

	// Command Hint
	if commandFound(command) {

		// Command hint by default (no space between cursor and last command character)
		hint = commandHint(command)

		// Check environment variables
		if c.envVarAsked(args, lastWord) {
			return c.envVarHint(args, last)
		}

		// If options are asked for root command, return commpletions.
		if len(command.Groups()) > 0 {
			var groups = command.Groups()
			groups = append(groups, c.console.CommandParser().Groups()...)
			for _, grp := range groups {
				if opt, yes := optionArgRequired(args, last, grp); yes {
					hint = optionArgumentHint(args, last, opt)
				}
			}
		}

		// If command has args, hint for args
		if arg, yes := commandArgumentRequired(lastWord, args, command); yes {
			hint = []rune(commandArgumentHints(args, last, command, arg))
		}

		// Brief subcommand hint
		if lastIsSubCommand(lastWord, command) {
			hint = []rune(cmdHint + command.Find(string(last)).ShortDescription)
		}

		// Handle subcommand if found
		if sub, ok := subCommandFound(lastWord, args, command); ok {
			return c.handleSubcommandHints(args, last, command, sub)
		}

	}

	// Handle system binaries, shell commands, etc...
	if commandFoundInPath(args[0]) {
		// hint = []rune(exeHint + util.ParseSummary(util.GetManPages(args[0])))
	}

	return
}

// commandHint - Yields the hint of a Wiregost command
func commandHint(command *flags.Command) (hint []rune) {
	return []rune(cmdHint + command.ShortDescription)
}

// handleSubcommandHints - Handles hints for a subcommand and its arguments, options, etc.
func (c *CommandCompleter) handleSubcommandHints(args []string, last []rune, rootCommand, command *flags.Command) (hint []rune) {

	// If command has args, hint for args
	if arg, yes := commandArgumentRequired(string(last), args, command); yes {
		hint = []rune(commandArgumentHints(args, last, command, arg))
		return
	}

	// Environment variables
	if c.envVarAsked(args, string(last)) {
		hint = c.envVarHint(args, last)
	}

	// If the last word in input is an option --name, yield argument hint if needed
	if len(command.Groups()) > 0 {
		if len(command.Groups()) > 0 {
			var groups = command.Groups()
			groups = append(groups, c.console.CommandParser().Groups()...)
			groups = append(groups, rootCommand.Groups()...)
			for _, grp := range groups {
				if opt, yes := optionArgRequired(args, last, grp); yes {
					hint = optionArgumentHint(args, last, opt)
				}
			}
		}
	}

	// If user asks for completions with "-" or "--".
	// (Note: This takes precedence on any argument hints, as it is evaluated after them)
	if commandOptionsAsked(args, string(last), command) {
		return optionHints(args, last, command)
	}

	return
}

// commandArgumentHints - Yields hints for arguments to commands if they have some
func commandArgumentHints(args []string, last []rune, command *flags.Command, arg string) (hint []rune) {

	found := argumentByName(command, arg)
	// Base Hint is just a description of the command argument
	hint = []rune(argHint + found.Description)

	return
}

// optionHints - Yields hints for proposed options lists/groups
func optionHints(args []string, last []rune, command *flags.Command) (hint []rune) {
	return
}

// optionArgumentHint - Yields hints for arguments to an option (generally the last word in input)
func optionArgumentHint(args []string, last []rune, opt *flags.Option) (hint []rune) {
	return []rune(valueHint + opt.Description)
}

// menuHint - Returns the Hint for a given menu context
func menuHint(args []string, current []rune) (hint []rune) {
	return
}

// specialCommandHint - Shows hints for Wiregost special commands
func specialCommandHint(args []string, current []rune) (hint []rune) {
	return current
}

// envVarHint - Yields hints for environment variables
func (c *CommandCompleter) envVarHint(args []string, last []rune) (hint []rune) {
	// Trim last in case its a path with multiple vars
	allVars := strings.Split(string(last), "/")
	lastVar := allVars[len(allVars)-1]

	// Base hint
	hint = []rune(envHint + lastVar)

	for exp, comp := range c.console.CurrentContext().expansionComps {
		if strings.HasPrefix(lastVar, string(exp)) {
			envVar := strings.TrimPrefix(lastVar, string(exp))
			grps := comp()
			for _, grp := range grps {
				if value, exists := grp.Descriptions[envVar]; exists {
					hintStr := string(hint) + " => " + value
					hint = []rune(hintStr)
					break
				}
			}
		}
	}

	return
}

var (
	// Hint signs
	menuHintStr = readline.RESET + readline.DIM + readline.BOLD + " menu  " + readline.RESET                                      // Dim
	envHint     = readline.RESET + readline.GREEN + readline.BOLD + " env  " + readline.RESET + readline.DIM + readline.GREEN     // Green
	cmdHint     = readline.RESET + readline.DIM + readline.BOLD + " command  " + readline.RESET + readline.DIM + "\033[38;5;244m" // Cream
	exeHint     = readline.RESET + readline.DIM + readline.BOLD + " shell " + readline.RESET + readline.DIM                       // Dim
	optionHint  = "\033[38;5;222m" + readline.BOLD + " options  " + readline.RESET + readline.DIM + "\033[38;5;222m"              // Cream-Yellow
	valueHint   = readline.RESET + readline.DIM + readline.BOLD + " value  " + readline.RESET + readline.DIM + "\033[38;5;244m"   // Dim
	// valueHint   = "\033[38;5;217m" + readline.BOLD + " Value  " + readline.RESET + readline.DIM + "\033[38;5;244m"         // Pink-Cream
	argHint = readline.DIM + "\033[38;5;217m" + readline.BOLD + " arg  " + readline.RESET + readline.DIM + "\033[38;5;244m" // Pink-Cream
)
