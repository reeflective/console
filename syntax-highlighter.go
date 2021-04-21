package gonsole

import (
	"strings"

	"github.com/jessevdk/go-flags"

	"github.com/maxlandon/readline"
)

// syntaxHighlighter - Entrypoint to all input syntax highlighting in the Wiregost console
func (c *CommandCompleter) syntaxHighlighter(input []rune) (line string) {

	// Format and sanitize input
	args, last, lastWord := formatInputHighlighter(input)

	// Remain is all arguments that have not been highlighted, we need it for completing long commands
	var remain = args

	// Detect base command automatically
	var command = c.detectedCommand(args)

	// Return input as is
	if noCommandOrEmpty(remain, last, command) {
		return string(input)
	}

	// Base command
	if commandFound(command) {

		// Get the corresponding *Command from the console
		gCommand := c.console.FindCommand(command.Name)
		if gCommand == nil {
			return
		}

		// Highlight the word, and return the shorter list of arguments to process.
		line, remain = c.highlightCommand(line, remain, command)

		// SubCommand
		if sub, ok := subCommandFound(lastWord, args, command); ok {
			subgCommand := gCommand.FindCommand(sub.Name)
			if gCommand != nil {
				line, remain = c.handleSubCommandSyntax(line, remain, command, sub, subgCommand)
			}
		}
	}

	// Process any expanded variables found, between others
	line = c.processRemain(line, remain)

	return
}

func (c *CommandCompleter) handleSubCommandSyntax(processed string, args []string, parent, command *flags.Command, gCommand *Command) (line string, remain []string) {

	line, remain = c.highlightCommand(processed, args, command)

	// SubCommand
	if sub, ok := subCommandFound(c.lastWord, args, command); ok {
		subgCommand := gCommand.FindCommand(sub.Name)
		if gCommand != nil {
			line, remain = c.handleSubCommandSyntax(line, remain, command, sub, subgCommand)
		}
	}
	return
}

func (c *CommandCompleter) highlightCommand(processed string, args []string, command *flags.Command) (line string, remain []string) {
	var color = c.getTokenHighlighting("{command}")
	line += color + args[0] + readline.RESET + " "
	remain = args[1:]
	return processed + line, remain
}

func (c *CommandCompleter) highlightSubCommand(input string, args []string, command *flags.Command) (line string, remain []string) {
	line = input
	var color = c.getTokenHighlighting("{command}")
	line += color + args[0] + readline.RESET + " "
	remain = args[1:]
	return
}

func (c *CommandCompleter) processRemain(input string, remain []string) (line string) {

	// Check the last is not the last space in input
	if len(remain) == 1 && remain[0] == " " {
		return input
	}

	// line = input + strings.Join(remain, " ")
	line = c.processEnvVars(input, remain)
	return
}

// evaluateExpansion - Given a single "word" argument, resolve any embedded expansion variables
func (c *CommandCompleter) evaluateExpansion(arg string) (expanded string) {
	// For each available per-menu expansion variable, evaluate and replace. Any group
	// successfully replacing the token will break the loop, and the remaining expanders will
	// not be evaluated.
	var evaluated = false
	for exp := range c.console.current.expansionComps {
		var color = c.getTokenHighlighting(string(exp))

		if strings.HasPrefix(arg, string(exp)) { // It is an env var.
			if args := strings.Split(arg, "/"); len(args) > 1 {
				var processed = []string{}
				for _, a := range args {
					processed = append(processed, c.evaluateExpansion(a))
					// if strings.HasPrefix(a, string(exp)) && a != " " { // It is an env var.
					//         processed = append(processed, color+a+readline.RESET)
					//         evaluated = true
					//         break
					// }
				}
				expanded = strings.Join(processed, "/")
				evaluated = true
				break
			}
			expanded = color + arg + readline.RESET
			evaluated = true
			break
		}
	}
	if !evaluated {
		expanded = arg
	}
	return
}

// processEnvVars - Highlights environment variables. NOTE: Rewrite with logic from console/env.go
func (c *CommandCompleter) processEnvVars(input string, remain []string) (line string) {

	var processed []string

	inputSlice := strings.Split(input, " ")

	// Check already processed input
	for _, arg := range inputSlice {
		if arg == "" || arg == " " {
			continue
		}
		processed = append(processed, c.evaluateExpansion(arg))
	}

	// Check remaining args (non-processed)
	for _, arg := range remain {
		if arg == "" {
			continue
		}

		processed = append(processed, c.evaluateExpansion(arg))
	}

	line = strings.Join(processed, " ")

	// Very important, keeps the line clear when erasing
	// line += " "

	return
}

func (c *CommandCompleter) getTokenHighlighting(token string) (highlight string) {
	// Get the effect from the config and load it
	if effect, found := c.console.config.Highlighting[token]; found {
		highlight = effect

		if defColor, exists := defaultColorCallbacks[effect]; exists {
			highlight = defColor
		}
	}

	return
}
