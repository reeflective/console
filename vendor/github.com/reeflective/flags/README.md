
<div align="center">
  <a href="https://github.com/reeflective/flags">
    <img alt="" src="" width="600">
  </a>
  <br> <h1> Flags </h1>

  <p>  Generate cobra commands from structs </p>
  <p>  jessevdk/go-flags and octago/sflags compliant tags. </p>
  <p>  Enhanced with advanced related CLI functionality, at minimum cost. </p>
</div>


<!-- Badges -->
<p align="center">
  <a href="https://github.com/reeflective/flags/actions/workflows/go.yml">
    <img src="https://github.com/reeflective/flags/actions/workflows/go.yml/badge.svg?branch=main"
      alt="Github Actions (workflows)" />
  </a>

  <a href="https://github.com/reeflective/flags">
    <img src="https://img.shields.io/github/go-mod/go-version/reeflective/flags.svg"
      alt="Go module version" />
  </a>

  <a href="https://pkg.go.dev/github.com/reeflective/flags">
    <img src="https://img.shields.io/badge/godoc-reference-blue.svg"
      alt="GoDoc reference" />
  </a>

  <a href="https://goreportcard.com/report/github.com/reeflective/flags">
    <img src="https://goreportcard.com/badge/github.com/reeflective/flags"
      alt="Go Report Card" />
  </a>

  <a href="https://codecov.io/gh/reeflective/flags">
    <img src="https://codecov.io/gh/reeflective/flags/branch/main/graph/badge.svg"
      alt="codecov" />
  </a>

  <a href="https://opensource.org/licenses/BSD-3-Clause">
    <img src="https://img.shields.io/badge/License-BSD_3--Clause-blue.svg"
      alt="License: BSD-3" />
  </a>
</p>


## Summary

The flags library allows to declare cobra CLI commands, flags and positional arguments from structs and field tags.
It originally aimed to enhance [go-flags](https://github.com/jessevdk/go-flags), but ended up shifting its approach in order to leverage the widely 
used and battle-tested [cobra](https://github.com/spf13/cobra) CLI library. In addition, it provides other generators leveraging the [carapace](https://github.com/rsteube/carapace)
completion engine, thus allowing for very powerful yet simple completion and as-you-type usage generation for 
the commands, flags and positional arguments.

In short, the main purpose of this library is to let users focus on writing programs. It requires very little 
time and focus spent on declaring CLI interface specs (commands, flags, groups of flags/commands) and associated 
functionality (completions and validations), and then generates powerful and ready to use CLI programs.


## Features 

### Commands, flags & positionals 
- Easily declare commands, flags, and positional arguments through struct tags.
- Various ways to structure the command trees in groups (tagged, or encapsulated in structs).
- Almost entirely retrocompatible with [go-flags](https://github.com/jessevdk/go-flags), with a ported and enlarged test suite.
- Advanced and versatile positional arguments declaration, with automatic binding to `cobra.Args`.
- Large array of native types supported as flags or positional arguments.

### Related functionality
- Easily declare validations on command flags or positional arguments, with [go-validator](https://github.com/go-playground/validator) tags.
- Generate advanced completions with the [carapace](https://github.com/rsteube/carapace) completion engine in a single call.
- Implement completers on positional/flag types, or declare builtin completers via struct tags. 
- Generated completions include commands/flags groups, descriptions, usage strings.
- Live validation of command-line input with completers running flags' validations.
- All of these features, cross-platform and cross-shell, almost for free.


## Documentation

- A good way to introduce you to this library is to [install and use the example application binary](https://github.com/reeflective/flags/tree/main/example).
  This example application will give you a taste of the behavior and supported features.
- The generation package [flags](https://github.com/reeflective/flags/tree/main/gen/flags) has a [godoc file](https://github.com/reeflective/flags/tree/main/gen/flags/flags.go) with all the valid tags for each component 
  (commands/groups/flags/positionals), along with some notes and advices. This is so that you can
  quickly get access to those from your editor when writing commands and functionality.
- Another [godoc file](https://github.com/reeflective/flags/tree/main/flags.go) provides quick access to global parsing options (for global behavior, 
  validators, etc) located in the root package of this library. Both godoc files will be merged.
- Along with the above, the following is the table of contents of the [wiki documentation](https://github.com/reeflective/flags/wiki):

### Development
* [Introduction and principles](https://github.com/reeflective/flags/wiki/Introduction)
* [Declaring and using commands](https://github.com/reeflective/flags/wiki/Commands)
* [Positional arguments](https://github.com/reeflective/flags/wiki/Positionals)
* [Flags](https://github.com/reeflective/flags/wiki/Flags)
* [Completions](https://github.com/reeflective/flags/wiki/Completions)
* [Validations](https://github.com/reeflective/flags/wiki/Validations)
* [Side features](https://github.com/reeflective/flags/wiki/Side-Features)

### Coming from other libraries
* [Changes from octago/sflags](https://github.com/reeflective/flags/wiki/Sflags)
* [Changes from jessevdk/go-flags](https://github.com/reeflective/flags/wiki/Go-Flags)


## Status

This library is currently in a pre-release candidate state, for several reasons:
- It has not been widely tested, and some features/enhancements remain to be done.
- There might be bugs, or behavior inconsistencies that I might have missed.
- The codebase is not huge, but significant nonetheless. I aimed to write it 
  as structured and cleanly as possible.

Please open a PR or an issue if you wish to bring enhancements to it. For newer features, 
please consider if there is a large number of people who might benefit from it, or if it 
has a chance of impairing on future development. If everything is fine, please propose !
Other contributions, as well as bug fixes and reviews are also welcome.


## Credits

- This library is _heavily_ based on [octago/sflags](https://github.com/octago/sflags) code (it is actually forked from it since most of its code was needed).
  The flags generation is almost entirely his, and this library would not be as nearly as powerful without it. He should also
  be credited for 99% of this library's 99% coverage rate. It is also the inspiration for the trajectory this project has taken, 
  which originally would just enhance go-flags.
- The [go-flags](https://github.com/jessevdk/go-flags) is probably the most widely used reflection-based CLI library. While it will be hard to find a lot of 
  similarities with this project's codebase, the internal logic for scanning arbitrary structures draws almost all of its
  inspiration out of this project.
- The completion engine [carapace](https://github.com/rsteube/carapace), a fantastic library for providing cross-shell, multi-command CLI completion with hundreds 
  of different system completers. The flags project makes use of it for generation the completers for the command structure.
