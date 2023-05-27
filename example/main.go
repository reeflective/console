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
	// All defaults are set, and nothing is needed to make it work.
	app := console.New("example")
	app.NewlineBefore = true
	app.NewlineAfter = true

	// Main Menu Setup ---------------------------------------------- //

	// By default the shell as created a single menu and
	// made it current, so you can access it and set it up.
	menu := app.ActiveMenu()

	// All menus currently each have a distinct, in-memory history source.
	// Replace the main (current) menu's history with one writing to our
	// application history file. The default history is named after its menu.
	hist, _ := embeddedHistory(".example-history")
	menu.AddHistorySource("local history", hist)

	// We bind a special handler for this menu, which will exit the
	// application (with confirm), when the shell readline receives
	// a Ctrl-D keystroke. You can map any error to any handler.
	menu.AddInterrupt(io.EOF, exitCtrlD)

	// Make a command yielder for our main menu.
	menu.SetCommands(makeflagsCommands(app))

	// Client Menu Setup -------------------------------------------- //

	// Create another menu, different from the main one.
	// It will have its own command tree, prompt engine, history sources, etc.
	clientMenu := app.NewMenu("client")

	// Here, for the sake of demonstrating custom interrupt
	// handlers and for sparing use to write a dedicated command,
	// we use a custom interrupt handler to switch back to main menu.
	clientMenu.AddInterrupt(io.EOF, errorCtrlSwitchMenu)

	// Add some commands to our client menu.
	// This is an example of binding "traditionally defined" cobra.Commands.
	clientMenu.SetCommands(makeClientCommands(app))

	// Run the app -------------------------------------------------- //

	// Everything is ready for a tour.
	// Run the console and take a look around.
	app.Start()
}
