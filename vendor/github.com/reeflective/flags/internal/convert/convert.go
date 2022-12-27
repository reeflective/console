package convert

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/reeflective/flags/internal/tag"
)

const (
	baseParseInt            = 10
	bitsizeParseInt         = 32
	requiredNumParsedValues = 2
)

// Internal errors.
var (
	errStringer    = errors.New("type assertion to `fmt.Stringer` failed")
	errUnmarshaler = errors.New("type assertion to `flags.Unmarshaler` failed")
)

// ErrConvertion is used to notify that converting
// a string value onto a native type has failed.
var ErrConvertion = errors.New("conversion error")

// marshaler is the interface implemented by types that can marshal themselves
// to a string representation of the flag. Retroported from jessevdk/go-flags.
type marshaler interface {
	// MarshalFlag marshals a flag value to its string representation.
	MarshalFlag() (string, error)
}

// unmarshaler is the interface implemented by types that can unmarshal a flag
// argument to themselves. The provided value is directly passed from the
// command line. Retroported from jessevdk/go-flags.
type unmarshaler interface {
	// UnmarshalFlag unmarshals a string value representation to the flag
	// value (which therefore needs to be a pointer receiver).
	UnmarshalFlag(value string) error
}

// --------------------------------------------------------------------------------------------------- //
//                                             Internal                                                //
// --------------------------------------------------------------------------------------------------- //
//
// 1) Main, entrypoint convert functions
// 2) Per-type convert functions (dispatched by entrypoints)
// 3) Other helpers

//
// 1) Main, entrypoint convert functions ----------------------------------------------- //
//

// Value converts a string to its underlying/native value type, therefore
// directly applying this value on the struct field it was created from.
func Value(val string, retval reflect.Value, options tag.MultiTag) error {
	// Use unmarshaller if available/possible
	if ok, err := convertUnmarshal(val, retval); ok {
		return err
	}

	valType := retval.Type()

	// Support for time.Duration
	if valType == reflect.TypeOf((*time.Duration)(nil)).Elem() {
		return convertDuration(val, retval)
	}

	switch valType.Kind() {
	// Strings & bools
	case reflect.String:
		retval.SetString(val)
	case reflect.Bool:
		return convertBool(val, retval)

	// Numbers
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return convertInt(val, valType, retval, options)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return convertUint(val, valType, retval, options)
	case reflect.Float32, reflect.Float64:
		return convertFloat(val, valType, retval)

	// Arrays
	case reflect.Slice:
		return convertSlice(val, valType, retval, options)
	case reflect.Map:
		return convertMap(val, valType, retval, options)

	// Types
	case reflect.Ptr:
		if retval.IsNil() {
			retval.Set(reflect.New(retval.Type().Elem()))
		}

		return Value(val, reflect.Indirect(retval), options)
	case reflect.Interface:
		if !retval.IsNil() {
			return Value(val, retval.Elem(), options)
		}
	}

	return nil
}

func convertToString(val reflect.Value, options tag.MultiTag) (string, error) {
	if ok, ret, err := convertMarshal(val); ok {
		return ret, err
	}

	if !val.IsValid() {
		return "", nil
	}

	valType := val.Type()

	// Support for time.Duration
	if valType == reflect.TypeOf((*time.Duration)(nil)).Elem() {
		stringer, ok := val.Interface().(fmt.Stringer)
		if !ok {
			return "", fmt.Errorf("convert duration: %w", errStringer)
		}

		return stringer.String(), nil
	}

	switch valType.Kind() {
	case reflect.String:
		return val.String(), nil
	case reflect.Bool:
		return convertBoolStr(val)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return convertIntStr(val, options)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return convertUintStr(val, options)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(val.Float(), 'g', -1, valType.Bits()), nil
	case reflect.Slice:
		return convertSliceStr(val, options)
	case reflect.Map:
		return convertMapStr(val, options)
	case reflect.Ptr:
		return convertToString(reflect.Indirect(val), options)
	case reflect.Interface:
		if !val.IsNil() {
			return convertToString(val.Elem(), options)
		}
	}

	return "", nil
}

