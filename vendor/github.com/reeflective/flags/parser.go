package flags

import (
	"fmt"
	"reflect"
	"unicode/utf8"

	"github.com/reeflective/flags/internal/scan"
	"github.com/reeflective/flags/internal/tag"
	"github.com/reeflective/flags/internal/validation"
)

// ParseStruct parses structure and returns list of flags based on this structure.
// This list of flags can be used by generators for flag, kingpin, cobra, pflag, urfave/cli.
func ParseStruct(cfg interface{}, optFuncs ...OptFunc) ([]*Flag, error) {
	// what we want is Ptr to Structure
	if cfg == nil {
		return nil, ErrObjectIsNil
	}

	v := reflect.ValueOf(cfg)

	if v.Kind() != reflect.Ptr {
		return nil, ErrNotPointerToStruct
	}

	if v.IsNil() {
		return nil, ErrObjectIsNil
	}

	switch e := v.Elem(); e.Kind() {
	case reflect.Struct:
		return parseStruct(e, optFuncs...)
	default:
		return nil, ErrNotPointerToStruct
	}
}

// ParseField parses a single struct field as a list (often only made of only one) flags.
// This function can be used when you want to scan only some fields for which you want a flag.
func ParseField(value reflect.Value, field reflect.StructField, optFuncs ...OptFunc) ([]*Flag, bool, error) {
	// Check struct tags, parse the field value if needed, and return the whole.
	flag, tag, scanOpts, err := parseInfo(field, optFuncs...)
	if err != nil {
		return nil, true, err
	}

	if flag == nil {
		return nil, false, nil
	}

	options := OptFunc(scan.CopyOpts(scanOpts))

	// We might have to scan for an arbitrarily nested structure of flags
	flagSet, val, err := parseVal(value, options)
	if err != nil {
		return flagSet, true, err
	}

	// The flag value, at this point, should always implement the flag Value interface,
	// otherwise it is not considered a valid flag
	if markedFlagNotImplementing(*tag, val) {
		errImpl := fmt.Errorf("%w: field %s (tagged flag '%s') does not implement Value interface",
			ErrNotValue, field.Name, flag.Name)

		return flagSet, true, errImpl
	}

	// If our value is nil, we don't have to perform further validations on it,
	// and we only add flags if we have parsed some on our struct field value.
	if val == nil {
		return flagSet, true, nil
	}

	// Set validators if any, user-defined or builtin
	if validator := validation.Bind(value, field, flag.Choices, scanOpts); validator != nil {
		val = &validateValue{
			Value:        val,
			validateFunc: validator,
		}
	}

	flag.Value = val
	flagSet = append(flagSet, flag)

	// If the user provided some custom flag
	// value handlers/scanners, run on it.
	if scanOpts.FlagFunc != nil {
		var name string
		if flag.Name != "" {
			name = flag.Name
		} else {
			name = flag.Short
		}

		// Any failure to apply a flag handler is reported and will lead to exit
		if err := scanOpts.FlagFunc(name, *tag, value); err != nil {
			return flagSet, true, fmt.Errorf("%w on flag %s: %s", ErrFlagHandler, name, err.Error())
		}
	}

	return flagSet, true, nil
}

// parseInfo parses the struct field tag, adapts for any scan options that would have been modified by tags.
func parseInfo(fld reflect.StructField, optFuncs ...OptFunc) (*Flag, *tag.MultiTag, scan.Opts, error) {
	var scanOpts []scan.OptFunc
	for _, optFunc := range optFuncs {
		scanOpts = append(scanOpts, scan.OptFunc(optFunc))
	}
	scanOptions := scan.DefOpts().Apply(scanOpts...)
	options := opts(scanOptions)

	// skip unexported and non anonymous fields
	if fld.PkgPath != "" && !fld.Anonymous {
		return nil, nil, scanOptions, nil
	}

	// We should have a flag and a tag, legacy or not, and with valid values.
	flag, tag, err := parseFlagTag(fld, options)
	if flag == nil || err != nil {
		return flag, tag, scanOptions, err
	}

	// Various prefixing checks and steps
	flag.EnvName = parseEnvTag(flag.Name, fld, options)
	prefix := flag.Name + options.FlagDivider

	if fld.Anonymous && options.Flatten {
		prefix = options.Prefix
	}

	scanOpts = append(scanOpts, scan.OptFunc(Prefix(prefix)))

	// Return an update list of scan options,
	// which might have been influenced by the tags.
	scanOptions = scanOptions.Apply(scanOpts...)

	return flag, tag, scanOptions, err
}

