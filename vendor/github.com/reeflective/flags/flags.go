// Package flags is the root package of the `github.com/reeflective/flags` library.
//
// If you are searching for the list of valid tags to use on structs for specifying
// commands/flags specs, check https://github.com/reeflective/flags/gen/flags/flags.go.
//
// 1) Importing the various packages -----------------------------------------------------
//
// This file gives a list of the various global parsing options that can be passed
// to the `Generate()` entrypoint function in the `gen/flags` package. Below is an
// example of how you want to import this package and use its options:
//
// package main
//
// import (
//    "github.com/reeflective/flags/example/commands"
//
//    "github.com/reeflective/flags"
//    genflags "github.com/reeflective/flags/gen/flags"
//
//    "github.com/reeflective/flags/validator"
//    "github.com/reeflective/flags/gen/completions"
// )
//
// func main() {
//     var opts []flags.OptFunc
//
//     opts = append(opts, flags.Validator(validator.New()))
//
//     rootData := &commands.Root{}
//     rootCmd := genflags.Generate(rootData, opts...)
//
//     comps, _ := completions.Generate(rootCmd, rootData, nil)
// }
//
//
// 2) Global parsing options (base) ------------------------------------------------------
//
// Most of the options below are inherited from github.com/octago/sflags, with some added.
//
// DescTag sets custom description tag. It is "desc" by default.
// func DescTag(val string)
//
// FlagTag sets custom flag tag. It is "flag" be default.
// func FlagTag(val string)
//
// Prefix sets prefix that will be applied for all flags (if they are not marked as ~).
// func Prefix(val string)
//
// EnvPrefix sets prefix that will be applied for all environment variables (if they are not marked as ~).
// func EnvPrefix(val string)
//
// FlagDivider sets custom divider for flags. It is dash by default. e.g. "flag-name".
// func FlagDivider(val string)
//
// EnvDivider sets custom divider for environment variables.
// It is underscore by default. e.g. "ENV_NAME".
// func EnvDivider(val string)
//
// Flatten set flatten option.
// Set to false if you don't want anonymous structure fields to be flatten.
// func Flatten(val bool)
//
// ParseAll orders the parser to generate a flag for all struct fields, even if there isn't a struct
// tag attached to them. This is because by default the library does not considers untagged field anymore.
// func ParseAll()
//
//
// 3) Special parsing options/functions---------------------------------------------------
//
// ValidateFunc describes a validation func, that takes string val for flag from command line,
// field that's associated with this flag in structure `data`. Also works for positional arguments.
// Should return error if validation fails.
//
// type ValidateFunc func(val string, field reflect.StructField, data interface{}) error
//
//
// Validator sets validator function for flags.
// Check existing validators in flags/validator and flags/validator/govalidator packages.
//
// func Validator(val ValidateFunc)
//
//
// FlagFunc is a generic function that can be applied to each
// value that will end up being a flags *Flag, so that users
// can perform more arbitrary operations on each, such as checking
// for completer implementations, bind to viper configurations, etc.
//
// type FlagFunc func(flag string, tag tag.MultiTag, val reflect.Value) error
//
//
// FlagHandler sets the handler function for flags, in order to perform arbitrary
// operations on the value of the flag identified by the <flag> name parameter of FlagFunc.
//
// func FlagHandler(val FlagFunc)
//
package flags
