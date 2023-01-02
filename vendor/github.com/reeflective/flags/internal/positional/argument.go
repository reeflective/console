package positional

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/reeflective/flags/internal/convert"
	"github.com/reeflective/flags/internal/tag"
)

// ErrRequired signals an argument field has not been
// given its minimum amount of positional words to use.
var ErrRequired = errors.New("required argument")

// WordConsumer is a function that has access to the array of positional slots,
// giving a few functions to manipulate the list of words we want to parse.
// As well, the current positional argument is a parameter, which is the only
// positional slot we can access within the function.
type WordConsumer func(args *Args, current *Arg, dash int) error

// WithWordConsumer allows to set a custom function to loop over
// the command words for a given positional slot. See WordConsumer.
func WithWordConsumer(args *Args, consumer WordConsumer) *Args {
	args.consumer = consumer

	return args
}

// Arg is a type used to store information and value references to
// a struct field we use as positional arg. This type is passed in
// many places, so that we can parse/convert and make informed
// decisions on how to handle those tasks.
type Arg struct {
	Index     int           // The position in the struct (n'th struct field used as a slot)
	Name      string        // name of the argument, either tag name or struct field
	Minimum   int           // minimum number of arguments we want.
	Maximum   int           // Maximum number of args we want (-1: infinite)
	StartMin  int           // Index of first positional word for which we are used
	StartMax  int           // if previous positional slots are full, this replaces startAt
	Tag       tag.MultiTag  // struct tag
	Value     reflect.Value // A reference to the field value itself
	Validator func(val string) error
}

// Args contains an entire list of positional argument "slots" (struct fields)
// along with everything needed to parse a list of words onto them, taking into
// account all of their requirements and constraints, and throwing an error with
// proper formatting and structure when one of these requirements are not satisfied.
type Args struct {
	// A list of positional struct fields
	slots []*Arg

	// Requirements
	totalMin    int  // Total count of required arguments
	totalMax    int  // the maximum number of required arguments
	allRequired bool // Are all positional slots required ?
	noTags      bool // Did we find at least one tag on a positional field ?

	// Internal word management
	words       []string // The list of arguments remaining to be parsed into their fields
	done        int      // A pointer that is being shared by all positional argument handlers
	parsed      int      // A counter used only by a single positional field
	needed      int      // A global value set when we know the total number of arguments
	offsetRange int      // Used to adjust the number of words still needed in relation to an argument min/max
	dash        int

	// Users can pass a custom handler to loop over the words
	// This consumer is called for each positional slot, either
	// sequentially (normal parsing) or concurrently (useful for completions)
	consumer WordConsumer
}

// Parse acceps a list of command-line words to be ALL parsed as positional
// arguments of a command. This function will parse each word into its proper
// positional struct field (following quantity constraints/requirements), and
// will return the list of words that have not been parsed into a field, along
// with an error if one/more positionals has failed to satisfy their requirements.
func (args *Args) Parse(words []string, dash int) (retargs []string, err error) {
	args.setWords(words) // Ensures initializing the counters
	args.dash = dash

	// Always set the return arguments when exiting.
	// This is used by command callers needing them
	// as lambda parameters to the implementation.
	defer func() { retargs = args.words }()

	// We consume all fields until one either errors out,
	// or all are fullfiled up to their minimum requirements.
	for _, arg := range args.slots {
		args.setNext(arg)

		// The positional slot consumes words as it needs, and only
		// returns an error when it cannot fulfill its requirements.
		err := args.consumeWords(args, arg, dash)

		// Either the positional argument has not had enough words
		if errors.Is(err, ErrRequired) {
			return retargs, args.positionalRequiredErr(*arg)
		}

		// Or we have failed to parse the word onto the struct field
		// value, most probably because it's the wrong type.
		if errors.Is(err, convert.ErrConvertion) {
			return retargs, err
		}

		// Or return the error as is
		if err != nil {
			return retargs, err
		}
	}

	// Finally, if we have some return arguments, we verify that
	// that the last positional was not a list with a maximum specified:
	// This is to keep retrocompatibility with go-flags. Should be moved.
	return retargs, args.checkRequirementsFinal()
}

