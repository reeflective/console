
Gonsole - Integrated Console Application library
=========

This package rests on a [readline](https://github.com/maxlandon/readline) console library, (giving advanced completion, hint, input and history system), 
and the [go-flags](https://github.com/jessievdk/go-flags) commands library. Also added, a bit of optional boilerplate for better user experience.

The purpose of this library is to offer a complete off-the-shelf console application, with some key aspects: 
- Better overall features than what is currently seen in such projects, including those not written in Go.
- A really simple but very powerful way of transforming code (structs and anything they might embed), into commands.
- An equally simple way to provide completions for any command/subcommand, any arguments of them, or any option arguments.


----
## Features Summary

The list of features supported or provided by this library can fall into 2 different categories:
the shell/console interface part, and the commands/parsing logic part.  Some of the features below
are simply extrated from my [readline](https://github.com/maxlandon/readline) library (everything below **Shell Details**).


#### Menus
- Declare different "menus" (`Context` in the code), to which you can bind commands, prompt and shell settings.
- Commands and settings bound to a menu are not reachable from another menu, unless you bound the commands to it as well.
- Users can switch between menus (programmatically, so consumers of this library should call `console.SwitchContext("name")`).

#### Commands
- The library is fundamentally a wrapper around the [go-flags](https://github.com/jessievdk/go-flags) commands/options, etc.
- This go-flag library allows you to create commands out of structs (however populated), and gonsole asks you to pass these structs to it.
- So if you get how to declare a go-flags compatible command, you know how to declare commands that work with this library.
- This works for any level of command nesting. 
- Also allows you to declare as many option groups you want, all of that will work.
- All commands have methods for adding completions either to their arguments, or to their options.

#### Others
- You can pass special completers that will be triggered if the rune (like `$` or `@`) is detected, anywhere in the line. These variables are expanded at command execution time, and work in completions as well.
- You can export the configuration for your application, its menus, and add some custom subcommands to the root one, for specialized actions over it.
- Also, an optional `help` command can be bound to the console, in additional to default `-h`/`--help` flags for every command.
- History sources can also be bound per menu/menu.

#### Shell details
- Vim / Emacs input and editing modes.
- Vim modes (Insert, Normal, Replace, Delete) with visual prompt Vim status indicator
- line editing using `$EDITOR` (`vi` in the example - enabled by pressing `[ESC]` followed by `[v]`)
- Vim registers (one default, 10 numbered, and 26 lettered) and Vim iterations

#### Completion engine
- Rather easy declaration of completion generators, which some level of customization.
- 3 types of completion categories (`Grid`, `List` and `Map`)
- In `List` completion groups, ability to have alternative candidates (used for displaying `--long` and `-s` (short) options, with descriptions)
- Completions working anywhere in the input line (your cursor can be anywhere)
- Completions are searchable with *Ctrl-F*, 

#### Prompt system & Colors
- 1-line and 2-line prompts, both being customizable.
- Function for refreshing the prompt, with optional behavior settings.

#### Hints & Syntax highlighting
- A hint line can be printed below the input line, with any type of information. See utilities for a default one.
- The Hint system is now refreshed depending on the cursor position as well, like completions.
- A syntax highlighting system. 

#### Command history 
- Ability to have 2 different history sources (I used this for clients connected to a server, used by a single user).
- History is searchable like completions.
- Default history is an in-memory list.
- Quick history navigation with *Up*/*Down* arrow keys in Emacs mode, and *j*/*k* keys in Vim mode.


## Simple Usage

The library is made to work with sane but powerful defaults. Paste the following and run it,
and take a look around to get a feel, without touching anything. Default editing mode is Vim.

```go

func main() {

	// Instantiate a new console, with a single, default menu.
	// All defaults are set, and nothing is needed to make it work
	console := gonsole.NewConsole()

	// By default the shell as created a single menu and
	// made it current, so you can access it and set it up.
	menu := console.CurrentMenu()

	// Set the prompt (config, for usability purposes). Each menu has its own.
	// See the documentation for more prompt setup possibilities.
	prompt := menu.PromptConfig()
	prompt.Left = "application-name"
	prompt.Multiline = false

	// Add a default help command, that can be used with any command, however nested:
	// 'help <command> <subcommand> <subcommand'
	// The console creates it and attaches it to all existing contexts.
	// "core" is the name of the group in which we will put this command.
	console.AddHelpCommand("core")

	// Add a configuration command if you want your users to be able
	// to modify it on the fly, export it as files or as JSON.
	// Please see the documentation and/or use this example to
	// see what can be done with this.
	console.AddConfigCommand("config", "core")

	// Everything is ready for a tour.
	// Run the console and take a look around.
	console.Run()
}

```

If you're still here, at least you want to declare and bind commands. Just as everything else possible with
this library, it is explained in the [Wiki](https://github.com/maxlandon/gonsole/wiki), although with more 
pictures than text (I like pictures), because the code is heavily documented (I don't like to repeat myself).


## Status & Support 

#### Support:
- Support for any issue opened.
- Answering any questions related.
- Being available for any blame you'd like to make for my humble but passioned work. I don't mind, I need to go up.

#### TO DO:
- [ ] Recursive option completion (`-sP`, `-oL`, etc)
- [ ] `config load` command
- [ ] Analyse args and pull out from comps if in line
- [ ] Add color for strings in input line (this will need a good part of murex parser code) 
- [ ] Add token parsing code from murex (this must be well thought out, like the quotes stuff, because it must also not interfere with other commands in special menus, the command parsing code, etc...)


## File Contents & Packages


## License