func parseVal(value reflect.Value, optFuncs ...OptFunc) ([]*Flag, Value, error) {
	// value is addressable, let's check if we can parse it
	if value.CanAddr() && value.Addr().CanInterface() {
		valueInterface := value.Addr().Interface()
		val := parseGenerated(valueInterface)

		if val != nil {
			return nil, val, nil
		}
		// check if field implements Value interface
		if val, casted := valueInterface.(Value); casted {
			return nil, val, nil
		}
	}

	switch value.Kind() {
	case reflect.Ptr:
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}

		val := parseGeneratedPtrs(value.Addr().Interface())

		if val != nil {
			return nil, val, nil
		}

		return parseVal(value.Elem(), optFuncs...)

	case reflect.Struct:
		flags, err := parseStruct(value, optFuncs...)

		return flags, nil, err

	case reflect.Map:
		val := parseMap(value)

		return nil, val, nil
	}

	return nil, nil, nil
}

func parseStruct(value reflect.Value, optFuncs ...OptFunc) ([]*Flag, error) {
	flags := []*Flag{}

	valueType := value.Type()
fields:
	for i := 0; i < value.NumField(); i++ {
		field := valueType.Field(i)
		fieldValue := value.Field(i)
		// skip unexported and non anonymous fields
		if field.PkgPath != "" && !field.Anonymous {
			continue fields
		}

		// Scan the field, potentially a structure, any error stops the process
		fieldFlags, found, err := ParseField(fieldValue, field, optFuncs...)
		if err != nil {
			return flags, err
		}

		if !found || len(fieldFlags) == 0 {
			continue fields
		}

		// And append the flag(s) if we have found some.
		flags = append(flags, fieldFlags...)

		continue fields
	}

	return flags, nil
}

func parseMap(value reflect.Value) Value {
	mapType := value.Type()
	keyKind := value.Type().Key().Kind()

	// check that map key is string or integer
	if !anyOf(MapAllowedKinds, keyKind) {
		return nil
	}

	if value.IsNil() {
		value.Set(reflect.MakeMap(mapType))
	}

	valueInterface := value.Addr().Interface()
	val := parseGeneratedMap(valueInterface)

	return val
}

// Tells us if a struct field tagged as a flag does not implement the Value interface.
func markedFlagNotImplementing(tag tag.MultiTag, val Value) bool {
	_, flagOld := tag.Get("flag")
	_, short := tag.Get("short")
	_, long := tag.Get("long")

	if (flagOld || short || long) && val == nil {
		return true
	}

	return false
}

func anyOf(kinds []reflect.Kind, needle reflect.Kind) bool {
	for _, kind := range kinds {
		if kind == needle {
			return true
		}
	}

	return false
}

func isStringFalsy(s string) bool {
	return s == "" || s == "false" || s == "no" || s == "0"
}

func getShortName(name string) (rune, error) {
	short := rune(0)
	runeCount := utf8.RuneCountInString(name)

	// Either an invalid option name
	if runeCount > 1 {
		msg := fmt.Sprintf("flag `%s'", name)

		return short, newError(ErrShortNameTooLong, msg)
	}

	// Or we have to decode and return
	if runeCount == 1 {
		short, _ = utf8.DecodeRuneInString(name)
	}

	return short, nil
}
