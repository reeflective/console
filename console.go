package console

import (
	"sync"

	"github.com/reeflective/readline"
)

// Console is an integrated console application instance.
type Console struct {
	// Application ------------------------------------------------------------------

	// shell - The underlying shell provides the core readline functionality,
	// including but not limited to: inputs, completions, hints, history.
	shell *readline.Shell

	// Different menus with different command trees, prompt engines, etc.
	menus menus

	// Execution --------------------------------------------------------------------

	// PreReadlineHooks - All the functions in this list will be executed,
	// in their respective orders, before the console starts reading
	// any user input (ie, before redrawing the prompt).
	PreReadlineHooks []func()

	// PreCmdRunLineHooks - Same as PreCmdRunHooks, but will have an effect on the
	// input line being ultimately provided to the command parser. This might
	// be used by people who want to apply supplemental, specific processing
	// on the command input line.
	PreCmdRunLineHooks []func(raw []string) (args []string, err error)

	// PreCmdRunHooks - Once the user has entered a command, but before executing
	// the target command, the console will execute every function in this list.
	// These hooks are distinct from the cobra.PreRun() or OnInitialize hooks,
	// and might be used in combination with them.
	PreCmdRunHooks []func()

	// PostCmdRunHooks are run after the target cobra command has been executed.
	// These hooks are distinct from the cobra.PreRun() or OnFinalize hooks,
	// and might be used in combination with them.
	PostCmdRunHooks []func()

	// True if the console is currently running a command. This is used by
	// the various asynchronous log/message functions, which need to adapt their
	// behavior (do we reprint the prompt, where, etc) based on this.
	isExecuting bool

	// concurrency management.
	mutex *sync.RWMutex

	// Other ------------------------------------------------------------------------

	// A list of tags by which commands may have been registered, and which
	// can be set to true in order to hide all of the tagged commands.
	filters []string
}

// New - Instantiates a new console application, with sane but powerful defaults.
// This instance can then be passed around and used to bind commands, setup additional
// things, print asynchronous messages, or modify various operating parameters on the fly.
func New() *Console {
	console := &Console{
		shell: readline.NewShell(),
		menus: make(menus),
		mutex: &sync.RWMutex{},
	}

	// Make a default menu and make it current.
	// Each menu is created with a default prompt engine.
	defaultMenu := console.NewMenu("")
	defaultMenu.active = true

	// Set the history for this menu
	for _, name := range defaultMenu.historyNames {
		console.shell.AddHistory(name, defaultMenu.histories[name])
	}

	// Command completion, syntax highlighting, multiline callbacks, etc.
	console.shell.AcceptMultiline = console.acceptMultiline
	console.shell.Completer = console.complete
	console.shell.SyntaxHighlighter = console.highlightSyntax

	return console
}

// Shell returns the console readline shell instance, so that the user can
// further configure it or use some of its API for lower-level stuff.
func (c *Console) Shell() *readline.Shell {
	return c.shell
}

func (c *Console) reloadConfig() {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	menu := c.menus.current()
	menu.prompt.bind(c.shell)
}

// SystemEditor - This function is a renamed-reexport of the underlying readline.StartEditorWithBuffer
// function, which enables you to conveniently edit files/buffers from within the console application.
// Naturally, the function will block until the editor is exited, and the updated buffer is returned.
// The filename parameter can be used to pass a specific filename.ext pattern, which might be useful
// if the editor has builtin filetype plugin functionality.
func (c *Console) SystemEditor(buffer []byte, filename string) ([]byte, error) {
	// runeUpdated, err := c.shell.StartEditorWithBuffer([]rune(string(buffer)), filename)
	//
	// return []byte(string(runeUpdated)), err
	return []byte{}, nil
}