// ParseConcurrent is to parse all positional arguments onto their slots
// without them to wait for the previous slot to be done parsing its words.
// This is used by things like completion engines, which just need to know
// which positional argument to complete, and optionally to ensure that the
// previously completed ones do not raise conversion errors.
func (args *Args) ParseConcurrent(words []string) {
	workers := &sync.WaitGroup{}

	for _, arg := range args.slots {
		// Make a copy of our positionals, so that each arg slot can
		// work on the same word list while doing different things.
		argsC := args.copyArgs()
		argsC.words = words

		workers.Add(1)

		go func(arg *Arg) {
			defer workers.Done()

			// If we don't have enough words for even
			// considering this positional to be completed.
			if len(argsC.words) < arg.StartMin {
				return
			}

			// Else, run the consumer function, to loop over words.
			if err := argsC.consumer(argsC, arg, 0); err != nil {
				// TODO change this
				return
			}
		}(arg)
	}

	workers.Wait()
}

// Positionals returns the list of "slots" that have been
// created when parsing a struct of positionals.
func (args *Args) Positionals() []*Arg {
	return args.slots
}

// copyArgs is used to make several instances of our args
// to work on the same list of command words (copies of it).
func (args *Args) copyArgs() *Args {
	return &Args{
		slots:       args.slots,
		totalMin:    args.totalMin,
		totalMax:    args.totalMax,
		allRequired: args.allRequired,
		needed:      args.totalMin,
		noTags:      args.noTags,
		done:        0,
		parsed:      0,
		consumer:    args.consumer,
	}
}

// consumePositionals parses one or more words from the current list of positionals into
// their struct fields, and returns once its own requirements are satisfied and/or the
// next positional arguments require words to be passed along.
func (args *Args) consumeWords(self *Args, arg *Arg, dash int) error {
	// As long as we've got a word, and nothing told us to quit.
	for !self.Empty() {
		// If we have reached the maximum number of args we accept.
		if (self.parsed == arg.Maximum) && arg.Maximum != -1 {
			return nil
		}

		// If we have the minimum required, but there are
		// "just enough" (we assume it at least) words for
		// the next arguments, leave them the words.
		if self.parsed >= arg.Minimum && self.allRemainingRequired() {
			return nil
		}

		// Skip all remaining args (potentially with an error)
		// if we reached the positional double dash.
		if dash != -1 && self.done == dash {
			break
		}

		// Else if we have not reached our maximum allowed number
		// of arguments, we are cleared to consume one.
		next := args.Pop()

		// If the positional slot has a validator function,
		// run it before trying to convert the value.
		if arg.Validator != nil {
			if err := arg.Validator(next); err != nil {
				return err
			}
		}
		// Parse the string value onto its native type, returning any errors.
		// We also break this loop immediately if we are not parsing onto a list.
		if err := convert.Value(next, arg.Value, arg.Tag); err != nil {
			return fmt.Errorf("%w: %s", convert.ErrConvertion, err.Error())
		} else if arg.Value.Type().Kind() != reflect.Slice {
			return nil
		}
	}

	// If we are still lacking some required words,
	// but we have exhausted the available ones.
	if self.parsed < arg.Minimum {
		return ErrRequired
	}

	// Or we consumed all the arguments we wanted, without
	// error, so either exit because we are the last, or go
	// with the next argument handler we bound.
	return nil
}

//
// Error check/build/format code ----------------------------------------------------------------------
//

// checkPositionals is only called if ALL positional slots have successfully worked,
// and makes some final checks about these positionals. Some checks are here for retrocompat.
func (args *Args) checkRequirementsFinal() error {
	slots := args.slots
	if len(slots) == 0 {
		return nil
	}

	current := slots[len(slots)-1]
	isSlice := current.Value.Type().Kind() == reflect.Slice || current.Value.Type().Kind() == reflect.Map

	// This is for retrocompatibility with jessevdk/go-flags, so that
	// any remaining slot being a list with a specified maximum value
	// cannot accept more than that, and will error out instead of
	// silently passing the excess args onto the Execute() parameters.
	if isSlice && current.Value.Len() == current.Maximum && len(args.words) > 0 {
		overweight := argHasTooMany(*current, len(args.words))
		msgErr := fmt.Sprintf("%s was not provided", overweight)

		return fmt.Errorf("%w: %s", ErrRequired, msgErr)
	}

	return nil
}

