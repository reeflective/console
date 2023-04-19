package console

import (
	"fmt"
	"sync"

	"github.com/reeflective/readline"
	"github.com/rsteube/carapace"
	"github.com/spf13/cobra"
)

// Menu - A menu is a simple way to seggregate commands based on
// the environment to which they belong. For instance, when using a menu
// specific to some host/user, or domain of activity, commands will vary.
type Menu struct {
	name    string
	active  bool
	prompt  *Prompt
	console *Console

	// Maps interrupt signals (CtrlC/IOF, etc) to specific error handlers.
	interruptHandlers map[error]func(c *Console)

	// The root cobra command/parser is the one returned by the handler provided
	// through the `menu.SetCommands()` function. This command is thus renewed after
	// each command invocation/execution.
	// You can still use it as you want, for instance to introspect the current command
	// state of your menu.
	*cobra.Command

	// Command spawner
	cmds Commands

	// expansionComps - A list of completion generators that are triggered when
	// the given string is detected (anywhere, even in other completions) in the input line.
	expansionComps map[rune]carapace.CompletionCallback

	// History sources peculiar to this menu.
	historyNames []string
	histories    map[string]readline.History

	// Concurrency management
	mutex *sync.RWMutex
}

func newMenu(name string, console *Console) *Menu {
	menu := &Menu{
		console:           console,
		name:              name,
		prompt:            &Prompt{console: console},
		Command:           &cobra.Command{},
		interruptHandlers: make(map[error]func(c *Console)),
		expansionComps:    make(map[rune]carapace.CompletionCallback),
		histories:         make(map[string]readline.History),
		mutex:             &sync.RWMutex{},
	}

	// Add a default in memory history to each menu
	if name != "" {
		name = "(" + name + ")"
	}

	histName := fmt.Sprintf("local history %s", name)
	hist := readline.NewInMemoryHistory()

	menu.historyNames = append(menu.historyNames, histName)
	menu.histories[histName] = hist

	return menu
}

// Name returns the name of this menu.
func (m *Menu) Name() string {
	return m.name
}

// Prompt returns the prompt object for this menu.
func (m *Menu) Prompt() *Prompt {
	return m.prompt
}

// AddHistorySource adds a source of history commands that will
// be accessible to the shell when the menu is active.
func (m *Menu) AddHistorySource(name string, source readline.History) {
	m.mutex.RLock()
	m.historyNames = append(m.historyNames, name)
	m.histories[name] = source
	m.mutex.RUnlock()
}

// AddHistorySourceFile adds a new source of history populated from
// and writing to the specified "filepath" parameter.
func (m *Menu) AddHistorySourceFile(name string, filepath string) {
	m.mutex.RLock()
	m.historyNames = append(m.historyNames, name)
	m.histories[name], _ = readline.NewHistoryFromFile(filepath)
	m.mutex.RUnlock()
}

// DeleteHistorySource removes a history source from the menu.
// This normally should only be used in two cases:
// - You want to replace the default in-memory history with another one.
// - You want to replace one of your history sources for some reason.
func (m *Menu) DeleteHistorySource(name string) {
	if name == m.Name() {
		if name != "" {
			name = " (" + name + ")"
		}

		name = fmt.Sprintf("local history%s", name)
	}

	delete(m.histories, name)

	for i, hname := range m.historyNames {
		if hname == name {
			m.historyNames = append(m.historyNames[:i], m.historyNames[i+1:]...)

			break
		}
	}
}

func (m *Menu) resetCommands() {
	if m.cmds != nil {
		m.Command = m.cmds()
	}

	if m.Command == nil {
		m.Command = &cobra.Command{
			Annotations: make(map[string]string),
		}
	}
}

// menus manages all created menus for the console application.
type menus map[string]*Menu

// current returns the current menu.
func (m *menus) current() *Menu {
	for _, menu := range *m {
		if menu.active {
			return menu
		}
	}

	// Else return the default menu.
	return (*m)[""]
}

// NewMenu - Create a new command menu, to which the user
// can attach any number of commands (with any nesting), as
// well as some specific items like history sources, prompt
// configurations, sets of expanded variables, and others.
func (c *Console) NewMenu(name string) *Menu {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	menu := newMenu(name, c)
	c.menus[name] = menu

	return menu
}

// CurrentMenu - Return the current console menu. Because the Context
// is just a reference, any modifications to this menu will persist.
func (c *Console) CurrentMenu() *Menu {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.menus.current()
}

// Menu returns one of the console menus by name, or nil if no menu is found.
func (c *Console) Menu(name string) *Menu {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.menus[name]
}

// SwitchMenu - Given a name, the console switches its command menu:
// The next time the console rebinds all of its commands, it will only bind those
// that belong to this new menu. If the menu is invalid, i.e that no commands
// are bound to this menu name, the current menu is kept.
func (c *Console) SwitchMenu(menu string) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// Only switch if the target menu was found.
	if target, found := c.menus[menu]; found && target != nil {
		if c.menus.current() != nil {
			c.menus.current().active = false
		}

		target.active = true

		// Remove the currently bound history sources
		// (old menu) and bind the ones peculiar to this one.
		c.shell.DeleteHistory()

		for _, name := range target.historyNames {
			c.shell.AddHistory(name, target.histories[name])
		}
	}
}
