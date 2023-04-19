package positional

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/reeflective/flags/internal/scan"
	"github.com/reeflective/flags/internal/tag"
	"github.com/reeflective/flags/internal/validation"
)

// ScanArgs scans an entire value (must be ensured to be a struct) and creates
// a list of positional arguments, along with many required minimum total number
// of arguments we need. Any non-nil error ends the scan, no matter where.
// The Args object returned is fully ready to parse a line of words onto itself.
func ScanArgs(val reflect.Value, stag tag.MultiTag, opts ...scan.OptFunc) (*Args, error) {
	stype := val.Type()            // Value type of the struct
	req, _ := stag.Get("required") // this is written on the struct, applies to all
	reqAll := len(req) != 0        // Each field will count as one required minimum

	// Prepare our scan options, some of which might be used on our positionals.
	opt := scan.DefOpts().Apply(opts...)

	// Holds our positional slots and manages them
	args := &Args{allRequired: reqAll, noTags: true}

	// Each positional field is scanned for its number requirements,
	// and underlying value to be used by the command's arg handlers/converters.
	for fieldCount := 0; fieldCount < stype.NumField(); fieldCount++ {
		field := stype.Field(fieldCount)
		fieldValue := val.Field(fieldCount)

		// The args objects stores everything related to this slot
		// when parsing is successful, or returns an unrecoverable error.
		err := args.scanArg(field, fieldValue, reqAll, opt)
		if err != nil {
			return nil, err
		}
	}

	// Depending on our position and type, we reset the maximum
	// number of words allowed for this argument, and update the
	// counter that will be used by handlers to sync their use
	// of words
	if err := args.adjustMaximums(); err != nil {
		return args, err
	}

	// Last minute internal counters adjustments
	args.needed = args.totalMin

	// By default, the positionals have a consumer made
	// to parse a list of command words onto our struct.
	args.consumer = args.consumeWords

	return args, nil
}

// scanArg scans a single struct field as positional argument, and sets everything related to it.
func (args *Args) scanArg(field reflect.StructField, value reflect.Value, reqAll bool, opt scan.Opts) error {
	ptag, name, err := parsePositionalTag(field)
	if err != nil {
		return err
	}

	if _, isSet := ptag.Get("required"); isSet {
		args.noTags = false
	}

	// Set min/max requirements depending on the tag, the overall
	// requirement settings (at struct level), also taking into
	// account the kind of field we are considering (slice or not)
	min, max := positionalReqs(value, ptag, reqAll)

	arg := &Arg{
		Index:    len(args.slots),
		Name:     name,
		Minimum:  min,
		Maximum:  max,
		Tag:      ptag,
		StartMin: args.totalMin,
		StartMax: args.totalMax,
		Value:    value,
	}

	args.slots = append(args.slots, arg)
	args.totalMin += min // min is never < 0

	// The total maximum number of arguments is used
	// by completers to know precisely when they should
	// start completing for a given positional field slot.
	if arg.Maximum != -1 {
		args.totalMax += arg.Maximum
	}

	// Set validators
	var choices []string

	choiceTags := ptag.GetMany("choice")

	for _, choice := range choiceTags {
		choices = append(choices, strings.Split(choice, " ")...)
	}

	if validator := validation.Bind(value, field, choices, opt); validator != nil {
		arg.Validator = validator
	}

	return nil
}

// parsePositionalTag extracts and fully parses a struct (positional) field tag.
func parsePositionalTag(field reflect.StructField) (tag.MultiTag, string, error) {
	tag, _, err := tag.GetFieldTag(field)
	if err != nil {
		return tag, field.Name, fmt.Errorf("%w: %s", scan.ErrScan, err)
	}

	name, _ := tag.Get("positional-arg-name")

	if len(name) == 0 {
		name = field.Name
	}

	return tag, name, nil
}

