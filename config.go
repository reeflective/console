package gonsole

import "github.com/maxlandon/readline"

var (
	highlightingItemsComps = map[string]string{
		"{command}":          "highlight the command words",
		"{command-argument}": "highlight the command arguments",
		"{option}":           "highlight the option name",
		"{option-argument}":  "highlight the option arguments",
		// We will dynamically add all <$-env> items as well.
	}
)

// ConsoleConfig - The console configuration (prompts, hints, modes, etc)
type ConsoleConfig struct {
	InputMode           readline.InputMode `json:"input_mode"`
	Prompts             map[string]*prompt `json:"prompts"`
	Hints               bool               `json:"hints"`
	MaxTabCompleterRows int                `json:"max_tab_completer_rows"`
	Highlighting        map[string]string  `json:"highlighting"`
}

// prompt - Contains all the information needed for the prompt of a given context.
type prompt struct {
	Left            string `json:"left"`
	Right           string `json:"right"`
	Newline         bool   `json:"newline"`
	Multiline       bool   `json:"multiline"`
	MultilinePrompt string `json:"multiline_prompt"`
}

// loadDefaultConfig - Sane defaults for the gonsole Console.
func (c *Console) loadDefaultConfig() {
	c.config = &ConsoleConfig{
		InputMode:           readline.Vim,
		Prompts:             map[string]*prompt{},
		Hints:               true,
		MaxTabCompleterRows: 50,
		Highlighting: map[string]string{
			"{command}":          readline.BOLD,
			"{command-argument}": readline.FOREWHITE,
			"{option}":           readline.BOLD,
			"{option-argument}":  readline.FOREWHITE,
		},
	}

	// Make a default prompt for this application
	c.config.Prompts[""] = &prompt{
		Left:            "gonsole",
		Right:           "",
		Newline:         true,
		Multiline:       true,
		MultilinePrompt: " > ",
	}

}

// ExportConfig - The console exports its configuration in a JSON struct.
func (c *Console) ExportConfig() (conf *ConsoleConfig) {
	return c.config
}

// LoadConfig - Loads a config struct, but does immediately refresh the prompt.
// Settings will apply as they are needed by the console.
func (c *Console) LoadConfig(conf *ConsoleConfig) {
	// Ensure no fields are nil
	if conf.Prompts == nil {
		p := &prompt{
			Left:            "gonsole",
			Right:           "",
			Newline:         true,
			Multiline:       true,
			MultilinePrompt: " > ",
		}
		conf.Prompts = map[string]*prompt{"": p}
	}
	if conf.Highlighting == nil {
		conf.Highlighting = map[string]string{
			"{command}":          readline.BOLD,
			"{command-argument}": readline.FOREWHITE,
			"{option}":           readline.BOLD,
			"{option-argument}":  readline.FOREWHITE,
		}
	}
	// Then load
	c.config = conf

	return
}
