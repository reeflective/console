package commands

import (
	"github.com/reeflective/flags/example/args"
	"github.com/reeflective/flags/example/opts"
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
	args.MultipleListsArgs  `command:"multiple-lists" description:"Declare multiple lists as positional arguments, and how words are dispatched" group:"positionals"`
	args.FirstListArgs      `command:"list-first" description:"Use several positionals, of which the first is a list, but not the last." group:"positionals"`
	args.MultipleMinMaxArgs `command:"overlap-min-max" description:"Use multiple lists as positionals, with overlapping min/max requirements" group:"positionals"`
	args.TagCompletedArgs   `command:"tag-completed" description:"Specify completers with struct tags" group:"positionals"`

	//
	// Flags commands -----------------------------------------------------------------
	//
	// These commands demonstrate how to declare command flags and their
	// associated functionality (requirements, completions, validations, etc)
	//
	// First remarks:
	// - As with the 'positionals' commands above, each of these command is embedded individually, and each is tagged with its group name.
	opts.BasicOptions   `command:"basic" alias:"ba" desc:"Shows how to use some basic flags (shows option stacking, and maps)" group:"flags"`
	opts.IgnoredOptions `command:"ignored" desc:"Contains types tagged as flags (automatically initialized), and types to be ignored (not tagged)" group:"flags"`
	opts.DefaultOptions `command:"defaults" desc:"Contains flags with default values, and others with validated choices" group:"flags"`
}
