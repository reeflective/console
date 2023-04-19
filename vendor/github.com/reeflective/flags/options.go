package flags

import (
	"github.com/reeflective/flags/internal/scan"
)

// ValidateFunc describes a validation func, that takes string val for flag from command line,
// field that's associated with this flag in structure cfg. Also works for positional arguments.
// Should return error if validation fails.
type ValidateFunc scan.ValidateFunc

// FlagFunc is a generic function that can be applied to each
// value that will end up being a flags *Flag, so that users
// can perform more arbitrary operations on each, such as checking
// for completer implementations, bind to viper configurations, etc.
type FlagFunc scan.FlagFunc

// OptFunc sets values in opts structure.
type OptFunc scan.OptFunc

type opts scan.Opts

// DescTag sets custom description tag. It is "desc" by default.
func DescTag(val string) OptFunc { return func(opt *scan.Opts) { opt.DescTag = val } }

// FlagTag sets custom flag tag. It is "flag" be default.
func FlagTag(val string) OptFunc { return func(opt *scan.Opts) { opt.FlagTag = val } }

// Prefix sets prefix that will be applied for all flags (if they are not marked as ~).
func Prefix(val string) OptFunc { return func(opt *scan.Opts) { opt.Prefix = val } }

// EnvPrefix sets prefix that will be applied for all environment variables (if they are not marked as ~).
func EnvPrefix(val string) OptFunc { return func(opt *scan.Opts) { opt.EnvPrefix = val } }

// FlagDivider sets custom divider for flags. It is dash by default. e.g. "flag-name".
func FlagDivider(val string) OptFunc { return func(opt *scan.Opts) { opt.FlagDivider = val } }

// EnvDivider sets custom divider for environment variables.
// It is underscore by default. e.g. "ENV_NAME".
func EnvDivider(val string) OptFunc { return func(opt *scan.Opts) { opt.EnvDivider = val } }

// Flatten set flatten option.
// Set to false if you don't want anonymous structure fields to be flatten.
func Flatten(val bool) OptFunc { return func(opt *scan.Opts) { opt.Flatten = val } }

// ParseAll orders the parser to generate a flag for all struct fields,
// even if there isn't a struct tag attached to them.
func ParseAll() OptFunc { return func(opt *scan.Opts) { opt.ParseAll = true } }

// Validator sets validator function for flags.
// Check existing validators in flags/validator and flags/validator/govalidator packages.
func Validator(val ValidateFunc) OptFunc {
	return func(opt *scan.Opts) { opt.Validator = scan.ValidateFunc(val) }
}

// FlagHandler sets the handler function for flags, in order to perform arbitrary
// operations on the value of the flag identified by the <flag> name parameter of FlagFunc.
func FlagHandler(val FlagFunc) OptFunc {
	return func(opt *scan.Opts) { opt.FlagFunc = scan.FlagFunc(val) }
}
