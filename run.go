package console

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kballard/go-shellquote"
	"github.com/spf13/cobra"
)

// Start - Start the console application (readline loop). Blocking.
// The error returned will always be an error that the console
// application does not understand or cannot handle.
func (c *Console) Start() (err error) {
	c.loadActiveHistories()

	// Print the console logo
	if c.printLogo != nil {
		c.printLogo(c)
	}

	for {
		// Always ensure we work with the active menu, with freshly
		// generated commands, bound prompts and some other things.
		menu := c.activeMenu()
		menu.resetPreRun()

		c.printed = false

		c.runPreReadHooks()

		// Block and read user input.
		line, err := c.shell.Readline()
		if err != nil {
			menu.handleInterrupt(err)
			continue
		}

		// Any call to the SwitchMenu() while we were reading user
		// input (through an interrupt handler) might have changed it,
		// so we must be sure we use the good one.
		menu = c.activeMenu()

		// Split the line into shell words.
		args, err := shellquote.Split(line)
		if err != nil {
			c.handleSplitError(err)
			continue
		}

		if len(args) == 0 {
			if c.NewlineAfter {
				fmt.Println()
			}

			continue
		}

		// Run user-provided pre-run line hooks,
		// which may modify the input line args.
		args = c.runLineHooks(args)

		// Run all pre-run hooks and the command itself
		c.execute(menu, args, false)
	}
}

// RunCommand is a convenience function to run a command in a given menu.
// After running, the menu commands are reset, and the prompts reloaded.
func (m *Menu) RunCommand(line string) (err error) {
	if len(line) == 0 {
		return
	}

	// Split the line into shell words.
	args, err := shellquote.Split(line)
	if err != nil {
		return fmt.Errorf("line error: %w", err)
	}

	// The menu used and reset is the active menu.
	// Prepare its output buffer for the command.
	m.resetPreRun()

	// Run the command and associated helpers.
	m.console.execute(m, args, !m.console.isExecuting)

	return
}

// execute - The user has entered a command input line, the arguments have been processed:
// we synchronize a few elements of the console, then pass these arguments to the command
// parser for execution and error handling.
// Our main object of interest is the menu's root command, and we explicitly use this reference
// instead of the menu itself, because if RunCommand() is asynchronously triggered while another
// command is running, the menu's root command will be overwritten.
func (c *Console) execute(menu *Menu, args []string, async bool) (err error) {
	if !async {
		c.mutex.RLock()
		c.isExecuting = true
		c.mutex.RUnlock()
	}

	defer func() {
		c.mutex.RLock()
		c.isExecuting = false
		c.mutex.RUnlock()
	}()

	// Our root command of interest, used throughout this function.
	cmd := menu.Command

	// Find the target command: if this command is filtered, don't run it.
	target, _, _ := cmd.Find(args)
	if c.isFiltered(target) {
		return
	}

	// Console-wide pre-run hooks, cannot.
	c.runPreRunHooks()

	// Assign those arguments to our parser.
	cmd.SetArgs(args)

	if c.NewlineBefore {
		fmt.Println()
	}

	// The command execution should happen in a separate goroutine,
	// and should notify the main goroutine when it is done.
	cmdCtx, cancel := context.WithCancelCause(context.Background())

	cmd.SetContext(cmdCtx)

	// Start monitoring keyboard and OS signals.
	sigchan := c.monitorSignals()

	// And start the command execution.
	go c.executeCommand(cmd, cancel)

	// Wait for the command to finish, or for an OS signal to be caught.
	select {
	case <-cmdCtx.Done():
		err = cmdCtx.Err()
	case signal := <-sigchan:
		cancel(errors.New(signal.String()))
		menu.handleInterrupt(errors.New(signal.String()))
	}

	if c.NewlineAfter {
		fmt.Println()
	}

	return err
}

// Run the command in a separate goroutine, and cancel the context when done.
func (c *Console) executeCommand(cmd *cobra.Command, cancel context.CancelCauseFunc) {
	if err := cmd.Execute(); err != nil {
		cancel(err)
		return
	}

	// And the post-run hooks in the same goroutine,
	// because they should not be skipped even if
	// the command is backgrounded by the user.
	c.runPostRunHooks()

	// Command successfully executed, cancel the context.
	cancel(nil)
}

// Generally, an empty command entered should just print a new prompt,
// unlike for classic CLI usage when the program will print its usage string.
// We simply remove any RunE from the root command, so that nothing is
// printed/executed by default. Pre/Post runs are still used if any.
func (c *Console) ensureNoRootRunner() {
	if c.activeMenu().Command != nil {
		c.activeMenu().RunE = func(cmd *cobra.Command, args []string) error {
			return nil
		}
	}
}

func (c *Console) loadActiveHistories() {
	c.shell.History.Delete()

	for _, name := range c.activeMenu().historyNames {
		c.shell.History.Add(name, c.activeMenu().histories[name])
	}
}

func (c *Console) runPreReadHooks() {
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

// monitorSignals - Monitor the signals that can be sent to the process
// while a command is running. We want to be able to cancel the command.
func (c *Console) monitorSignals() <-chan os.Signal {
	sigchan := make(chan os.Signal, 1)

	signal.Notify(
		sigchan,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		// syscall.SIGKILL,
	)

	return sigchan
}

func (c *Console) handleSplitError(err error) {
	fmt.Printf("Line error: %s\n", err.Error())

	if c.NewlineAfter {
		fmt.Println()
	}
}