func convertMarshal(val reflect.Value) (bool, string, error) {
	// Check first for the Marshaler interface
	if val.IsValid() && val.Type().NumMethod() > 0 && val.CanInterface() {
		if marshaler, ok := val.Interface().(marshaler); ok {
			ret, err := marshaler.MarshalFlag()

			return true, ret, fmt.Errorf("marshal error: %w", err)
		}
	}

	return false, "", nil
}

func convertUnmarshal(val string, retval reflect.Value) (bool, error) {
	// Use any unmarshalling implementation found on the concrete type.
	if unm, found := typeIsUnmarshaller(retval); found && unm != nil {
		return convertWithUnmarshaler(val, retval, unm)
	}

	// Or recursively call ourselves with embedded types
	if retval.Type().Kind() != reflect.Ptr && retval.CanAddr() {
		return convertUnmarshal(val, retval.Addr())
	}

	if retval.Type().Kind() == reflect.Interface && !retval.IsNil() {
		return convertUnmarshal(val, retval.Elem())
	}

	return false, nil
}

func convertWithUnmarshaler(val string, retval reflect.Value, unm unmarshaler) (bool, error) {
	// If we have an existing value, just use it
	if !retval.IsNil() {
		if err := unm.UnmarshalFlag(val); err != nil {
			return true, fmt.Errorf("unmarshal error: %w", err)
		}

		return true, nil
	}

	// Else we need to re-assign from the new value
	retval.Set(reflect.New(retval.Type().Elem()))

	unm, found := retval.Interface().(unmarshaler)
	if !found {
		return false, fmt.Errorf("convert marshal: %w", errUnmarshaler)
	}

	// And finally perform the custom unmarshaling
	if err := unm.UnmarshalFlag(val); err != nil {
		return true, fmt.Errorf("unmarshal error: %w", err)
	}

	return true, nil
}

//
// 2) Per-type convert functions (dispatched by entrypoints) ------------------------------- //
//

func convertDuration(val string, retval reflect.Value) error {
	parsed, err := time.ParseDuration(val)
	if err != nil {
		return fmt.Errorf("convert duration: %w", err)
	}

	retval.SetInt(int64(parsed))

	return nil
}

func convertBool(val string, retval reflect.Value) error {
	if val == "" {
		retval.SetBool(true)
	} else {
		value, err := strconv.ParseBool(val)
		if err != nil {
			return fmt.Errorf("convert bool: %w", err)
		}

		retval.SetBool(value)
	}

	return nil
}

func convertBoolStr(val reflect.Value) (string, error) {
	if val.Bool() {
		return "true", nil
	}

	return "false", nil
}

func convertInt(val string, valType reflect.Type, retval reflect.Value, options tag.MultiTag) error {
	base, err := getBase(options, baseParseInt)
	if err != nil {
		return err
	}

	parsed, err := strconv.ParseInt(val, base, valType.Bits())
	if err != nil {
		return fmt.Errorf("convert int: %w", err)
	}

	retval.SetInt(parsed)

	return nil
}

func convertIntStr(val reflect.Value, options tag.MultiTag) (string, error) {
	base, err := getBase(options, baseParseInt)
	if err != nil {
		return "", err
	}

	return strconv.FormatInt(val.Int(), base), nil
}

func convertUint(val string, valType reflect.Type, retval reflect.Value, options tag.MultiTag) error {
	base, err := getBase(options, baseParseInt)
	if err != nil {
		return err
	}

	parsed, err := strconv.ParseUint(val, base, valType.Bits())
	if err != nil {
		return fmt.Errorf("convert uint: %w", err)
	}

	retval.SetUint(parsed)

	return nil
}

func convertUintStr(val reflect.Value, options tag.MultiTag) (string, error) {
	base, err := getBase(options, baseParseInt)
	if err != nil {
		return "", err
	}

	return strconv.FormatUint(val.Uint(), base), nil
}

