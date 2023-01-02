package main

import (
	"io"

	"github.com/reeflective/console"
)

const (
	shortUsage = "Console application example, with cobra commands/flags/completions generated from structs"
)

func main() {
	// Instantiate a new app, with a single, default menu.
	// All defaults are set, and nothing is needed to make it work
	app := console.New()

	// Assuming that the user has a system-wide readline.yml configuration
	// (containing keybinds, completion behavior, etc), load it. If the file
	// does not exist, the readline shell will use default but sane settings.
	//
	// This also shows how library consumers can access the
	// underlying readline shell for lower-level configuration.
	app.Shell().Config().LoadSystem()

	// Create another menu, different from the main one.
	// It will have its own command tree, prompt engine, history sources, etc.
	createMenus(app)

	// By default the shell as created a single menu and
	// made it current, so you can access it and set it up.
	menu := app.CurrentMenu()

	// All menus currently each have a distinct, in-memory history source.
	// Replace the main (current) menu's history with one writing to our
	// application history file. The default history is named after its menu.
	// menu.DeleteHistorySource(menu.Name())
	menu.AddHistorySourceFile("local history", ".example-history")

	// We bind a special handler for this menu, which will exit the
	// application (with confirm), when the shell readline receives
	// a Ctrl-D keystroke. You can map any error to any handler.
	menu.AddInterrupt(io.EOF, exitCtrlD)
	menu.AddInterrupt(console.ErrCtrlC, switchMenu)

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
	app.Run()
}
