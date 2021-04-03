package gonsole

import (
	"fmt"
	"strings"
)

// Run - Start the console application (readline loop). Blocking.
// The error returned will always be an error that the console
// application does not understand or cannot handle.
func (c *Console) Run() (err error) {

	for {
		// Recompute the prompt for the current context
		c.Shell.SetPrompt(c.current.Prompt.render())

		// Set the shell history sources with context ones
		c.Shell.SetHistoryCtrlR(c.current.historyCtrlRName, c.current.historyCtrlR)
		c.Shell.SetHistoryCtrlE(c.current.historyCtrlEName, c.current.historyCtrlE)

		// Instantiate and bind all commands for the current
		// context, respecting any filter used to hide some of them.
		c.bindCommands()

		// Block and read user input. Provides completion, syntax, hints, etc.
		// Various types of errors might arise from here. We handle them
		// in a special function, where we can specify behavior for certain errors.
		line, err := c.Shell.Readline()
		if err != nil {

		}

		// The user has entered an input line command. Any previous errors
		// have been handled, and we will go all the way toward command execution,
		// even if the command line is empty. If the context prompt is asked to
		// leave a newline between prompt and output, we print it now.
		if c.LeaveNewline {
			fmt.Println()
		}

		// The line might need some sanitization, like removing empty/redundant spaces,
		// but also in case where there are weird slashes and other kung-fu bombs.
		args, empty := c.sanitizeInput(line)
		if empty {
			continue
		}

		// We then pass the processed command line to the command parser,
		// where any error arising from parsing or execution will be handled.
		// Thus we don't need to handle any error here.
		c.execute(args)
	}
}

// sanitizeInput - Trims spaces and other unwished elements from the input line.
func (c *Console) sanitizeInput(line string) (sanitized []string, empty bool) {

	// Assume the input is not empty
	empty = false

	// Trim border spaces
	trimmed := strings.TrimSpace(line)
	if len(line) < 1 {
		empty = true
		return
	}
	unfiltered := strings.Split(trimmed, " ")

	// Catch any eventual empty items
	for _, arg := range unfiltered {
		if arg != "" {
			sanitized = append(sanitized, arg)
		}
	}

	return
}
