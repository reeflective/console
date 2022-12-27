// Package flags generates CLI application commands/flags by parsing structures.
package flags

// Flag structure might be used by cli/flag libraries for their flag generation.
type Flag struct {
	Name       string // name as it appears on command line
	Short      string // optional short name
	EnvName    string
	Usage      string   // help message
	Value      Value    // value as set
	DefValue   []string // default value (as text); for usage message
	Hidden     bool
	Deprecated bool

	// If true, the option _must_ be specified on the command line. If the
	// option is not specified, the parser will generate an ErrRequired type
	// error.
	Required bool

	// If non empty, only a certain set of values is allowed for an option.
	Choices []string

	// The optional value of the option. The optional value is used when
	// the option flag is marked as having an OptionalArgument. This means
	// that when the flag is specified, but no option argument is given,
	// the value of the field this option represents will be set to
	// OptionalValue. This is only valid for non-boolean options.
	OptionalValue []string
}
