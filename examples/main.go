package main

import (
	"github.com/reeflective/console"
	"github.com/reeflective/flags"
	"github.com/reeflective/flags/example/commands"
	"github.com/reeflective/flags/gen/completions"
	genflags "github.com/reeflective/flags/gen/flags"
	"github.com/reeflective/flags/validator"
	"github.com/rsteube/carapace"
	"github.com/spf13/cobra"
)

const (
	shortUsage = "Console application example, with commands/flags/completions generated from structs"
)

func main() {
	// Instantiate a new console, with a single, default menu.
	// All defaults are set, and nothing is needed to make it work
	console := console.New()

	// Assuming that the user has a system-wide readline.yml configuration
	// (containing keybinds, completion behavior, etc), load it. If the file
	// does not exist, the readline shell will use default but sane settings.
	//
	// This also shows how library consumers can access the
	// underlying readline shell for lower-level configuration.
	console.Shell().Config().LoadSystem()

	// By default the shell as created a single menu and
	// made it current, so you can access it and set it up.
	menu := console.CurrentMenu()

	// Now that the application is set up, we need a tree of commands,
	// and optionally completions for this tree. Check the comments in
	// the function below for details.
	rootCmd, completer := getCommands()

	// The root command above has os.Args[0] (here 'example') as its name
	// by default (due to reeflective/flags behavior). We can bind this
	// command to our current menu in two different ways:
	//
	// - Replacing the menu root altogether: the 'example' command name will
	//   not need to be entered every single time one of its subcommands is
	//   called: thus, in this case, the root command is just a root parser.
	//
	// - Adding this command to the current menu's root command, with
	//   menu.AddCommand(rootCmd). In this case, the 'example' name will
	//   need to be entered each time before one of the subcommands.
	//
	menu.Command = rootCmd

	// The completer is provided by github.com/rsteube/carapace.
	// This library provides an extensive and efficient completion
	// engine for cobra commands, along with hundreds of different
	// completers for various system stuff.
	// The console wraps it up into a function that is passed to the
	// underlying readline shell, to produce seamless completions for
	// the entire application.
	menu.Carapace = completer

	// Everything is ready for a tour.
	// Run the console and take a look around.
	console.Run()
}

func getCommands() (*cobra.Command, *carapace.Carapace) {
	// Our root command structure encapsulates
	// the entire command tree for our application.
	rootData := &commands.Root{}

	// Options can be used for several purposes:
	// influence the flags naming conventions, register
	// other scan handlers for specialized work, etc...
	var opts []flags.OptFunc

	// One example of specialized handler is the validator,
	// which checks for struct tags specifying validations:
	// when found, this handler wraps the generated flag into
	// a special value which will validate the user input.
	opts = append(opts, flags.Validator(validator.New()))

	// Run the scan: this generates the entire command tree
	// into a cobra root command (and its subcommands).
	// By default, the name of the command is os.Args[0].
	rootCmd := genflags.Generate(rootData, opts...)

	// Since we now dispose of a cobra command, we can further
	// set it up to our liking: modify/set fields and options, etc.
	// There is virtually no restriction to the modifications one
	// can do on them, except that their RunE() is already bound.
	rootCmd.SilenceUsage = true
	rootCmd.Short = shortUsage

	// The completion generator is another example of specialized
	// scan handler: it will generate completers if it finds tags
	// specifying what to complete, or completer implementations
	// by the positional arguments / command flags' types themselves.
	comps, _ := completions.Generate(rootCmd, rootData, nil)

	return rootCmd, comps
}