func convertFloat(val string, valType reflect.Type, retval reflect.Value) error {
	parsed, err := strconv.ParseFloat(val, valType.Bits())
	if err != nil {
		return fmt.Errorf("convert float: %w", err)
	}

	retval.SetFloat(parsed)

	return nil
}

func convertSlice(val string, valType reflect.Type, retval reflect.Value, options tag.MultiTag) error {
	elemtp := valType.Elem()

	elemvalptr := reflect.New(elemtp)
	elemval := reflect.Indirect(elemvalptr)

	if err := Value(val, elemval, options); err != nil {
		return err
	}

	retval.Set(reflect.Append(retval, elemval))

	return nil
}

func convertSliceStr(val reflect.Value, options tag.MultiTag) (string, error) {
	if val.Len() == 0 {
		return "", nil
	}

	ret := "["

	for i := 0; i < val.Len(); i++ {
		if i != 0 {
			ret += ", "
		}

		item, err := convertToString(val.Index(i), options)
		if err != nil {
			return "", err
		}

		ret += item
	}

	return ret + "]", nil
}

func convertMap(val string, valType reflect.Type, retval reflect.Value, options tag.MultiTag) error {
	parts := strings.SplitN(val, ":", requiredNumParsedValues)

	key := parts[0]

	var value string

	if len(parts) == requiredNumParsedValues {
		value = parts[1]
	}

	keytp := valType.Key()
	keyval := reflect.New(keytp)

	if err := Value(key, keyval, options); err != nil {
		return err
	}

	valuetp := valType.Elem()
	valueval := reflect.New(valuetp)

	if err := Value(value, valueval, options); err != nil {
		return err
	}

	if retval.IsNil() {
		retval.Set(reflect.MakeMap(valType))
	}

	retval.SetMapIndex(reflect.Indirect(keyval), reflect.Indirect(valueval))

	return nil
}

func convertMapStr(val reflect.Value, options tag.MultiTag) (string, error) {
	ret := "{"

	for i, key := range val.MapKeys() {
		if i != 0 {
			ret += ", "
		}

		keyitem, err := convertToString(key, options)
		if err != nil {
			return "", err
		}

		item, err := convertToString(val.MapIndex(key), options)
		if err != nil {
			return "", err
		}

		ret += keyitem + ":" + item
	}

	return ret + "}", nil
}

//
// 3) Other helpers ------------------------------------------------------------------------ //
//

func typeIsUnmarshaller(retval reflect.Value) (unmarshaler, bool) {
	if retval.Type().NumMethod() == 0 || retval.CanInterface() {
		return nil, false
	}

	if unm, isImplemented := retval.Interface().(unmarshaler); isImplemented {
		return unm, true
	}

	return nil, false
}

func getBase(options tag.MultiTag, base int) (int, error) {
	var err error

	var ivbase int64

	if sbase, _ := options.Get("base"); sbase != "" {
		ivbase, err = strconv.ParseInt(sbase, baseParseInt, bitsizeParseInt)
		base = int(ivbase)
	}

	if err != nil {
		return base, fmt.Errorf("base int: %w", err)
	}

	return base, nil
}

func isPrint(s string) bool {
	for _, c := range s {
		if !strconv.IsPrint(c) {
			return false
		}
	}

	return true
}

func quoteIfNeeded(s string) string {
	if !isPrint(s) {
		return strconv.Quote(s)
	}

	return s
}

func quoteIfNeededV(s []string) []string {
	ret := make([]string, len(s))

	for i, v := range s {
		ret[i] = quoteIfNeeded(v)
	}

	return ret
}

func quoteV(s []string) []string {
	ret := make([]string, len(s))

	for i, v := range s {
		ret[i] = strconv.Quote(v)
	}

	return ret
}

func unquoteIfPossible(s string) (string, error) {
	if len(s) == 0 || s[0] != '"' {
		return s, nil
	}

	unquoted, err := strconv.Unquote(s)
	if err != nil {
		return unquoted, fmt.Errorf("error unquoting: %w", err)
	}

	return unquoted, nil
}
