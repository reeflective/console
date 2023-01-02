package main

import (
	"github.com/reeflective/console"
)

// configureReadline shows how to access and configure the
// underlying readline shell of the console.
func configureReadline(app *console.Console) {
	// Assuming that the user has a system-wide readline.yml configuration
	// (containing keybinds, completion behavior, etc), load it. If the file
	// does not exist, the readline shell will use default but sane settings.
	app.Shell().Config().LoadSystem()

	// This is very useful when we want completions not to overflow our
	// terminal: the completion system adapts to the available space so
	// that its results, whatever they are, will not get in our way.
	app.Shell().EnableGetCursorPos = true
}