// positionalReqs determines the correct quantity requirements for a positional field,
// depending on its parsed struct tag values, and the underlying type of the field.
func positionalReqs(val reflect.Value, mtag tag.MultiTag, all bool) (min, max int) {
	required, max, set := parseArgsNumRequired(mtag)

	// At least for each requirements are global
	if all && required == 0 {
		min = 1
	}

	// When the argument field is not a slice, we have to adjust for some defaults
	isSlice := val.Type().Kind() == reflect.Slice || val.Type().Kind() == reflect.Map
	if !isSlice {
		max = 1
	}

	switch {
	case !isSlice && required > 0:
		// Individual fields cannot have more than one required
		min = 1
	case !set && !isSlice && all:
		// If we have a struct of untagged fields, but all required,
		// we automatically set min/max to one if the field is individual.
		min = 1
	case set && isSlice && required > 0:
		// If a slice has at least one required, add this minimum
		// Increase the total number of positional args wanted.
		min += required
	}

	return min, max
}

// parseArgsNumRequired sets the minimum/maximum requirements for an argument field.
func parseArgsNumRequired(fieldTag tag.MultiTag) (required, maximum int, set bool) {
	required = 0
	maximum = -1

	sreq, set := fieldTag.Get("required")

	// If no requirements, -1 means unlimited
	if sreq == "" || !set {
		return
	}

	required = 1

	rng := strings.SplitN(sreq, "-", requiredNumParsedValues)

	if len(rng) > 1 {
		if preq, err := strconv.ParseInt(rng[0], baseParseInt, bitsizeParseInt); err == nil {
			required = int(preq)
		}

		if preq, err := strconv.ParseInt(rng[1], baseParseInt, bitsizeParseInt); err == nil {
			maximum = int(preq)
		}
	} else {
		if preq, err := strconv.ParseInt(sreq, baseParseInt, bitsizeParseInt); err == nil {
			required = int(preq)
		}
	}

	return required, maximum, set
}

// adjustMaximums analyzes the position of a positional argument field,
// and adjusts its maximum so that handlers can work on them correctly.
func (args *Args) adjustMaximums() error {
	for _, arg := range args.slots {
		val := arg.Value
		isSlice := val.Type().Kind() == reflect.Slice ||
			val.Type().Kind() == reflect.Map

		// First, the maximum index at which we should start
		// parsing words can never be smaller than the minimum one
		if arg.StartMax < arg.StartMin {
			arg.StartMax = arg.StartMin
		}

		// The maximum is not left to -1 if the field is unique.
		if arg.Maximum == -1 && !isSlice {
			arg.Maximum = 1

			if args.allRequired {
				arg.Minimum = 1
			}

			continue
		}

		if isSlice && args.allRequired && args.noTags {
			arg.Minimum = 1
		}
	}

	return nil
}

// func (args *Args) adjustMaximums() error {
// 	// hasSliceNoMax := false
//
// 	for _, arg := range args.slots {
// 		val := arg.Value
// 		isSlice := val.Type().Kind() == reflect.Slice ||
// 			val.Type().Kind() == reflect.Map
//
// 		// First, the maximum index at which we should start
// 		// parsing words can never be smaller than the minimum one
// 		if arg.StartMax < arg.StartMin {
// 			arg.StartMax = arg.StartMin
// 		}
//
// 		// If we have a slice with no maximum before, it's always
// 		// going to shadow all remaining positional slots.
// 		// if hasSliceNoMax && args.allRequired {
// 		// 	return args.errorSliceShadowing(arg.Name, arg.Index)
// 		// }
//
// 		// The maximum is not left to -1 if the field is unique.
// 		if arg.Maximum == -1 && !isSlice {
// 			arg.Maximum = 1
//
// 			continue
// 		}
//
// 		if isSlice && args.allRequired && args.noTags {
// 			arg.Minimum = 1
// 		}
//
// 		// if isSlice && arg.Maximum == -1 {
// 		// 	hasSliceNoMax = true
// 		// }
// 	}
//
// 	return nil
// }

func (args *Args) errorSliceShadowing(arg string, index int) error {
	shadowed := ""
	for _, arg := range args.slots[index+1:] {
		shadowed += fmt.Sprintf(" `%s`,", arg.Name)
	}
	shadowed = strings.TrimSuffix(shadowed, ",")

	return fmt.Errorf("Positional `%s` is a slice with no maximum: will shadow%s positionals", arg, shadowed)
}
