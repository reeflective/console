// Package flags (github.com/reeflective/flags/gen/flags) provides the entrypoints
// to command trees and flags generation for arbitrary structures. The following
// is an overview of the possible tags to use for each component, as well as some
// details and advices for generation.
//
// 1 - Generation advices and details ****************************************************
//
// A) Parsing options
// General parsing options ([]flags.OptFunc) are located in the root directory of
// the module. Since both packages are named 'flags', it is advised, when using
// parsing options from the root one, to use the following import statement:
//
// import (
//      "github.com/reeflective/flags"
//      gen "github.com/reeflective/gen/flags"
// )
//
// B) Retrocompatiblity
// For library users coming from github.com/octago/sflags:
// - When parsing structs with no tags (in which case every field is a flag),
// the option `flags.ParseAll()` should be passed to the `Generate()` call.
//
//
// 2 - Valid tags ************************************************************************
//
// A) Commands -------------------------------------------------------------------
// command:              When specified on a struct field, makes the struct
//                       field a (sub)command with the given name (optional).
//                       Note that a struct marked as a command does not mandatorily
//                       have to implement the `flags.Commander` interface.
// subcommands-optional: When specified on a command struct field, makes
//                       any subcommands of that command optional (optional)
// alias:                When specified on a command struct field, adds the
//                       specified name as an alias for the command. Can be
//                       be specified multiple times to add more than one
//                       alias (optional)
// group:                If the group name is not nil, this command will be
//                       grouped under this heading in the help usage.
//
//
// B) Flags ----------------------------------------------------------------------
//
// a) github.com/jessevdk/go-flags tag specifications (some have been removed):
//
// flag:             Short and/or long names for the flag, space-separated.
//                   (ex: `flag:"-v --verbose`).
// short:            The short name of the option (single character)
// long:             The long name of the option
// required:         If non empty, makes the option required to appear on the command
//                   line. If a required option is not present, the parser will
//                   return ErrRequired (optional)
// description:      The description of the option (optional)
// desc:             Same as 'description'
// long-description: The long description of the option. Currently only
//                   displayed in generated man pages (optional)
// no-flag:          If non-empty, this field is ignored as an option (optional)
// optional:         If non-empty, makes the argument of the option optional. When an
//                   argument is optional it can only be specified using
//                   --option=argument (optional)
// optional-value:   The value of an optional option when the option occurs
//                   without an argument. This tag can be specified multiple
//                   times in the case of maps or slices (optional)
// default:          The default value of an option. This tag can be specified
//                   multiple times in the case of slices or maps (optional)
// default-mask:     When specified, this value will be displayed in the help
//                   instead of the actual default value. This is useful
//                   mostly for hiding otherwise sensitive information from
//                   showing up in the help. If default-mask takes the special
//                   value "-", then no default value will be shown at all
//                   (optional)
// env:              The default value of the option is overridden from the
//                   specified environment variable, if one has been defined.
//                   (optional)
// env-delim:        The 'env' default value from environment is split into
//                   multiple values with the given delimiter string, use with
//                   slices and maps (optional)
// choice:           Limits the values for an option to a set of values.
//                   You can either specify multiple values in a single tag
//                   if they are space-separated, and/or with multiple tags.
//                   (e.g. `long:"animal" choice:"cat bird" choice:"dog"`)
// hidden:           If non-empty, the option is not visible in the help or man page.
//
// b) github.com/octago/sflags tag specification:
//
// `flag:"-"`           Field is ignored by this package.
// `flag:"myName"`      Field appears in flags as "myName".
// `flag:"~myName"`     If this field is from nested struct, prefix from parent struct will be ingored.
// `flag:"myName a"`    You can set short name for flags by providing it's value after a space.
// `flag:",hidden"`     This field will be removed from generated help text.
// `flag:",deprecated"` This field will be marked as deprecated in generated help text
//
//
// C) Positionals ----------------------------------------------------------------
//
// The following tags can/must be specified on the struct containing positional args:
//
// positional-args:     When specified on a field with a struct type,
//                      uses the fields of that struct to parse remaining
//                      positional command line arguments into (in order
//                      of the fields).
//                      Positional arguments are optional by default,
//                      unless the "required" tag is specified together
//                      with the "positional-args" tag.
//
// required:            If non empty, will make ALL of the fields in the positional
//                      struct to be required. However, each field can still specify
//                      its own quantity requirements/range if its a slice/map.
//                      If you can, please check at the online documentation for various
//                      examples of positional declaration and their behavior.
//
// The following tags can be specified on each individual field of a positional struct:
//
// positional-arg-name: used on a field in a positional argument struct; name
//                      of the positional argument placeholder to be shown in
//                      the help (optional)
//
// description:         The description of the argument (optional)
//
// required:            The "required" tag can be set on each argument field.
//                      If it is set on a slice of map field, then its value
//                      determines the minimum amount of rest arguments that
//                      needs to be provided (e.g. `required:"2"`).
//                      You can also specify a range (e.g. `required:"1-3"`).
//                      When several fields are slices/or arrays, they may still
//                      each declare ranges (even if overlapping). When that is
//                      the case, the slices are filled from first to last.
//                      Ex:
//                      struct {
//                             List  []string `required:"1-2"`
//                             Other []string `required:"1-2"`
//                             Final string  `required:"yes"`
//                      }
//                      If given ["one", "two", "three", "four"], final will have
//                      "four", Other will have ["three"], and List will have ["one", "two"].
//
//                      When the last field is a slice/map with no maximum, then it
//                      will hold all excess arguments. On the contrary, and as general
//                      rule, all arguments not fitting into the struct fields will be
//                      given as args to the command's `Execute(args []string)` function.
//
//                      Also, and when a double dash is passed in the arguments,
//                      all args after the dash will not be parsed into struct fields.
//                      If those fields' requirements are not satisfied, however, they
//                      will throw an error.
//                      Various examples of positional arguments declaration can be found
//                      on the online documentation.
//
//
// D) Groups (of flags or commands) ----------------------------------------------
//
// group:         When specified on a struct field, makes the struct
//                field a separate flags group with the given name (optional).
// commands:      When specified on a struct field containing commands,
//                the value of the tag is used as a name to group commands
//                together in the help usage.
// namespace:     When specified on a group struct field, the namespace
//                gets prepended to every option's long name and
//                subgroup's namespace of this group, separated by
//                the parser's namespace delimiter (optional) (flags only)
// env-namespace: When specified on a group struct field, the env-namespace
//                gets prepended to every option's env key and
//                subgroup's env-namespace of this group, separated by
//                the parser's env-namespace delimiter (optional) (flags only)
// persistent:    If non-empty, all flags belonging to this group will be
//                persistent across subcommands.
//
//
// D) Completions (flags or positionals) -------------------------------------------
//
// a) Tagged completions
//
// complete: This is the only tag required to provide completions for a given positional
//           argument or flag struct field. The following directives and formattings are
//           are accepted (all directives can also be written as lowercase):
//
// `FilterExt` only complete files that are part of the given extensions.
// ex: `complete:"FilterExt,json,go,yaml"` will only propose JSON/Go/YAML files.
//
// `FilterDirs` only complete files within a given set of directories.
// ex: `complete:"FilterDirs,/home/user,/usr"` will complete from those root directories.
//
// `Files` completes all files found in the current filesystem context.
// ex: `complete:"Files"`
//
// `Dirs` completes all directories in the current filesystem context.
// ex: `complete:"dirs"` (lowercase is still valid)
//
// b) Additional completions
//
// Completers can also be implement by positional/flags field types, with:
// `func (m *myType) Complete(ctx carapace.Context) carapace.ActionCallback`
//
// The `ctx` argument can be altogether ignored for most completions: it
// provides low-level access to the completion context for those who need,
// but the engine itself is already very performant at handling prefixing/formatting.
// Please check the carapace documentation for writing completers.
//
// Also, note that the flags library is quite efficient at identifying the kind of
// the positional/flag field (whether it's a map/slice or not), and if it detects
// a []YourType, where `YourType` individually implements the completer, flags will
// wrap it into a compliant list completer. As well, if it detects the list/map itself
// declares the completer, it will use it as is.
//
// Please check https://rsteube.github.io/carapace/carapace.html for library usage information.
//
//
// E) Validations ------------------------------------------------------------------
//
// All positionals and flags struct fields can also declare validations compliant with
// "github.com/go-playground/validator/v10" tag specifications, and provided that the
// following parsing option is given to the `Generate()` call:
//
// flags.Validator(validator.New())
//
// The tag should be named `validate`. Examples:
// EmailField `flag:"-e --email" validate:"ipv4"`   // A command flag field
// Interfaces `required:"1" validate:"email"`       // A positional field
//
// Check the documentation for adding other custom validations directly through the
// go-validator engine.
//
package flags
