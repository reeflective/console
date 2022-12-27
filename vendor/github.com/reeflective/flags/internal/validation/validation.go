package validation

import (
	"errors"
	"reflect"
	"strings"

	"github.com/reeflective/flags/internal/scan"
)

// ErrInvalidChoice indicates that the provided flag argument is not among the valid choices.
var ErrInvalidChoice = errors.New("invalid choice")

// ValueValidator is the interface implemented by types that can validate a
// flag argument themselves. The provided value is directly passed from the
// command line. This interface has been retroported from jessevdk/go-flags.
type ValueValidator interface {
	// IsValidValue returns an error if the provided string value is valid for
	// the flag.
	IsValidValue(value string) error
}

// Bind builds a validation function including all validation routines (builtin or user-defined) available.
func Bind(value reflect.Value, field reflect.StructField, choices []string, opt scan.Opts) func(val string) error {
	if opt.Validator == nil && len(choices) == 0 {
		return nil
	}

	validation := func(argValue string) error {
		allValues := strings.Split(argValue, ",")

		// The validation is performed on each individual item of a (potential) array
		for _, val := range allValues {
			if len(choices) > 0 {
				if err := validateChoice(val, choices); err != nil {
					return err
				}
			}

			// If choice is valid or arbitrary, run custom validator.
			if opt.Validator != nil {
				if err := opt.Validator(val, field, value.Interface()); err != nil {
					return err
				}
			}

			// Retroporting from jessevdk/go-flags
			if validator, implemented := value.Interface().(ValueValidator); implemented {
				if err := validator.IsValidValue(val); err != nil {
					return err
				}
			}
		}

		return nil
	}

	return validation
}

// validateChoice checks the given value(s) is among valid choices.
func validateChoice(val string, choices []string) error {
	values := strings.Split(val, ",")

	for _, value := range values {
		if !stringInSlice(value, choices) {
			return ErrInvalidChoice
		}
	}

	return nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}

	return false
}
