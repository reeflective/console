package scan

import (
	"reflect"

	"github.com/reeflective/flags/internal/tag"
)

const (
	DefaultDescTag     = "desc"
	DefaultFlagTag     = "flag"
	DefaultEnvTag      = "env"
	DefaultFlagDivider = "-"
	DefaultEnvDivider  = "_"
	DefaultFlatten     = true
)

// ValidateFunc describes a validation func, that takes string val for flag from command line,
// field that's associated with this flag in structure cfg. Also works for positional arguments.
// Should return error if validation fails.
type ValidateFunc func(val string, field reflect.StructField, cfg interface{}) error

// FlagFunc is a generic function that can be applied to each
// value that will end up being a flags *Flag, so that users
// can perform more arbitrary operations on each, such as checking
// for completer implementations, bind to viper configurations, etc.
type FlagFunc func(flag string, tag tag.MultiTag, val reflect.Value) error

// OptFunc sets values in opts structure.
type OptFunc func(opt *Opts)

type Opts struct {
	DescTag     string
	FlagTag     string
	Prefix      string
	EnvPrefix   string
	FlagDivider string
	EnvDivider  string
	Flatten     bool
	ParseAll    bool
	Validator   ValidateFunc
	FlagFunc    FlagFunc
}

func (o Opts) Apply(optFuncs ...OptFunc) Opts {
	for _, optFunc := range optFuncs {
		optFunc(&o)
	}

	return o
}

func CopyOpts(val Opts) OptFunc { return func(opt *Opts) { *opt = val } }

func DefOpts() Opts {
	return Opts{
		DescTag:     DefaultDescTag,
		FlagTag:     DefaultFlagTag,
		FlagDivider: DefaultFlagDivider,
		EnvDivider:  DefaultEnvDivider,
		Flatten:     DefaultFlatten,
	}
}
