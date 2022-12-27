package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

const (
	validTag = "validate"
)

// invalidVarError wraps an error raised by validator on a struct field,
// and automatically modifies the error string for more efficient ones.
type invalidVarError struct {
	fieldName    string
	fieldValue   string
	validatorErr error
}

// Error implements the Error interface, but replacing some identifiable
// validation errors with more efficient messages, more adapted to CLI.
func (err *invalidVarError) Error() string {
	var tagname string

	// Match the part containing the tag name
	retag := regexp.MustCompile(`the '.*' tag`)

	matched := retag.FindString(err.validatorErr.Error())
	if matched != "" {
		parts := strings.Split(matched, " ")
		if len(parts) > 1 {
			tagname = strings.Trim(parts[1], "'")
		}

		return fmt.Sprintf("%s is not a valid %s", err.fieldValue, tagname)
	}

	// Or simply replace the empty key with the field name.
	return strings.ReplaceAll(err.validatorErr.Error(), "''", fmt.Sprintf("'%s'", err.fieldName))
}

// New returns a validation function to be applied on all flag struct fields of your command tree.
// It makes use of a singleton validator object, not exposed for customizations. If you want to add
// some validations or make any other customizations on the validator, use NewWith(custom) method.
func New() func(val string, field reflect.StructField, cfg interface{}) error {
	valid := validator.New()

	// We wrap this singleton is a correct ValidateFunc
	validation := func(val string, field reflect.StructField, obj interface{}) error {
		validationTag := field.Tag.Get(validTag)
		if err := valid.Var(val, validationTag); err != nil {
			return &invalidVarError{field.Name, val, err}
		}

		return nil
	}

	return validation
}

// NewWith returns a validation function to be applied on all flag struct fields. It takes a go-playground/validator
// object (meant to be used as a singleton), on which the user can prealably register any custom validation routines.
func NewWith(custom *validator.Validate) func(val string, field reflect.StructField, cfg interface{}) error {
	if custom == nil {
		return nil
	}

	// We wrap this singleton is a correct ValidateFunc
	validation := func(val string, field reflect.StructField, obj interface{}) error {
		validationTag := field.Tag.Get(validTag)

		// The val string parameter is probably not needed as we will have to apply / Set() the value
		// on the flag first before validating its value (var interface{} here)
		if err := custom.Var(obj, validationTag); err != nil {
			return &invalidVarError{field.Name, val, err}
		}

		return nil
	}

	return validation
}
