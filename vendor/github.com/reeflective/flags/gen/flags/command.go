package flags

import (
	"fmt"
	"os"
	"reflect"

	"github.com/reeflective/flags"
	"github.com/reeflective/flags/internal/scan"
	"github.com/reeflective/flags/internal/tag"
	"github.com/spf13/cobra"
)

var onFinalize func()

// WithReset registers a function to cobra finalizers: when invoked, this function
// will create a new instance of the command-tree struct passed to Generate(), will
// rescan it and bind new commands with their blank/default state.
// This should normally only be used if you are using commands within a closed
// loop Go program/shell, where commands will be invoked more than once.
//
// The associated function is returned, so that programs can reuse it at other times.
func WithReset() func() {
	cobra.OnFinalize(onFinalize)
	return onFinalize
}

// Generate returns a root cobra Command to be used directly as an entry-point.
// The data interface parameter can be nil, or arbitrarily:
// - A simple group of options to bind at the local, root level
// - A struct containing substructs for postional parameters, and other with options.
func Generate(data interface{}, opts ...flags.OptFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:         os.Args[0],
		Annotations: map[string]string{},
	}

	// Scan the struct and bind all commands to this root.
	generate(cmd, data, opts...)

	// Make a handler to be used when this command
	// tree is used in a closed-loop Go program/shell.
	onFinalize = func() {
		cmd.ResetCommands()

		// Instantiate a new command struct
		val := reflect.ValueOf(data)
		if val.Kind() == reflect.Ptr {
			val = reflect.Indirect(val)
		}

		data := reflect.New(val.Type()).Interface()

		// And scan again to rebind all commands
		// to their blank/default state.
		generate(cmd, data, opts...)
	}

	return cmd
}

// generate wraps all main steps' invocations, to be reused in various cases.
func generate(cmd *cobra.Command, data interface{}, opts ...flags.OptFunc) {
	// Make a scan handler that will run various scans on all
	// the struct fields, with arbitrary levels of nesting.
	scanner := scanRoot(cmd, nil, opts)

	// And scan the struct recursively, for arg/option groups and subcommands
	if err := scan.Type(data, scanner); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err.Error())
		os.Exit(1)
	}

	// Subcommands, optional or not
	if cmd.HasSubCommands() {
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			return nil
		}
	} else if _, isCmd, impl := flags.IsCommand(reflect.ValueOf(data)); isCmd {
		setRuns(cmd, impl)
	}
}

// scan is in charge of building a recursive scanner, working on a given struct field at a time,
// checking for arguments, subcommands and option groups. It also checks if additional handlers
// should be applied on the given struct field, such as when our application can run itself as
// a module.
func scanRoot(cmd *cobra.Command, group *cobra.Group, opts []flags.OptFunc) scan.Handler {
	handler := func(val reflect.Value, sfield *reflect.StructField) (bool, error) {
		// Parse the tag or die tryin. We should find one, or we're not interested.
		mtag, _, err := tag.GetFieldTag(*sfield)
		if err != nil {
			return true, fmt.Errorf("%w: %s", tag.ErrTag, err.Error())
		}

		// First, having a tag means this field should have our attention for
		// something, whether it be an option, a positional, something to include
		// as a module option, and so on. If we have to run some handlers no matter
		// what the outcome is, we defer them first.

		// If the field is marked as -one or more- positional arguments, we
		// return either on a successful scan of them, or with an error doing so.
		if found, err := positionals(cmd, mtag, val, opts); found || err != nil {
			return found, err
		}

		// Else, if the field is marked as a subcommand, we either return on
		// a successful scan of the subcommand, or with an error doing so.
		if found, err := command(cmd, group, mtag, val, opts); found || err != nil {
			return found, err
		}

		// Else, if the field is a struct group of options
		if found, err := flagsGroup(cmd, val, sfield, opts); found || err != nil {
			return found, err
		}

		// Else, try scanning the field as a simple option flag
		return flagScan(cmd, opts)(val, sfield)
	}

	return handler
}

