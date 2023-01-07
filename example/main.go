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

	// Apply some configuration stuff to the underlying readline shell:
	// input modes, indicators, completion and prompt behavior, etc...
	configureReadline(app)

	// Create another menu, different from the main one.
	// It will have its own command tree, prompt engine, history sources, etc.
	createMenus(app)

	// By default the shell as created a single menu and
	// made it current, so you can access it and set it up.
	menu := app.CurrentMenu()

	// All menus currently each have a distinct, in-memory history source.
	// Replace the main (current) menu's history with one writing to our
	// application history file. The default history is named after its menu.
	// menu.DeleteHistorySource("")
	menu.AddHistorySourceFile("local history", ".example-history")

	// We bind a special handler for this menu, which will exit the
	// application (with confirm), when the shell readline receives
	// a Ctrl-D keystroke. You can map any error to any handler.
	menu.AddInterrupt(io.EOF, exitCtrlD)
	menu.AddInterrupt(console.ErrCtrlC, switchMenu)

	// Use the command yielder function and pass it to our menu of interest.
	menu.SetCommands(flagsCommands)

	// Everything is ready for a tour.
	// Run the console and take a look around.
	app.Run()
}
