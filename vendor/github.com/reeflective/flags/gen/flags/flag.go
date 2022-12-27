package flags

import (
	"fmt"
	"os"
	"strings"

	"github.com/reeflective/flags"
	"github.com/spf13/pflag"
)

// flagSet describes interface,
// that's implemented by pflag library and required by flags.
type flagSet interface {
	VarPF(value pflag.Value, name, shorthand, usage string) *pflag.Flag
}

var _ flagSet = (*pflag.FlagSet)(nil)

// GenerateTo takes a list of sflag.Flag,
// that are parsed from some config structure, and put it to dst.
func generateTo(src []*flags.Flag, dst flagSet) {
	for _, srcFlag := range src {
		flag := dst.VarPF(srcFlag.Value, srcFlag.Name, srcFlag.Short, srcFlag.Usage)

		// Annotations used for things like completions
		flag.Annotations = map[string][]string{}

		var annots []string

		flag.NoOptDefVal = strings.Join(srcFlag.OptionalValue, " ")

		if boolFlag, casted := srcFlag.Value.(flags.BoolFlag); casted && boolFlag.IsBoolFlag() {
			// pflag uses -1 in this case,
			// we will use the same behaviour as in flag library
			flag.NoOptDefVal = "true"
		} else if srcFlag.Required {
			// Only non-boolean flags can be required.
			annots = append(annots, "required")
		}

		flag.Hidden = srcFlag.Hidden

		if srcFlag.Deprecated {
			// we use Usage as Deprecated message for a pflag
			flag.Deprecated = srcFlag.Usage
			if flag.Deprecated == "" {
				flag.Deprecated = "Deprecated"
			}
		}

		// Register annotations to be used by clients and completers
		flag.Annotations["flags"] = annots
	}
}

// Parse parses cfg, that is a pointer to some structure, puts it to the new
// pflag.FlagSet and returns it.
//
// This is generally not needed if you intend to generate a directly working CLI:
// This function is used for generating things like completions for flags, etc.
func ParseFlags(cfg interface{}, optFuncs ...flags.OptFunc) (*pflag.FlagSet, error) {
	flagSet := pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)

	err := parseTo(cfg, flagSet, optFuncs...)
	if err != nil {
		return nil, err
	}

	return flagSet, nil
}

// parseTo parses cfg, that is a pointer to some structure,
// and puts it to dst.
func parseTo(cfg interface{}, dst flagSet, optFuncs ...flags.OptFunc) error {
	flagSet, err := flags.ParseStruct(cfg, optFuncs...)
	if err != nil {
		return fmt.Errorf("%w: %s", flags.ErrParse, err.Error())
	}

	generateTo(flagSet, dst)

	return nil
}

// ParseToDef parses cfg, that is a pointer to some structure and
// puts it to the default pflag.CommandLine.
func parseToDef(cfg interface{}, optFuncs ...flags.OptFunc) error {
	err := parseTo(cfg, pflag.CommandLine, optFuncs...)
	if err != nil {
		return err
	}

	return nil
}
