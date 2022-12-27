package flags

import (
	"errors"
	"fmt"
)

var (
	// ErrParse is a general error used to wrap more specific errors.
	ErrParse = errors.New("parse error")

	// ErrNotPointerToStruct indicates that a provided data container is not
	// a pointer to a struct. Only pointers to structs are valid data containers
	// for options.
	ErrNotPointerToStruct = errors.New("object must be a pointer to struct or interface")

	// ErrNotCommander is returned when an embedded struct is tagged as a command,
	// but does not implement even the most simple interface, Commander.
	ErrNotCommander = errors.New("provided data does not implement Commander")

	// ErrObjectIsNil is returned when the struct/object/pointer is nil.
	ErrObjectIsNil = errors.New("object cannot be nil")

	// ErrInvalidTag indicates an invalid tag or invalid use of an existing tag.
	ErrInvalidTag = errors.New("invalid tag")

	// ErrTag indicates an error while parsing flag tags.
	ErrTag = errors.New("tag error")

	// ErrShortNameTooLong indicates that a short flag name was specified,
	// longer than one character.
	ErrShortNameTooLong = errors.New("short names can only be 1 character long")

	// ErrFlagHandler indicates that the custom handler for a flag has failed.
	ErrFlagHandler = errors.New("custom handler for flag failed")

	// ErrNotValue indicates that a struct field type does not implement the
	// Value interface. This only happens when the said type is a user-defined one.
	ErrNotValue = errors.New("invalid field marked as flag")
)

// simple wrapper for errors.
func newError(err error, msg string) error {
	return fmt.Errorf("%s: %w", msg, err)
}

// ParserError represents the type of error.
// type ParserError uint

// ORDER IN WHICH THE ERROR CONSTANTS APPEAR MATTERS.
// const (
//         // ErrUnknown indicates a generic error.
//         ErrUnknown ParserError = iota
//
//         // ErrExpectedArgument indicates that an argument was expected.
//         ErrExpectedArgument
//
//         // ErrUnknownFlag indicates an unknown flag.
//         ErrUnknownFlag
//
//         // ErrUnknownGroup indicates an unknown group.
//         ErrUnknownGroup
//
//         // ErrMarshal indicates a marshalling error while converting values.
//         ErrMarshal
//
//         // ErrHelp indicates that the built-in help was shown (the error
//         // contains the help message).
//         ErrHelp
//
//         // ErrNoArgumentForBool indicates that an argument was given for a
//         // boolean flag (which don't not take any arguments).
//         ErrNoArgumentForBool
//
//         // ErrRequired indicates that a required flag was not provided.
//         ErrRequired
//
//         // ErrShortNameTooLong indicates that a short flag name was specified,
//         // longer than one character.
//         // ErrShortNameTooLong
//
//         // ErrDuplicatedFlag indicates that a short or long flag has been
//         // defined more than once.
//         ErrDuplicatedFlag
//
//         // ErrTag indicates an error while parsing flag tags.
//         // ErrTag
//
//         // ErrCommandRequired indicates that a command was required but not
//         // specified.
//         ErrCommandRequired
//
//         // ErrUnknownCommand indicates that an unknown command was specified.
//         ErrUnknownCommand
//
//         // ErrInvalidChoice indicates an invalid option value which only allows
//         // a certain number of choices.
//         ErrInvalidChoice
//
//         // ErrInvalidTag indicates an invalid tag or invalid use of an existing tag.
//         // ErrInvalidTag
// )

// func (e ParserError) String() string {
//         errs := [...]string{
//                 // Public
//                 "unknown",              // ErrUnknown
//                 "expected argument",    // ErrExpectedArgument
//                 "unknown flag",         // ErrUnknownFlag
//                 "unknown group",        // ErrUnknownGroup
//                 "marshal",              // ErrMarshal
//                 "help",                 // ErrHelp
//                 "no argument for bool", // ErrNoArgumentForBool
//                 "duplicated flag",      // ErrDuplicatedFlag
//                 // "tag",                  // ErrTag
//                 "command required",     // ErrCommandRequired
//                 "unknown command",      // ErrUnknownCommand
//                 "invalid choice",       // ErrInvalidChoice
//                 // "invalid tag",          // ErrInvalidTag
//         }
//         if len(errs) > int(e) {
//                 return "unrecognized error type"
//         }
//
//         return errs[e]
// }
//
