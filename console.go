package console

import (
	"regexp"
	"strings"
	"sync"

	"github.com/reeflective/readline"
)

// Console is an integrated console application instance.
type Console struct {
	// Application ------------------------------------------------------------------

	// shell - The underlying shell provides the core readline functionality,
	// including but not limited to: inputs, completions, hints, history.
	shell *readline.Instance

	// Contexts - The various menus hold a list of command instantiators
	// structured by groups. These groups are needed for completions and helps.
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
	PreCmdRunHooks []func()

	// True if the console is currently running a command. This is used by
	// the various asynchronous log/message functions, which need to adapt their
	// behavior (do we reprint the prompt, where, etc) based on this.
	isExecuting bool

	// concurrency management.
	mutex *sync.RWMutex

	// Other ------------------------------------------------------------------------

	// NOTE: Use via annotations on commands
	// A list of tags by which commands may have been registered, and which
	// can be set to true in order to hide all of the tagged commands.
	filters []string
}

// NewConsole - Instantiates a new console application, with sane but powerful defaults.
// This instance can then be passed around and used to bind commands, setup additional
// things, print asynchronous messages, or modify various operating parameters on the fly.
func NewConsole() *Console {
	c := &Console{
		shell: readline.NewInstance(),
		menus: make(menus),
		mutex: &sync.RWMutex{},
	}

	// Make a default menu and make it current.
	// Each menu is created with a default prompt engine.
	defaultMenu := c.NewMenu("")
	defaultMenu.active = true

	// Command completion, syntax highlighting, etc.
	c.shell.Completer = c.complete
	c.shell.SyntaxHighlighter = c.highlightSyntax

	return c
}

// Shell returns the console readline shell instance, so that the user can
// further configure it or use some of its API for lower-level stuff.
func (c *Console) Shell() *readline.Instance {
	return c.shell
}

func (c *Console) reloadConfig() {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	menu := c.menus.current()
	menu.prompt.bind(c.shell)
}

// sanitizeInput - Trims spaces and other unwished elements from the input line.
func (c *Console) sanitizeInput(line string) (sanitized []string, empty bool) {
	// Assume the input is not empty
	empty = false

	// Trim border spaces
	trimmed := strings.TrimSpace(line)
	if len(line) < 1 {
		empty = true
		return
	}

	// Parse arguments for quotes, and split according to these quotes first:
	// they might influence heavily on the go-flags argument parsing done further
	// Split all strings with '' and ""
	r := regexp.MustCompile(`[^\s"']+|"([^"]*)"|'([^']*)'`)
	unfiltered := r.FindAllString(trimmed, -1)

	var test []string
	for _, arg := range unfiltered {
		if strings.HasPrefix(arg, "'") && strings.HasSuffix(arg, "'") {
			trim := strings.TrimPrefix(arg, "'")
			trim = strings.TrimSuffix(trim, "'")
			test = append(test, trim)
			continue
		}
		test = append(test, arg)
	}

	// Catch any eventual empty items
	for _, arg := range test {
		// for _, arg := range unfiltered {
		if arg != "" {
			sanitized = append(sanitized, arg)
		}
	}

	return
}

// SystemEditor - This function is a renamed-reexport of the underlying readline.StartEditorWithBuffer
// function, which enables you to conveniently edit files/buffers from within the console application.
// Naturally, the function will block until the editor is exited, and the updated buffer is returned.
// The filename parameter can be used to pass a specific filename.ext pattern, which might be useful
// if the editor has builtin filetype plugin functionality.
func (c *Console) SystemEditor(buffer []byte, filename string) ([]byte, error) {
	runeUpdated, err := c.shell.StartEditorWithBuffer([]rune(string(buffer)), filename)

	return []byte(string(runeUpdated)), err
}
