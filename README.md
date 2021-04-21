
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
the shell/console interface part, and the commands/parsing logic part. We start with the latter,
in order to highlight some of the key differences that this library provides, then we go on with
the shell *per se*, with most of the interface features that you will find in there. 

### 1 - Commands 

#### Menus
- The console allows to declare different "menus" (`Context` in the code), to which you can bind commands, prompt and shell settings.
- Any commands and or settings bound to a given menu are not reachable from another menu, unless you bound the commands to it as well.
- Users can switch between menus (programmatically, so consumers of this library should call `console.SwitchContext("name")`).
- By default, if you don't intend to use multiple menus, you don't have to bother with them, and the console will provide the necessary methods for you to declare commands.

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


### 2 - Shell (or readline) 

#### Input & Editing 
- Vim / Emacs input and editing modes.
- Optional, live-refresh Vim status.
- Vim modes (Insert, Normal, Replace, Delete) with visual prompt Vim status indicator
- line editing using `$EDITOR` (`vi` in the example - enabled by pressing `[ESC]` followed by `[v]`)
- Vim registers (one default, 10 numbered, and 26 lettered)
- Vim iterations
- Most default keybindings you might find in Emacs-like readline. Some are still missing though

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

The library is made to work with sane but powerful defaults. In order to have a running console instance, simply do:
```go

```


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
