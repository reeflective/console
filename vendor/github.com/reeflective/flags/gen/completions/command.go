package completions

import (
	"fmt"
	"reflect"

	"github.com/reeflective/flags"
	genflags "github.com/reeflective/flags/gen/flags"
	"github.com/reeflective/flags/internal/scan"
	"github.com/reeflective/flags/internal/tag"
	comp "github.com/rsteube/carapace"
	"github.com/spf13/cobra"
)

var onFinalize func()

// WithReset has a similar role to flags.WithReset(), but for completions:
// when invoked, this function will create a new instance of the command-tree
// struct passed to Generate(), will rescan it and bind new completers with
// their blank/default state.
// In the carapace library, this function is called after each completer invocation.
func WithReset() func() {
	return onFinalize
}

// Gen uses a carapace completion builder to register various completions
// to its underlying cobra command, parsing again the native struct for type
// and struct tags' information.
// Returns the carapace, so you can work with completions should you like.
func Generate(cmd *cobra.Command, data interface{}, comps *comp.Carapace) (*comp.Carapace, error) {
	// Generate the completions a first time.
	completions, err := generate(cmd.Root(), data, comps)
	if err != nil {
		return completions, err
	}

	// And make a handler to be ran after each completion routine,
	// so that commands/flags are reset to their blank/default state.
	onFinalize = func() {
		resetCommands := genflags.WithReset()
		resetCommands()

		// Instantiate a new command struct
		val := reflect.ValueOf(data).Elem()
		data = val.Addr().Interface()

		if _, err := generate(cmd.Root(), data, comps); err != nil {
			return
		}
	}

	return completions, nil
}

// generate wraps all main steps' invocations, to be reused in various cases.
func generate(cmd *cobra.Command, data interface{}, comps *comp.Carapace) (*comp.Carapace, error) {
	if comps == nil {
		comps = comp.Gen(cmd)
	}

	// Each command has, by default, a map of flag completions,
	// which is used for flags that are not contained in a struct group.
	defaultFlagComps := flagSetComps{}

	// A command always accepts embedded subcommand struct fields, so scan them.
	compScanner := completionScanner(cmd, comps, &defaultFlagComps)

	// Scan the struct recursively, for both arg/option groups and subcommands
	if err := scan.Type(data, compScanner); err != nil {
		return comps, fmt.Errorf("%w: %s", scan.ErrScan, err.Error())
	}

	return comps, nil
}

// completionScanner is in charge of building a recursive scanner, working on a given
// struct field at a time, checking for arguments, subcommands and option groups.
func completionScanner(cmd *cobra.Command, comps *comp.Carapace, flags *flagSetComps) scan.Handler {
	handler := func(val reflect.Value, sfield *reflect.StructField) (bool, error) {
		mtag, none, err := tag.GetFieldTag(*sfield)
		if none || err != nil {
			return true, fmt.Errorf("%w: %s", scan.ErrScan, err.Error())
		}

		// If the field is marked as -one or more- positional arguments, we
		// return either on a successful scan of them, or with an error doing so.
		if found, err := positionals(comps, mtag, val); found || err != nil {
			return found, err
		}

		// Else, if the field is marked as a subcommand, we either return on
		// a successful scan of the subcommand, or with an error doing so.
		if found, err := command(cmd, mtag, val); found || err != nil {
			return found, err
		}

		// Else, try scanning the field as a group of commands/options,
		// and only use the completion stuff we find on them.
		if found, err := groupComps(comps, cmd, val, sfield); found || err != nil {
			return found, err
		}

		// Else, try scanning the field as a simple option flag
		return flagComps(comps, flags)(val, sfield)
	}

	return handler
}

// command finds if a field is marked as a command, and if yes, scans it.
func command(cmd *cobra.Command, tag tag.MultiTag, val reflect.Value) (bool, error) {
	// Parse the command name on struct tag...
	name, _ := tag.Get("command")
	if len(name) == 0 {
		return false, nil
	}

	// ... and check the field implements at least the Commander interface
	_, implements, commander := flags.IsCommand(val)
	if !implements {
		return false, nil
	}

	var subc *cobra.Command

	// The command already exists, bound to our current command.
	for _, subcmd := range cmd.Commands() {
		if subcmd.Name() == name {
			subc = subcmd
		}
	}

	if subc == nil {
		return false, fmt.Errorf("%w: %s", errCommandNotFound, name)
	}

	// Simply generate a new carapace around this command,
	// so that we can register different positional arguments
	// without overwriting those of our root command.
	if _, err := generate(subc, commander, nil); err != nil {
		return true, err
	}

	return true, nil
}
