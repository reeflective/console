package commands

import (
	"github.com/reeflective/flags/example/args"
	"github.com/reeflective/flags/example/opts"
	"github.com/reeflective/flags/example/validated"
	"github.com/spf13/cobra"
)

const (
	ShortUsage = "A CLI application showing various ways to declare positional/flags/commands with structs and fields."
	LongUsage  = `
All of the application's commands for positional arguments comes with:
- An explanation of its behavior and of what it aims to demonstrate
- An extract of the relevant source code.
- Explanations for each of the fields to be found in this source code.

Other commands, demonstrating flags and validations, don't come with
a special or lengthy description, since their behavior is obvious.
Also, most the commands positional arguments and flags/args are completed.
`
)

// Root is the root command of our application, and encapsulates
// all of the application subcommands as embedded structs.
type Root struct {
	//
	// Positional arguments commands -------------------------------------------------
	//
	// These commands demonstrate how to declare positional arguments and their
	// associated functionality (requirements, completions, validations, etc)
	//
	// First remarks:
	// - The commands are registered individually, but each is tagged with a group (eg, they are not grouped in a struct)
	args.MultipleListsArgs  `command:"multiple-ambiguous" description:"Demonstrates an ambiguous use of several lists and tag min-max requirements" group:"positionals"`
	args.FirstListArgs      `command:"list-first" description:"Use several positionals, of which the first is a list, but not the last." group:"positionals"`
	args.MultipleMinMaxArgs `command:"overlap-min-max" description:"Use multiple lists as positionals, with overlapping min/max requirements" group:"positionals"`
	args.TagCompletedArgs   `command:"tag-completed" description:"Specify completers with struct tags" group:"positionals"`
	args.RestSliceMax       `command:"rest-slice-max" desc:"Shows how declaring a rest slice will behave when having a maximum words allowed" group:"positionals"`

	//
	// Flags commands ----------------------------------------------------------------
	//
	// These commands demonstrate how to declare command flags and their
	// associated functionality (requirements, completions, validations, etc)
	//
	// First remarks:
	// - As with the 'positionals' commands above, each of these command is embedded individually, and each is tagged with its group name.
	opts.BasicOptions   `command:"basic" alias:"ba" desc:"Shows how to use some basic flags (shows option stacking, and maps)" group:"flags"`
	opts.IgnoredOptions `command:"ignored" desc:"Contains types tagged as flags (automatically initialized), and types to be ignored (not tagged)" group:"flags"`
	opts.DefaultOptions `command:"defaults" desc:"Contains flags with default values, and others with validated choices" group:"flags"`

	// Validated args/flags ----------------------------------------------------------
	//
	// The following struct demonstrates how to group commands within a struct.
	// This allows to print them under a given heading in the documentation/usage,
	// and optionally to separate your command groups by package/type.
	validated.Commands `commands:"validated"`
}

// AddCommandsLongHelp adds long help messages to our newly generated commands.
func AddCommandsLongHelp(root *cobra.Command) {
	// Positional commands
	cmd, _, _ := root.Find([]string{"multiple-ambiguous"})
	cmd.Long = args.MultipleListsArgsHelp
	cmd, _, _ = root.Find([]string{"list-first"})
	cmd.Long = args.FirstListArgsHelp
	cmd, _, _ = root.Find([]string{"overlap-min-max"})
	cmd.Long = args.MultipleMinMaxArgsHelp
	cmd, _, _ = root.Find([]string{"tag-completed"})
	cmd.Long = args.TagCompletedArgsHelp
	cmd, _, _ = root.Find([]string{"rest-slice-max"})
	cmd.Long = args.RestSliceMaxHelp
}
