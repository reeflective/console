package gonsole

import (
	"fmt"

	"github.com/jessevdk/go-flags"
	"github.com/maxlandon/readline"
)

// AddHelpCommand - The console will automatically add a command named "help", which accepts any
// (optional) command and/or any of its subcommands, and prints the corresponding help. If no
// argument is passed, prints the list of available of commands for the current context.
// The name of the group is left to the user's discretion, for putting the command in a given group/topic.
// Command names and their subcommands will be automatically completed.
func (c *Console) AddHelpCommand(group string) {
	for _, cc := range c.contexts {
		help := cc.AddCommand("help",
			"print menu, command or subcommand help for the current context (menu)",
			"",
			group,
			[]string{""},
			func() interface{} { return &Help{console: c} })
		help.AddArgumentCompletion("Command", c.Completer.contextCommands)
	}
}

// Help - Print help for the current context (lists all commands)
type Help struct {
	Positional struct {
		Command    string `description:"(optional) command to print help for"`
		SubCommand string `description:"(optional) subcommand of the root commmand passed as argument"`
	} `positional-args:"true"`

	// Needed to access commands
	console *Console
}

// Execute - Print help for the current context (lists all commands)
func (h *Help) Execute(args []string) (err error) {

	parser := h.console.CommandParser()

	// If no component argument is asked for
	if h.Positional.Command == "" {
		h.console.printMenuHelp(h.console.CurrentContext().Name)
		return
	}

	var command *flags.Command
	for _, cmd := range parser.Commands() {
		if cmd.Name == h.Positional.Command {
			command = cmd
		}
	}
	if command == nil {
		fmt.Printf(errorStr+"Invalid command: %s%s%s\n",
			readline.BOLD, h.Positional.Command, readline.RESET)
		return
	}

	if h.Positional.SubCommand == "" {
		h.console.printCommandHelp(command)
		return
	}

	var sub *flags.Command
	for _, cmd := range command.Commands() {
		if cmd.Name == h.Positional.SubCommand {
			sub = cmd
		}
	}
	if sub == nil {
		fmt.Printf(errorStr+"Invalid command: %s%s%s\n",
			readline.BOLD, h.Positional.SubCommand, readline.RESET)
		return
	}

	h.console.printCommandHelp(sub)

	return
}

func (c *CommandCompleter) contextCommands() (completions []*readline.CompletionGroup) {

	for _, cmd := range c.console.CommandParser().Commands() {
		// Check command group: add to existing group if found
		var found bool
		for _, grp := range completions {
			if grp.Name == c.console.GetCommandGroup(cmd) {
				found = true
				grp.Suggestions = append(grp.Suggestions, cmd.Name)
				grp.Descriptions[cmd.Name] = readline.Dim(cmd.ShortDescription)
			}
		}
		// Add a new group if not found
		if !found {
			grp := &readline.CompletionGroup{
				Name:        c.console.GetCommandGroup(cmd),
				Suggestions: []string{cmd.Name},
				Descriptions: map[string]string{
					cmd.Name: readline.Dim(cmd.ShortDescription),
				},
			}
			completions = append(completions, grp)
		}
	}
	// Make adjustments to the CompletionGroup list: set maxlength depending on items, check descriptions, etc.
	for _, grp := range completions {
		// If the length of suggestions is too long and we have
		// many groups, use grid display.
		if len(completions) >= 10 {
			// if len(completions) >= 10 && len(grp.Suggestions) >= 10 {
			grp.DisplayType = readline.TabDisplayGrid
		} else {
			// By default, we use a map of command to descriptions
			grp.DisplayType = readline.TabDisplayList
		}
	}
	return
}

func (c *CommandCompleter) subCommands() (completions []*readline.CompletionGroup) {

	// First argument is the 'help' command, second is 'command'
	return
}