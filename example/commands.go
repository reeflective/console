package main

import (
	"fmt"

	"github.com/reeflective/console"
	"github.com/reeflective/flags"
	"github.com/reeflective/flags/example/commands"
	"github.com/reeflective/flags/gen/completions"
	genflags "github.com/reeflective/flags/gen/flags"
	"github.com/reeflective/flags/validator"
	"github.com/spf13/cobra"
)

// We wrap our command generation routine in this because one of the commands
// needs access to the console (local var in this example app).
// In general, you would just need to declare the inner lambda function.
func makeflagsCommands(app *console.Console) console.Commands {
	// flagsCommands is the `Commands()` function example when
	// you make use of the `reeflective/flags` generation library.
	return func() *cobra.Command {
		rootData := &commands.Root{}

		// Options can be used for several purposes:
		// influence the flags naming conventions, register
		// other scan handlers for specialized work, etc...
		var opts []flags.OptFunc

		// One example of specialized handler is the validator,
		// which checks for struct tags specifying validations:
		// when found, this handler wraps the generated flag into
		// a special value which will validate the user input.
		opts = append(opts, flags.Validator(validator.New()))

		// Run the scan: this generates the entire command tree
		// into a cobra root command (and its subcommands).
		// By default, the name of the command is os.Args[0].
		rootCmd := genflags.Generate(rootData, opts...)

		// Since we now dispose of a cobra command, we can further
		// set it up to our liking: modify/set fields and options, etc.
		// There is virtually no restriction to the modifications one
		// can do on them, except that their RunE() is already bound.
		rootCmd.SilenceUsage = true
		rootCmd.Short = shortUsage
		rootCmd.Long = shortUsage + "\n" + commands.LongUsage

		// We might also have longer help strings contained in our
		// various commands' packages, which we also bind now.
		commands.AddCommandsLongHelp(rootCmd)

		// The completion generator is another example of specialized
		// scan handler: it will generate completers if it finds tags
		// specifying what to complete, or completer implementations
		// by the positional arguments / command flags' types themselves.
		completions.Generate(rootCmd, rootData, nil)

		// And let's add a command declared in a traditional "cobra" way.
		clientMenuCommand := &cobra.Command{
			Use:     "client",
			Short:   "Switch to the client menu (also works with CtrlC)",
			GroupID: "core",
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Println("Switching to client menu")
				app.SwitchMenu("client")
			},
		}
		rootCmd.AddGroup(&cobra.Group{ID: "core", Title: "core"})
		rootCmd.AddCommand(clientMenuCommand)

		rootCmd.SetHelpCommandGroupID("core")
		rootCmd.InitDefaultHelpCmd()

		return rootCmd
	}
}
