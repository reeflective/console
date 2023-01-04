package main

import (
	"fmt"
	"io"

	"github.com/jandedobbeleer/oh-my-posh/src/engine"
	"github.com/reeflective/console"
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
}

// errorCtrlSwitchMenu is a custom interrupt handler which will
// switch back to the main menu when the current menu receives
// a CtrlD (io.EOF) error.
func errorCtrlSwitchMenu(c *console.Console) {
	fmt.Println("Switching back to main menu")
	c.SwitchMenu("")
}
