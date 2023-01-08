package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/engine"
	"github.com/reeflective/console"
	"github.com/spf13/cobra"
)

// In here we create some menus which hold different command trees.
func createMenus(c *console.Console) {
	clientMenu := c.NewMenu("client")

	// Here, for the sake of demonstrating custom interrupt
	// handlers and for sparing use to write a dedicated command,
	// we use a custom interrupt handler to switch back to main menu.
	clientMenu.AddInterrupt(io.EOF, errorCtrlSwitchMenu)

	// Add custom segments to the prompt for this menu only,
	// and load the configuration making use of them.
	prompt := clientMenu.Prompt()
	prompt.LoadConfig("prompt.omp.json")

	engine.Segments[engine.SegmentType("module")] = func() engine.SegmentWriter { return module }

	// Add some commands to our client menu.
	// This is an example of binding "traditionally defined" cobra.Commands.
	clientMenu.SetCommands(makeClientCommands(c))
}

// errorCtrlSwitchMenu is a custom interrupt handler which will
// switch back to the main menu when the current menu receives
// a CtrlD (io.EOF) error.
func errorCtrlSwitchMenu(c *console.Console) {
	fmt.Println("Switching back to main menu")
	c.SwitchMenu("")
}

// A little set of commands for the client menu, (wrapped so that
// we can pass the console to them, because the console is local).
func makeClientCommands(app *console.Console) console.Commands {
	return func() *cobra.Command {
		root := &cobra.Command{}

		ticker := &cobra.Command{
			Use:   "ticker",
			Short: "Triggers some asynchronous notifications to the shell, demonstrating async logging",
			Run: func(cmd *cobra.Command, args []string) {
				timer := time.Tick(2 * time.Second)
				messages := []string{
					"Info:    notification 1",
					"Info:    notification 2",
					"Warning: notification 3",
					"Info:    notification 4",
					"Error:   done notifying",
				}
				go func() {
					count := 0
					for {
						<-timer
						if count == 5 {
							app.Log("This message is more important, printing it below the prompt first")
							return
						}
						app.LogTransient(messages[count])
						count++
					}
				}()
			},
		}
		root.AddCommand(ticker)

		main := &cobra.Command{
			Use:   "main",
			Short: "A command to return to the main menu (you can also use CtrlD for the same result)",
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Println("Switching back to main menu")
				app.SwitchMenu("")
			},
		}
		root.AddCommand(main)

		shell := &cobra.Command{
			Use:                "!",
			Short:              "Execute the remaining arguments with system shell",
			DisableFlagParsing: true,
			RunE: func(cmd *cobra.Command, args []string) error {
				if len(args) == 0 {
					return errors.New("command requires one or more arguments")
				}

				path, err := exec.LookPath(args[0])
				if err != nil {
					return err
				}

				shellCmd := exec.Command(path, args[1:]...)

				// Load OS environment
				shellCmd.Env = os.Environ()

				out, err := shellCmd.CombinedOutput()
				if err != nil {
					return err
				}

				fmt.Print(string(out))

				return nil
			},
		}
		root.AddCommand(shell)

		return root
	}
}
