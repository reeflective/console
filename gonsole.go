package gonsole

import (
	"github.com/jessevdk/go-flags"
	"github.com/maxlandon/readline"
)

// Console - An integrated console instance.
type Console struct {
	// Shell - The underlying Shell provides the core readline functionality,
	// including but not limited to: inputs, completions, hints, history.
	Shell *readline.Instance

	// Contexts - The various contexts hold a list of command instantiators
	// structured by groups. These groups are needed for completions and helps.
	contexts map[string]*Context
	current  *Context // The name of the current context

	// parser - Contains the whole aspect of command registering, parsing,
	// processing, and execution. There is only one parser at a time,
	// because it is recreated & repopulated at each console execution loop.
	parser *flags.Parser

	// PreLoopHooks - All the functions in this list will be executed,
	// in their respective orders, before the console starts reading
	// any user input (ie, before redrawing the prompt).
	PreLoopHooks []func()

	// PreRunHooks - Once the user has entered a command, but before executing it
	// with the application go-flags parser, the console will execute every func
	// in this list.
	PreRunHooks []func()

	// PreRunLineHooks - Same as PreRunHooks, but will have an effect on the
	// input line being ultimately provided to the command parser. This might
	// be used by people who want to apply supplemental, specific processing
	// on the command input line.
	PreRunLineHooks []func() (args []string, err error)

	// True if the console is currently running a command. This is used by
	// the various asynchronous log/message functions, which need to adapt their
	// behavior (do we reprint the prompt, where, etc) based on this.
	isExecuting bool

	// If true, leavs a newline between command line input and their output.
	LeaveNewline bool
}

// NewConsole - Instantiates a new console application, with sane but powerful defaults.
// This instance can then be passed around and used to bind commands, setup additional
// things, print asynchronous messages, or modify various operating parameters on the fly.
func NewConsole() (c *Console) {

	// Setup the readline instance, and input mode
	c.Shell = readline.NewInstance()
	c.Shell.Multiline = true
	c.Shell.ShowVimMode = true
	c.Shell.VimModeColorize = true

	// Setup the prompt (all contexts)
	c.Shell.MultilinePrompt = " > "

	// Setup completers, hints, etc
	// c.Shell.TabCompleter = completers.TabCompleter
	c.Shell.MaxTabCompleterRows = 50
	// c.Shell.HintText = completers.CommandCompleter
	// c.Shell.SyntaxHighlighter = completers.SyntaxHighlighter

	// Default context, "" (empty name)
	c.current = c.NewContext("")
	c.current.Prompt.Left = "gonsole"

	// Setup CtrlR history with an in-memory one by default
	c.current.SetHistoryCtrlR("client history (in-memory)", new(readline.ExampleHistory))

	// Create parser and set default options
	c.parser = flags.NewNamedParser("", flags.HelpFlag|flags.IgnoreUnknown)

	return
}

// HideCommands - Commands, in addition to their contexts, can be shown/hidden based
// on a filter string. For example, some commands applying to a Windows host might
// be scattered around different groups, but, having all the filter "windows".
// If "windows" is used as the argument here, all windows commands for the current
// context are subsquently hidden, until ShowCommands("windows") is called.
func (c *Console) HideCommands(filter string) {

}

// ShowCommands - Commands, in addition to their contexts, can be shown/hidden based
// on a filter string. For example, some commands applying to a Windows host might
// be scattered around different groups, but, having all the filter "windows".
// Use this function if you have previously called HideCommands("filter") and want
// these commands to be available back under their respective context.
func (c *Console) ShowCommands(filter string) {

}
