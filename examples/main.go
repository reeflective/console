package main

import "github.com/maxlandon/gonsole"

func main() {

	// Instantiate a new console, with a single, default menu.
	// All defaults are set, and nothing is needed to make it work
	console := gonsole.NewConsole()

	// By default the shell as created a single menu and
	// made it current, so you can access it and set it up.
	menu := console.CurrentMenu()

	// Set the prompt (config, for usability purposes). Each menu has its own.
	// See the documentation for more prompt setup possibilities.
	prompt := menu.PromptConfig()
	prompt.Left = "application-name"
	prompt.Multiline = false

	// Add a default help command, that can be used with any command, however nested:
	// 'help <command> <subcommand> <subcommand'
	// The console creates it and attaches it to all existing contexts.
	// "core" is the name of the group in which we will put this command.
	console.AddHelpCommand("core")

	// Add a configuration command if you want your users to be able
	// to modify it on the fly, export it as files or as JSON.
	// Please see the documentation and/or use this example to
	// see what can be done with this.
	console.AddConfigCommand("config", "core")

	// Everything is ready for a tour.
	// Run the console and take a look around.
	console.Run()
}
