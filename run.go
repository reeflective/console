package console

import (
	"fmt"

	"github.com/kballard/go-shellquote"
)

// Run - Start the console application (readline loop). Blocking.
// The error returned will always be an error that the console
// application does not understand or cannot handle.
func (c *Console) Run() (err error) {
	// Since we avoid loading prompt engines before running the application
	// (due to a library consumer having to load a custom prompt configuration)
	// we ensure all menus have a non-nil engine.
	c.checkPrompts()

	for {
		c.reloadConfig()          // Rebind the prompt helpers, and similar stuff.
		c.runPreLoopHooks()       // Run user-provided pre-loop hooks
		menu := c.menus.current() // We work with the active menu.

		// Block and read user input. Provides completion, syntax, hints, etc.
		// Various types of errors might arise from here. We handle them in a
		// special function, where we can specify behavior for certain errors.
		line, err := c.shell.Readline()
		if err != nil {
			menu.handleInterrupt(err)

			continue
		}

		// Parse the raw command line into shell-compliant arguments.
		args, err := shellquote.Split(line)
		if err != nil {
			fmt.Printf("Line error: %s\n", err.Error())

			continue
		}

		// Run user-provided pre-run line hooks,
		// which may modify the input line
		args = c.runLineHooks(args)

		// Run all hooks and the command itself
		c.execute(args)
	}
}

func (c *Console) runPreLoopHooks() {
	for _, hook := range c.PreReadlineHooks {
		hook()
	}
}

func (c *Console) runLineHooks(args []string) []string {
	processed := args

	// Or modify them again
	for _, hook := range c.PreCmdRunLineHooks {
		processed, _ = hook(processed)
	}

	return processed
}

func (c *Console) runPreRunHooks() {
	for _, hook := range c.PreCmdRunHooks {
		hook()
	}
}

func (c *Console) runPostRunHooks() {
	for _, hook := range c.PostCmdRunHooks {
		hook()
	}
}

// execute - The user has entered a command input line, the arguments
// have been processed: we synchronize a few elements of the console,
// then pass these arguments to the command parser for execution and error handling.
func (c *Console) execute(args []string) {
	c.runPreRunHooks()

	// Asynchronous messages do not mess with the prompt from now on,
	// until end of execution. Once we are done executing the command,
	// they can again.
	c.mutex.RLock()
	c.isExecuting = true
	c.mutex.RUnlock()

	defer func() {
		c.mutex.RLock()
		c.isExecuting = false
		c.mutex.RUnlock()
	}()

	// Assign those arguments to our parser
	c.menus.current().SetArgs(args)

	// Execute the command line, with the current menu' parser.
	// Process the errors raised by the parser.
	// A few of them are not really errors, and trigger some stuff.
	c.menus.current().Execute()

	c.runPostRunHooks()
}
