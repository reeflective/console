package completions

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/reeflective/flags"
	genflags "github.com/reeflective/flags/gen/flags"
	"github.com/reeflective/flags/internal/scan"
	"github.com/reeflective/flags/internal/tag"
	comp "github.com/rsteube/carapace"
	"github.com/spf13/cobra"
)

// ErrShortNameTooLong indicates that a short flag name was specified,
// longer than one character.
var ErrShortNameTooLong = errors.New("short names can only be 1 character long")

// flagSetComps is an alias for storing per-flag completions.
type flagSetComps map[string]comp.Action

// flagsGroup finds if a field is marked as a subgroup of options, and if yes, scans it recursively.
func groupComps(comps *comp.Carapace, cmd *cobra.Command, val reflect.Value, fld *reflect.StructField) (bool, error) {
	mtag, none, err := tag.GetFieldTag(*fld)
	if none || err != nil {
		return true, fmt.Errorf("%w: %s", scan.ErrScan, err.Error())
	}

	// If not tagged as group, skip it.
	if _, isGroup := mtag.Get("group"); !isGroup {
		return false, nil
	}

	// description, _ := mtag.Get("description")

	var ptrval reflect.Value

	if val.Kind() == reflect.Ptr {
		ptrval = val

		if ptrval.IsNil() {
			ptrval.Set(reflect.New(ptrval.Type()))
		}
	} else {
		ptrval = val.Addr()
	}

	// We are either waiting for:
	// A group of options ("group" is the legacy name)
	optionsGroup, isSet := mtag.Get("group")

	// Parse the options for completions
	if isSet && optionsGroup != "" {
		err := addFlagComps(comps, mtag, ptrval.Interface())

		return true, err
	}

	// Or a group of commands and/or options, which we also scan,
	// as each command will produce a new carapace, a new set of
	// flag/positional completers, etc
	_, isSet = mtag.Get("commands")

	// Parse for commands
	if isSet {
		defaultFlagComps := flagSetComps{}

		scannerCommand := completionScanner(cmd, comps, &defaultFlagComps)
		err := scan.Type(ptrval.Interface(), scannerCommand)

		return true, fmt.Errorf("%w: %s", scan.ErrScan, err.Error())
	}

	return true, nil
}

// addFlagComps scans a struct (potentially nested), for a set of flags, and without
// binding them to the command, parses them for any completions specified/implemented.
func addFlagComps(comps *comp.Carapace, mtag tag.MultiTag, data interface{}) error {
	var flagOpts []flags.OptFunc

	// New change, in order to easily propagate parent namespaces
	// in heavily/specially nested option groups at bind time.
	delim, _ := mtag.Get("namespace-delimiter")

	namespace, _ := mtag.Get("namespace")
	if namespace != "" {
		flagOpts = append(flagOpts, flags.Prefix(namespace+delim))
	}

	envNamespace, _ := mtag.Get("env-namespace")
	if envNamespace != "" {
		flagOpts = append(flagOpts, flags.EnvPrefix(envNamespace))
	}

	// All completions for this flag set only.
	// The handler will append to the completions map as each flag is parsed
	flagCompletions := flagSetComps{}
	compScanner := flagCompsScanner(&flagCompletions)
	flagOpts = append(flagOpts, flags.FlagHandler(compScanner))

	// Parse the group into a flag set, but don't keep them,
	// we're just interested in running the handler on their values.
	_, err := genflags.ParseFlags(data, flagOpts...)
	if err != nil {
		return fmt.Errorf("%w: %s", flags.ErrParse, err.Error())
	}

	// If we are done parsing the flags without error and we have
	// some completers found on them (implemented or tagged), bind them.
	if len(flagCompletions) > 0 {
		comps.FlagCompletion(comp.ActionMap(flagCompletions))
	}

	return nil
}

// flagScan builds a small struct field handler so that we can scan
// it as an option and add it to our current command flags.
func flagComps(comps *comp.Carapace, flagComps *flagSetComps) scan.Handler {
	flagScanner := func(val reflect.Value, sfield *reflect.StructField) (bool, error) {
		compScanner := flagCompsScanner(flagComps)

		// Parse a single field, returning one or more generic Flags
		_, found, err := flags.ParseField(val, *sfield, flags.FlagHandler(compScanner))
		if err != nil {
			return found, err
		}

		// If we are done parsing the flags without error and we have
		// some completers found on them (implemented or tagged), bind them.
		if len(*flagComps) > 0 {
			comps.FlagCompletion(comp.ActionMap(*flagComps))
		}

		if !found {
			return false, nil
		}

		return true, nil
	}

	return flagScanner
}

