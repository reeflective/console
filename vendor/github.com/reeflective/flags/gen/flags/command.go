package flags

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/reeflective/flags"
	"github.com/reeflective/flags/internal/scan"
	"github.com/reeflective/flags/internal/tag"
	"github.com/spf13/cobra"
)

// Generate returns a root cobra Command to be used directly as an entry-point.
// The data interface parameter can be nil, or arbitrarily:
// - A simple group of options to bind at the local, root level
// - A struct containing substructs for postional parameters, and other with options.
func Generate(data interface{}, opts ...flags.OptFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:              os.Args[0],
		Annotations:      map[string]string{},
		TraverseChildren: true,
	}

	// Scan the struct and bind all commands to this root.
	generate(cmd, data, opts...)

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
		cmd.RunE = unknownSubcommandAction
	} else {
		setRuns(cmd, data)
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

	// Initialize the field if nil
	data := initialize(val)

	// Always populate the maximum amount of information
	// in the new subcommand, so that when it scans recursively,
	// we can have a more granular context.
	subc := newCommand(name, tag, grp)

	// Set the group to which the subcommand belongs
	tagged, _ := tag.Get("group")
	setGroup(cmd, subc, grp, tagged)

	// Scan the struct recursively, for arg/option groups and subcommands
	scanner := scanRoot(subc, grp, opts)
	if err := scan.Type(data, scanner); err != nil {
		return true, fmt.Errorf("%w: %s", scan.ErrScan, err.Error())
	}

	// Bind the various pre/run/post implementations of our command.
	if _, isSet := tag.Get("subcommands-optional"); !isSet && subc.HasSubCommands() {
		subc.RunE = unknownSubcommandAction
	} else {
		data := initialize(val)
		setRuns(subc, data)
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

func unknownSubcommandAction(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}

	err := fmt.Sprintf("unknown subcommand %q for %q", args[0], cmd.Name())

	if suggestions := cmd.SuggestionsFor(args[0]); len(suggestions) > 0 {
		err += "\n\nDid you mean this?\n"
		for _, s := range suggestions {
			err += fmt.Sprintf("\t%v\n", s)
		}

		err = strings.TrimSuffix(err, "\n")
	}

	return fmt.Errorf(err)
}

func setRuns(cmd *cobra.Command, data interface{}) {
	// No implementation means that this command
	// requires subcommands by default.
	if data == nil {
		return
	}

	// If our command hasn't any positional argument handler,
	// we must make one to automatically put any of them in Execute
	if cmd.Args == nil {
		cmd.Args = func(cmd *cobra.Command, args []string) error {
			setRemainingArgs(cmd, args)

			return nil
		}
	}

	// Pre-runners
	if runner, ok := data.(flags.PreRunner); ok && runner != nil {
		cmd.PreRun = func(c *cobra.Command, _ []string) {
			retargs := getRemainingArgs(c)
			runner.PreRun(retargs)
		}
	}
	if runner, ok := data.(flags.PreRunnerE); ok && runner != nil {
		cmd.PreRunE = func(c *cobra.Command, _ []string) error {
			retargs := getRemainingArgs(c)
			return runner.PreRunE(retargs)
		}
	}

	// Runners
	if commander, ok := data.(flags.Commander); ok && commander != nil {
		cmd.RunE = func(c *cobra.Command, _ []string) error {
			retargs := getRemainingArgs(c)
			cmd.SetArgs(retargs)
			return commander.Execute(retargs)
		}
	} else if runner, ok := data.(flags.RunnerE); ok && runner != nil {
		cmd.RunE = func(c *cobra.Command, _ []string) error {
			retargs := getRemainingArgs(c)
			return runner.RunE(retargs)
		}
	}

	if runner, ok := data.(flags.Runner); ok && runner != nil {
		cmd.Run = func(c *cobra.Command, _ []string) {
			retargs := getRemainingArgs(c)
			runner.Run(retargs)
		}
	}

	// Post-runners
	if runner, ok := data.(flags.PostRunner); ok && runner != nil {
		cmd.PreRun = func(c *cobra.Command, _ []string) {
			retargs := getRemainingArgs(c)
			runner.PostRun(retargs)
		}
	}
	if runner, ok := data.(flags.PostRunnerE); ok && runner != nil {
		cmd.PreRunE = func(c *cobra.Command, _ []string) error {
			retargs := getRemainingArgs(c)
			return runner.PostRunE(retargs)
		}
	}
}

func initialize(val reflect.Value) interface{} {
	// Initialize if needed
	var ptrval reflect.Value

	// We just want to get interface, even if nil
	if val.Kind() == reflect.Ptr {
		ptrval = val
	} else {
		ptrval = val.Addr()
	}

	// Once we're sure it's a command, initialize the field if needed.
	if ptrval.IsNil() {
		ptrval.Set(reflect.New(ptrval.Type().Elem()))
	}

	return ptrval.Interface()
}
