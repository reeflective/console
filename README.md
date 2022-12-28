
<div align="center">
  <a href="https://github.com/reeflective/console">
    <img alt="" src="" width="600">
  </a>
  <br> <h1> Console </h1>

  <p>  Application library for Cobra commands  </p>
</div>


<!-- Badges -->
<p align="center">
  <a href="https://github.com/reeflective/console/actions/workflows/go.yml">
    <img src="https://github.com/reeflective/console/actions/workflows/go.yml/badge.svg?branch=main"
      alt="Github Actions (workflows)" />
  </a>

  <a href="https://github.com/reeflective/console">
    <img src="https://img.shields.io/github/go-mod/go-version/reeflective/console.svg"
      alt="Go module version" />
  </a>

  <a href="https://godoc.org/reeflective/go/console">
    <img src="https://img.shields.io/badge/godoc-reference-blue.svg"
      alt="GoDoc reference" />
  </a>

  <a href="https://goreportcard.com/report/github.com/reeflective/console">
    <img src="https://goreportcard.com/badge/github.com/reeflective/console"
      alt="Go Report Card" />
  </a>

  <a href="https://codecov.io/gh/reeflective/console">
    <img src="https://codecov.io/gh/reeflective/console/branch/master/graph/badge.svg"
      alt="codecov" />
  </a>

  <a href="https://opensource.org/licenses/BSD-3-Clause">
    <img src="https://img.shields.io/badge/License-BSD_3--Clause-blue.svg"
      alt="License: BSD-3" />
  </a>
</p>

Console is an all-in-one console application library built on top of a [readline](https://github.com/reeflective/readline) shell and using [Cobra](https://github.com/spf13/cobra) commands. 
It aims so as to provide users with a modern interface at at minimal cost while allowing them to focus on developing 
their commands and application core: the console will then transparently interface with these commands, and provide
the various features below almost for free.

## Features

### Menus & Commands 
- Declare & use multiple menus with their own command tree, prompt engines and special handlers.
- Bind cobra commands to provide the core functionality (see documentation for binding usage).
- Virtually all cobra settings can be modified, set and used freely, like in normal CLI workflows.
- Ability to bind handlers to special interrupt errors (eg. CtrlC/CtrlD), per menu.

### Shell interface
- Shell interface is powered by a [readline](https://github.com/reeflective/readline) instance.
- All features of readline are supported in the console. It also allows the console to give:
- Configurable bind keymaps, with live reload and sane defaults, and system-wide configuration.
- Out-of-the-box, advanced completions for commands, flags, positional and flag arguments.
- Provided by readline and [carapace](https://github.com/rsteube/carapace): automatic usage & validation command/flags/args hints.
- Syntax highlighting for commands (might be extended in the future).

### Other features 
- Support for an arbitrary number of history sources, per menu.
- Support for [oh-my-posh](https://github.com/JanDeDobbeleer/oh-my-posh) prompts, per menu and with custom configuration files for each.

<!-- ![readme-main-gif](https://github.com/maxlandon/gonsole/blob/assets/readme-main.gif) -->

---- 
## Documentation Contents

### Developers
* [Menus](https://github.com/maxlandon/gonsole/wiki/Menus)
* [Configurations Overview](https://github.com/maxlandon/gonsole/wiki/Configurations-Overview)
* [Setting Prompts & Input Modes](https://github.com/maxlandon/gonsole/wiki/Prompts-&-Input-Modes)
* [Default commands](https://github.com/maxlandon/gonsole/wiki/Default-Commands)
* [Declaring commands](https://github.com/maxlandon/gonsole/wiki/Declaring-Commands)
* [Querying state from commands](https://github.com/maxlandon/gonsole/wiki/Querying-State-From-Commands)
* [Completions (writing and binding)](https://github.com/maxlandon/gonsole/wiki/Completions)
* [Additional Expansion completions](https://github.com/maxlandon/gonsole/wiki/Expansion-Completers)
* [History Sources Declaration](https://github.com/maxlandon/gonsole/wiki/History-Sources-Declaration)
* [Asynchronous Logs & Prompt Refresh](https://github.com/maxlandon/gonsole/wiki/Prompt-Refresh)

### Users
- [Vim Keys & Shortcuts](https://github.com/maxlandon/gonsole/wiki/Vim-Keys-&-Shortcuts)
- [History sources](https://github.com/maxlandon/gonsole/wiki/History-Sources)
- [Completions & Tab Search](https://github.com/maxlandon/gonsole/wiki/Completions-&-Tab-Search)
- [Help and config commands](https://github.com/maxlandon/gonsole/wiki/Help-&-Config-Commands)
