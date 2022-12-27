package console

import (
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

	// A menu being very similar to a shell context, it embeds a single cobra
	// root command, which is considered in its traditional role here: a global parser.
	*cobra.Command

	// The completer allows to further register completions, including those taking
	// care of parsing/expanding environment variables.
	*carapace.Carapace

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
		console:        console,
		name:           name,
		prompt:         newPrompt(console),
		Command:        &cobra.Command{},
		expansionComps: make(map[rune]carapace.CompletionCallback),
		histories:      make(map[string]readline.History),
		mutex:          &sync.RWMutex{},
	}

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
	m.historyNames = append(m.historyNames, name)
	m.histories[name] = source
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
		c.shell.DeleteHistorySource()

		for _, name := range target.historyNames {
			c.shell.AddHistorySource(name, target.histories[name])
		}
	}
}
