package flags

import (
	"reflect"
)

// PreRunner is the equivalent of cobra cmd.PreRun(cmd *cobra.Command, args []string).
// The args parameter is populated following the same rules as `Commander.Execute()`.
type PreRunner interface {
	PreRun(args []string)
}

// PreRunnerE is the equivalent of cobra cmd.PreRunE(cmd *cobra.Command, args []string) error
// The args parameter is populated following the same rules as `Commander.Execute()`.
type PreRunnerE interface {
	PreRunE(args []string) error
}

// Commander is the simplest and smallest interface that a type must
// implement to be a valid, local, client command. This command can
// be used either in a single-run CLI app, or in a closed-loop shell.
type Commander interface {
	// Execute runs the command implementation.
	// The args parameter is any argument that has not been parsed
	// neither on any parent command and/or its options, or this
	// command and/or its args/options.
	Execute(args []string) (err error)
}

// Runner is the equivalent of cobra cmd.Run(cmd *cobra.Command, args []string)
// It will be ignored if the `flags.Commander` interface is satisfied.
// The args parameter is populated following the same rules as `Commander.Execute()`.
type Runner interface {
	Run(args []string)
}

// PostRunnerE is the equivalent of cobra cmd.RunE(cmd *cobra.Command, args []string) error
// The args parameter is populated following the same rules as `Commander.Execute()`.
type RunnerE interface {
	RunE(args []string) error
}

// PostRunnerE is the equivalent of cobra cmd.PostRun(cmd *cobra.Command, args []string)
// The args parameter is populated following the same rules as `Commander.Execute()`.
type PostRunner interface {
	PostRun(args []string)
}

// PostRunnerE is the equivalent of cobra cmd.PostRunE(cmd *cobra.Command, args []string) error
// The args parameter is populated following the same rules as `Commander.Execute()`.
type PostRunnerE interface {
	PostRunE(args []string) error
}

// IsCommand checks both tags and implementations on a pointer to a struct,
// initializing the value itself if it's nil (useful for callers).
func IsCommand(val reflect.Value) (reflect.Value, bool, Commander) {
	// Initialize if needed
	var ptrval reflect.Value

	// We just want to get interface, even if nil
	if val.Kind() == reflect.Ptr {
		ptrval = val
	} else {
		ptrval = val.Addr()
	}

	// Assert implementation
	cmd, implements := ptrval.Interface().(Commander)
	if !implements {
		return ptrval, false, nil
	}

	// Once we're sure it's a command, initialize the field if needed.
	if ptrval.IsNil() {
		ptrval.Set(reflect.New(ptrval.Type().Elem()))
	}

	return ptrval, true, cmd
}
