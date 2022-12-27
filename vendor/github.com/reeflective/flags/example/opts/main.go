package opts

import (
	"github.com/reeflective/flags/gen/completions"
	"github.com/reeflective/flags/gen/flags"
)

//
// This file contains the root command and the main function.
// All options subcommands are bound to this root.
//

type rootCommand struct {
	BasicOptions   `command:"basic" desc:"Shows how to use some basic flags (shows option stacking, and maps)"`
	IgnoredOptions `command:"ignored" desc:"Contains types tagged as flags (automatically initialized), and types to be ignored (not tagged)"`
	DefaultOptions `command:"defaults" desc:"Contains flags with default values, and others with validated choices"`
}

func main() {
	rootData := &rootCommand{}
	rootCmd := flags.Generate(rootData)
	rootCmd.SilenceUsage = true
	rootCmd.Short = "A CLI application showing several ways of declaring and setting (groups of) option flags"

	// Completions (recursive)
	comps, _ := completions.Generate(rootCmd, rootData, nil)
	comps.Standalone()

	// Execute the command (application here)
	if err := rootCmd.Execute(); err != nil {
		return
	}
}