// command finds if a field is marked as a subcommand, and if yes, scans it. We have different cases:
// - When our application can run its commands as modules, we must build appropriate handlers.
func command(cmd *cobra.Command, grp *cobra.Group, tag tag.MultiTag, val reflect.Value, opts []flags.OptFunc) (bool, error) {
	// Parse the command name on struct tag...
	name, _ := tag.Get("command")
	if len(name) == 0 {
		return false, nil
	}

	// ... and check the field implements at least the Commander interface
	val, implements, cmdType := flags.IsCommand(val)
	if !implements && len(name) != 0 && cmdType == nil {
		return false, flags.ErrNotCommander
	} else if !implements && len(name) == 0 {
		return false, nil // Skip to next field
	}

	// Always populate the maximum amount of information
	// in the new subcommand, so that when it scans recursively,
	// we can have a more granular context.
	subc := newCommand(name, tag, grp)

	// Set the group to which the subcommand belongs
	tagged, _ := tag.Get("group")
	setGroup(cmd, subc, grp, tagged)

	// Bind the various pre/run/post implementations of our command.
	setRuns(subc, cmdType)

	// Scan the struct recursively, for arg/option groups and subcommands
	scanner := scanRoot(subc, grp, opts)
	if err := scan.Type(val.Interface(), scanner); err != nil {
		return true, fmt.Errorf("%w: %s", scan.ErrScan, err.Error())
	}

	// If we have more than one subcommands and that we are NOT
	// marked has having optional subcommands, remove our run function
	// function, so that help printing can behave accordingly.
	if _, isSet := tag.Get("subcommands-optional"); !isSet {
		if len(subc.Commands()) > 0 {
			cmd.RunE = nil
		}
	}

	// And bind this subcommand back to us
	cmd.AddCommand(subc)

	return true, nil
}

// builds a quick command template based on what has been specified through tags, and in context.
func newCommand(name string, mtag tag.MultiTag, parent *cobra.Group) *cobra.Command {
	subc := &cobra.Command{
		Use:         name,
		Annotations: map[string]string{},
	}

	if desc, _ := mtag.Get("description"); desc != "" {
		subc.Short = desc
	} else if desc, _ := mtag.Get("desc"); desc != "" {
		subc.Short = desc
	}

	subc.Long, _ = mtag.Get("long-description")
	subc.Aliases = mtag.GetMany("alias")
	_, subc.Hidden = mtag.Get("hidden")

	return subc
}

// setRuns binds the various pre/run/post implementations to a cobra command.
func setRuns(cmd *cobra.Command, impl flags.Commander) {
	// No implementation means that this command
	// requires subcommands by default.
	if impl == nil {
		return
	}

	// If our command hasn't any positional argument handler,
	// we must make one to automatically put any of them in Execute
	cmd.Args = func(cmd *cobra.Command, args []string) error {
		setRemainingArgs(cmd, args)

		return nil
	}

	// The args passed to the command have already been parsed,
	// this is why we mute the args []string function parameter.
	cmd.RunE = func(c *cobra.Command, _ []string) error {
		retargs := getRemainingArgs(c)
		cmd.SetArgs(retargs)

		return impl.Execute(retargs)
	}
}

func setGroup(parent, subc *cobra.Command, parentGroup *cobra.Group, tagged string) {
	var group *cobra.Group

	// The group tag on the command has priority
	if tagged != "" {
		for _, grp := range parent.Groups() {
			if grp.ID == tagged {
				group = grp
			}
		}

		if group == nil {
			group = &cobra.Group{ID: tagged, Title: tagged}
			parent.AddGroup(group)
		}
	} else if parentGroup != nil {
		group = parentGroup
	}

	// Use the group we settled on
	if group != nil {
		subc.GroupID = group.ID
	}
}