// positionalErrorHandler makes a handler to be used in our argument handlers,
// when they fail, to compute a precise error message on argument requirements.
func (args *Args) positionalRequiredErr(arg Arg) error {
	if names := args.getRequiredNames(arg); len(names) > 0 {
		var msg string

		if len(names) == 1 {
			msg = fmt.Sprintf("%s was not provided", names[0])
		} else {
			msg = fmt.Sprintf("%s and %s were not provided",
				strings.Join(names[:len(names)-1], ", "), names[len(names)-1])
		}

		return fmt.Errorf("%w: %s", ErrRequired, msg)
	}

	return nil
}

// getRequiredNames is used by an argument handler to build the correct list of arguments we need.
func (args *Args) getRequiredNames(current Arg) (names []string) {
	// For each of the EXISTING positional argument fields
	for index, arg := range args.slots {
		// Ignore all positional arguments that have not
		// thrown an error: they have what they need.
		if index < current.Index {
			continue
		}

		// Non required positional won't appear in the message
		if !isRequired(arg) {
			continue
		}

		// If the positional is a single slot, we need its name
		if arg.Value.Type().Kind() != reflect.Slice {
			names = append(names, "`"+arg.Name+"`")

			continue
		}

		// If we have less words to parse than
		// the minimum required by this argument.
		if arg.Value.Len() < arg.Minimum {
			names = append(names, argHasNotEnough(*arg))

			continue
		}
	}

	return names
}

// makes a correct sentence when we don't have enough args.
func argHasNotEnough(arg Arg) string {
	var arguments string

	if arg.Minimum > 1 {
		arguments = "arguments, but got only " + fmt.Sprintf("%d", arg.Value.Len())
	} else {
		arguments = "argument"
	}

	argRequired := "`" + arg.Name + " (at least " + fmt.Sprintf("%d",
		arg.Minimum) + " " + arguments + ")`"

	return argRequired
}

// makes a correct sentence when we have too much args.
func argHasTooMany(arg Arg, added int) string {
	// The argument might be explicitly disabled...
	if arg.Maximum == 0 {
		return "`" + arg.Name + " (zero arguments)`"
	}

	// Or just build the list accordingly.
	var parsed string

	if arg.Maximum > 1 {
		parsed = "arguments, but got " + fmt.Sprintf("%d", arg.Value.Len()+added)
	} else {
		parsed = "argument"
	}

	hasTooMany := "`" + arg.Name + " (at most " + fmt.Sprintf("%d", arg.Maximum) + " " + parsed + ")`"

	return hasTooMany
}

func isRequired(p *Arg) bool {
	return (p.Value.Type().Kind() != reflect.Slice && (p.Minimum > 0)) || // Both must be true
		p.Minimum != -1 || p.Maximum != -1 // And either of these
}

//
// Code related to command word manipulation and requirements/counters management. --------------------
//

// setWords uses a list of command words
// to be parsed as positional arguments,.
func (args *Args) setWords(words []string) {
	args.words = words
	args.parsed = 0
}

// setNext (re)sets the number of words parsed by
// a single positional slot to 0, so that the next
// positional using the words can set its own values.
func (args *Args) setNext(arg *Arg) {
	args.parsed = 0
	args.offsetRange = arg.Minimum
}

func (args *Args) Empty() bool {
	return len(args.words) == 0
}

func (args *Args) allRemainingRequired() bool {
	if args.dash != -1 {
		return len(args.words[:args.dash]) <= args.needed
	}

	return len(args.words) <= args.needed
}

// Pop returns the first word in the words
// list, and remove this word from the list.
// Also updates the various counters in list.
func (args *Args) Pop() string {
	if args.Empty() {
		return ""
	}

	// Pop the last word
	arg := args.words[0]
	args.words = args.words[1:]

	// The current positional slot
	// just parsed a word.
	args.parsed++

	// the global counter is increased
	args.done++

	if args.dash != -1 {
		args.dash--
	}

	// We only update the number of words
	// we still need when this positional
	// slot is below its minimum requirement
	if args.offsetRange > 0 {
		args.needed--
		args.offsetRange-- // So that we stop once we have the minimum
	}

	return arg
}

func (args *Args) peek() string {
	if args.Empty() {
		return ""
	}

	return args.words[0]
}