// flagCompsScanner builds a scanner that will register some completers for an option flag.
func flagCompsScanner(actions *flagSetComps) flags.FlagFunc {
	handler := func(flag string, tag tag.MultiTag, val reflect.Value) error {
		// First get any completer implementation, and identifies if
		// type is an array, and if yes, where the completer is implemented.
		completer, isRepeatable, itemsImplement := typeCompleter(val)

		// Check if the flag has some choices: if yes, we simply overwrite
		// the completer implementation with a builtin one.
		if choices := choiceCompletions(tag, val); choices != nil {
			completer = choices
			itemsImplement = true
		}

		// Or we might find struct tags specifying some completions,
		// in which case we also override the completer implementation
		if tagged, found := taggedCompletions(tag); found {
			completer = tagged
			itemsImplement = true
		}

		// We are done if no completer is found whatsoever.
		if completer == nil {
			return nil
		}

		// Then, and irrespectively of where the completer comes from,
		// we adapt it considering the kind of type we're dealing with.
		if isRepeatable && itemsImplement {
			(*actions)[flag] = comp.ActionMultiParts(",", func(c comp.Context) comp.Action {
				return completer(c).Invoke(c).Filter(c.Parts).ToA()
			})
		} else {
			(*actions)[flag] = comp.ActionCallback(completer)
		}

		return nil
	}

	return handler
}

// scanOption finds if a field is marked as an option, and if yes, scans it and stores the object.
// func scanOption(mtag tag.MultiTag, field reflect.StructField, val reflect.Value) error {
// longname, _ := mtag.Get("long")                                      DONE
// shortname, _ := mtag.Get("short")                                    DONE
// iniName, _ := mtag.Get("ini-name")
//
// // Need at least either a short or long name
// if longname == "" && shortname == "" && iniName == "" {
//         return nil
// }
//
// short, err := getShortName(shortname)
// if err != nil {
//         return err
// }
//
// description, _ := mtag.Get("description")                            DONE
// def := mtag.GetMany("default")
//
// optionalValue := mtag.GetMany("optional-value")
// valueName, _ := mtag.Get("value-name")
// defaultMask, _ := mtag.Get("default-mask")
//
// optionalTag, _ := mtag.Get("optional")
// optional := !isStringFalsy(optionalTag)
// requiredTag, _ := mtag.Get("required")                               DONE
// required := !isStringFalsy(requiredTag)
// choices := mtag.GetMany("choice")                                    DONE
// hiddenTag, _ := mtag.Get("hidden")
// hidden := !isStringFalsy(hiddenTag)
//
// envDefaultKey, _ := mtag.Get("env")
// envDefaultDelim, _ := mtag.Get("env-delim")
// argsDelim, _ := mtag.Get("args-delim")
//
// option := &Flag{
//         Description:      description,
//         ShortName:        short,
//         LongName:         longname,
//         Default:          def,
//         EnvDefaultKey:    envDefaultKey,
//         EnvDefaultDelim:  envDefaultDelim,
//         OptionalArgument: optional,
//         OptionalValue:    optionalValue,
//         Required:         required,
//         ValueName:        valueName,
//         DefaultMask:      defaultMask,
//         Choices:          choices,
//         Hidden:           hidden,
//
//         // group: g,
//
//         field: field,
//         value: val,
//         tag:   mtag,
// }
//
// if option.isBool() && option.Default != nil {
//         return newErrorf(ErrInvalidTag,
//                 "boolean flag `%s' may not have default values, they always default to `false' and can only be turned on",
//                 option.shortAndLongName())
// }
//
// if len(argsDelim) > 1 {
//         return newErrorf(ErrInvalidTag,
//                 "Argument delimiter for flag `%s' cannot be longer than 1 (rune)",
//                 option.shortAndLongName())
// }
//
// argumentDelim, size := utf8.DecodeRuneInString(argsDelim)
// if size == 0 {
//         argumentDelim, _ = utf8.DecodeRuneInString(defaultArgumentDelimiter)
// }
//
// option.ArgsDelim = argumentDelim

// g.flags = append(g.flags, option)

// 	return nil
// }
