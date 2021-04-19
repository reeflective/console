package gonsole

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/maxlandon/readline"
	"gopkg.in/AlecAivazis/survey.v1"
)

// AddConfigCommand - The console will add a command used to manage all elements of the console
// for any context, and to save such elements into configurations, ready for export. You can
// choose both the command name and the group, for avoiding command collision with your owns.
func (c *Console) AddConfigCommand(name, group string) {
	c.configCommandName = name

	for _, cc := range c.contexts {

		// Root
		conf := cc.AddCommand(name,
			"manage the console configuration elements and exports/imports",
			"",
			group,
			[]string{""},
			func() interface{} { return &Config{console: c} })
		conf.SubcommandsOptional = true

		// Set values
		set := conf.AddCommand("set",
			"set elements of the console, stored in a configuration",
			"",
			"builtin",
			[]string{""},
			func() interface{} { return &ConfigSet{console: c} })

		set.AddCommand("input",
			"set the input editing mode of the console (Vim/Emacs)",
			"",
			"",
			[]string{""},
			func() interface{} { return &InputMode{console: c} })

		set.AddCommand("hints",
			"turn the console hints on/off",
			"",
			"",
			[]string{""},
			func() interface{} { return &Hints{console: c} })

		set.AddCommand("max-tab-completer-rows",
			"set the maximum number of completion rows",
			"",
			"",
			[]string{""},
			func() interface{} { return &MaxTabCompleterRows{console: c} })

		prompt := set.AddCommand("prompt",
			"set prompt strings for one of the available contexts",
			"",
			"",
			[]string{""},
			func() interface{} { return &PromptSet{console: c} })
		prompt.AddArgumentCompletionDynamic("Prompt", c.Completer.PromptItems)

		multiline := set.AddCommand("prompt-multiline",
			"set/enable/disable multiline prompt strings for one of the available contexts",
			"",
			"",
			[]string{""},
			func() interface{} { return &PromptSet{console: c} })
		multiline.AddArgumentCompletionDynamic("Prompt", c.Completer.PromptItems)

		set.AddCommand("highlight",
			"set the highlighting of tokens in the command line",
			"",
			"",
			[]string{""},
			func() interface{} { return &HighlightSyntax{console: c} })

		// Export configuration
		export := conf.AddCommand("export",
			"export the current console configuration as a JSON object in a file, or STDOUT",
			"",
			"builtin",
			[]string{""},
			func() interface{} { return &ConfigExport{console: c} })
		export.AddOptionCompletionDynamic("Save", c.Completer.CompleteLocalPath)

	}
}

// AddConfigSubCommand - Allows the user to bind specialized subcommands to the config root command. This is useful if, for
// example, you want to save the console configuration on a remote server. You can then add any subcommand to this added one.
func (c *Console) AddConfigSubCommand(name, short, long, group string, filters []string, data func() interface{}) *Command {
	for _, cc := range c.contexts {
		for _, cmd := range cc.Commands() {
			if cmd.Name == c.configCommandName {
				return cmd.AddCommand(name, short, long, group, filters, data)
			}
		}
	}
	return nil
}

// Config - Manage console configuration. Prints current by default
type Config struct {
	console *Console
}

// Execute - Manage console configuration. Prints current by default
func (c *Config) Execute(args []string) (err error) {
	conf := c.console.config

	fmt.Println(readline.Bold(readline.Blue(" Console configuration\n")))

	// Elements applying to all contexts.
	fmt.Println(readline.Yellow("Global"))

	var input string
	if conf.InputMode == readline.Vim {
		input = readline.Bold("Vim")
	} else {
		input = readline.Bold("Emacs")
	}
	pad := fmt.Sprintf("%-15s", "Input mode")
	fmt.Printf(" "+pad+"    %s%s%s\n", readline.BOLD, input, readline.RESET)
	pad = fmt.Sprintf("%-15s", "Console hints")
	fmt.Printf(" "+pad+"    %s%t%s\n", readline.BOLD, conf.Hints, readline.RESET)
	fmt.Println()

	// Print context-specific configuration elements
	cc := c.console.current
	promptConf := conf.Prompts[cc.Name]

	fmt.Println(readline.Yellow(" " + cc.Name))

	pad = fmt.Sprintf("%-15s", "Prompt (left)")
	fmt.Printf(" "+pad+"    %s%s%s\n", readline.BOLD, promptConf.Left, readline.RESET)
	pad = fmt.Sprintf("%-15s", "Prompt (right)")
	fmt.Printf(" "+pad+"    %s%s%s\n", readline.BOLD, promptConf.Right, readline.RESET)
	pad = fmt.Sprintf("%-15s", "Multiline")
	fmt.Printf(" "+pad+"    %s%t%s\n", readline.BOLD, promptConf.Multiline, readline.RESET)
	pad = fmt.Sprintf("%-15s", "Multiline prompt")
	fmt.Printf(" "+pad+"    %s%s%s\n", readline.BOLD, promptConf.MultilinePrompt, readline.RESET)
	pad = fmt.Sprintf("%-15s", "Newline")
	fmt.Printf(" "+pad+"    %s%t%s\n", readline.BOLD, promptConf.Newline, readline.RESET)

	// Check if this config has been saved (they should be identical)
	// req := &clientpb.GetConsoleConfigReq{}
	// res, err := transport.RPC.LoadConsoleConfig(context.Background(), req, grpc.EmptyCallOption{})
	// if err != nil {
	//         fmt.Printf(util.Warn + "Could not check if current config is saved\n")
	//         return
	// }
	// // An error thrown in the request means we did not find the configuration.
	// if res.Response.Err != "" {
	//         fmt.Printf(util.Warn + "Current configuration is not saved, type 'config save' to do so.\n")
	//         return
	// }
	//
	// cf := res.Config
	// if (cf.ServerPromptRight == conf.ServerPrompt.Right) && (cf.ServerPromptLeft == conf.ServerPrompt.Left) &&
	//         (cf.SliverPromptRight == conf.SliverPrompt.Right) && (cf.SliverPromptLeft == conf.SliverPrompt.Left) &&
	//         (cf.Vim == conf.Vim) && (cf.Hints == conf.Hints) {
	//         fmt.Printf(util.Info + "Current configuration is saved\n")
	// } else {
	//         fmt.Printf(util.Warn + "Current configuration is not saved, type 'config save' to do so.\n")
	// }
	return
}

// ConfigSet - Set configuration elements of the console
type ConfigSet struct {
	console *Console
}

// Execute - Set configuration elements of the console
func (c *ConfigSet) Execute(args []string) (err error) {
	return
}

// InputMode - Set the input editing mode of the console
type InputMode struct {
	Positional struct {
		Input string `description:"Input/editing mode"`
	} `positional-args:"true"`
	console *Console
}

// Execute - Set the input editing mode of the console
func (i *InputMode) Execute(args []string) (err error) {
	conf := i.console.config

	switch i.Positional.Input {
	case "vi", "vim":
		conf.InputMode = readline.Vim
		i.console.Shell.InputMode = readline.Vim
	case "emacs":
		conf.InputMode = readline.Emacs
		i.console.Shell.InputMode = readline.Emacs
	default:
		fmt.Printf(errorStr+"Invalid argument: %s (must be 'vim'/'vi' or 'emacs')\n", i.Positional.Input)
	}
	fmt.Printf(info+"Console input mode: %s\n", readline.Yellow(i.Positional.Input))

	return
}

// Hints - Turn the hints on/off
type Hints struct {
	Positional struct {
		Display string `description:"show / hide command hints" required:"yes"`
	} `positional-args:"yes" required:"yes"`
	console *Console
}

// Execute - Turn the hints on/off
func (c *Hints) Execute(args []string) (err error) {
	conf := c.console.config

	switch c.Positional.Display {
	case "show", "on":
		conf.Hints = true
		fmt.Printf(info+"Console hints: %s\n", readline.Yellow(c.Positional.Display))
	case "hide", "off":
		conf.Hints = false
		c.console.Shell.HintText = nil
		fmt.Printf(info+"Console hints: %s\n", readline.Yellow(c.Positional.Display))
	default:
		fmt.Printf(errorStr+"Invalid argument: %s (must be 'hide'/'on' or 'show'/'off')\n", c.Positional.Display)
		return nil
	}
	return
}

// MaxTabCompleterRows - Set the maximum number of completion rows
type MaxTabCompleterRows struct {
	Positional struct {
		Rows int `description:"maximum number of completion rows to print" required:"yes"`
	} `positional-args:"yes" required:"yes"`
	console *Console
}

// Execute - Set the maximum number of completion rows
func (m *MaxTabCompleterRows) Execute(args []string) (err error) {
	conf := m.console.config
	conf.MaxTabCompleterRows = m.Positional.Rows
	fmt.Printf(info+"Max tab completer rows: %d\n", m.Positional.Rows)
	return
}

// ConfigExport - Export the current console configuration as a JSON object in a file.
type ConfigExport struct {
	Options struct {
		Save   string `long:"save" short:"s" description:"path to save the configuration (default: working dir)"`
		Output bool   `long:"output" short:"o" description:"if set, only print the JSON config to STDOUT"`
	} `group:"export options"`
	console *Console
}

// Execute - Export the current console configuration as a JSON object in a file.
func (c *ConfigExport) Execute(args []string) (err error) {
	conf := c.console.config

	// Pretty-print format marshaling
	configBytes, err := json.MarshalIndent(conf, "", "    ")
	if err != nil {
		fmt.Printf(errorStr+"Error marshaling config to JSON: %s .\n", err.Error())
		return
	}

	// Print to STDOUT if asked
	if c.Options.Output {
		fmt.Println(configBytes)
		return
	}

	// Else save to file
	save := c.Options.Save
	if save == "" {
		save, _ = os.Getwd()
	}
	shaID := md5.New()
	saveTo, err := saveLocation(save, fmt.Sprintf("gonsole_%x.cfg", hex.EncodeToString(shaID.Sum([]byte{})[:5])))
	if err != nil {
		fmt.Printf(errorStr+"%s\n", err)
		return
	}
	err = ioutil.WriteFile(saveTo, configBytes, 0600)
	if err != nil {
		fmt.Printf(errorStr+"Failed to write to %s\n", err)
		return
	}
	fmt.Printf(errorStr+"Console configuration JSON saved to: %s\n", saveTo)

	return
}

func saveLocation(save, defaultName string) (string, error) {
	var saveTo string
	if save == "" {
		save, _ = os.Getwd()
	}
	fi, err := os.Stat(save)
	if os.IsNotExist(err) {
		log.Printf(info+"%s does not exist\n", save)
		if strings.HasSuffix(save, "/") {
			log.Printf("%s is dir\n", save)
			os.MkdirAll(save, 0700)
			saveTo, _ = filepath.Abs(path.Join(saveTo, defaultName))
		} else {
			log.Printf("%s is not dir\n", save)
			saveDir := filepath.Dir(save)
			_, err := os.Stat(saveTo)
			if os.IsNotExist(err) {
				os.MkdirAll(saveDir, 0700)
			}
			saveTo, _ = filepath.Abs(save)
		}
	} else {
		log.Printf("%s does exist\n", save)
		if fi.IsDir() {
			log.Printf("%s is dir\n", save)
			saveTo, _ = filepath.Abs(path.Join(save, defaultName))
		} else {
			log.Printf("%s is not dir\n", save)
			prompt := &survey.Confirm{Message: "Overwrite existing file?"}
			var confirm bool
			survey.AskOne(prompt, &confirm, nil)
			if !confirm {
				return "", errors.New("File already exists")
			}
			saveTo, _ = filepath.Abs(save)
		}
	}
	return saveTo, nil
}

// HighlightSyntax - Set the highlighting of tokens in the command line.
type HighlightSyntax struct {
	Positional struct {
		Color string `description:"color to use for highlighting. Can be anything (some defaults colors/effects are completed)" required:"yes"`
		Token string `description:"token (word type) to highlight with the given color (completed)" required:"yes"`
	} `positional-args:"true" required:"yes"`
	console *Console
}

// Execute - Set the highlighting of tokens in the command line.
func (h *HighlightSyntax) Execute(args []string) (err error) {
	for token := range h.console.config.Highlighting {
		if token == h.Positional.Token {
			h.console.config.Highlighting[token] = h.Positional.Color
		}
	}
	return
}

// PromptSet - Set prompt strings for one of the available contexts.
type PromptSet struct {
	Positional struct {
		Prompt string `description:"prompt string. Pass an empty '' to deactivate it (default colors/effect/items completed)" required:"yes"`
	} `positional-args:"yes" required:"yes"`
	Options struct {
		Right         bool   `long:"right" short:"r" description:"apply changes to the right-side prompt"`
		Left          bool   `long:"left" short:"l" description:"apply changes to the left-side prompt"`
		NewlineBefore bool   `long:"newline-before" short:"b" description:"if true, a blank line is left before the prompt is printed"`
		NewlineAfter  bool   `long:"newline-after" short:"a" description:"if true, a blank line is left before the command output is printed"`
		Context       string `long:"context" short:"c" description:"name of the context for which to set the prompt (completed)" default:"default"`
	} `group:"export options"`
	console *Console
}

// Execute - Set prompt strings for one of the available contexts.
func (p *PromptSet) Execute(args []string) (err error) {
	if len(args) > 0 {
		fmt.Printf(warn+"Detected undesired remaining arguments: %s\n", readline.Bold(strings.Join(args, " ")))
		fmt.Printf("    Please use \\ dashes for each space in prompt string (input readline doesn't detect them)\n")
		fmt.Printf(readline.Yellow("    The current value has therefore not been saved.\n"))
		return
	}

	var cc *Context
	if p.Options.Context == "current" {
		cc = p.console.current
	} else {
		cc = p.console.GetContext(p.Options.Context)
	}
	if cc == nil {
		fmt.Printf(errorStr+"Invalid menu/context name: %s .\n", p.Options.Context)
		return
	}

	conf := p.console.config

	// Which prompt side did we set
	var side string
	if p.Options.Right {
		side = "(right)"
		cc.Prompt.Right = p.Positional.Prompt
		conf.Prompts[cc.Name].Right = p.Positional.Prompt
	}
	if p.Options.Left {
		side = "(left)"
		cc.Prompt.Left = p.Positional.Prompt
		conf.Prompts[cc.Name].Left = p.Positional.Prompt
	}
	if !p.Options.Left && !p.Options.Right {
		side = "(left)"
		cc.Prompt.Left = p.Positional.Prompt
		conf.Prompts[cc.Name].Left = p.Positional.Prompt
	}

	// TODO: should be changed because not handy to use like this
	if p.Options.NewlineAfter {
		cc.Prompt.Newline = true
		conf.Prompts[cc.Name].Newline = true
	}

	if p.Positional.Prompt == "\"\"" || p.Positional.Prompt == "''" {
		fmt.Printf(info + "Detected empty prompt string: deactivating the corresponding prompt.\n")
		return
	}

	fmt.Printf(info+"Server prompt %s : %s\n", side, readline.Bold(p.Positional.Prompt))
	return
}

// PromptSetMultiline - Set multiline prompt strings for one of the available contexts.
type PromptSetMultiline struct {
	Positional struct {
		Prompt string `description:"multine prompt string. Leave empty and '--enable' to activate. Pass '' to deactivate"`
	} `positional-args:"yes"`
	Options struct {
		Enable  bool   `long:"enable" short:"e" description:"if true, the prompt will be a 2-line prompt"`
		Context string `long:"context" short:"c" description:"name of the context for which to set the prompt (completed)" default:"current"`
	} `group:"export options"`
	console *Console
}

// Execute - Set multiline prompt strings for one of the available contexts.
func (p *PromptSetMultiline) Execute(args []string) (err error) {
	if len(args) > 0 {
		fmt.Printf(warn+"Detected undesired remaining arguments: %s\n", readline.Bold(strings.Join(args, " ")))
		fmt.Printf("    Please use \\ dashes for each space in prompt string (input readline doesn't detect them)\n")
		fmt.Printf(readline.Yellow("    The current value has therefore not been saved.\n"))
		return
	}

	var cc *Context
	if p.Options.Context == "current" {
		cc = p.console.current
	} else {
		cc = p.console.GetContext(p.Options.Context)
	}
	if cc == nil {
		fmt.Printf(errorStr+"Invalid menu/context name: %s .\n", p.Options.Context)
		return
	}

	conf := p.console.config

	if p.Positional.Prompt == "\"\"" || p.Positional.Prompt == "''" {
		p.console.Shell.Multiline = false
		conf.Prompts[cc.Name].Multiline = false
		fmt.Printf(info + "Detected empty prompt string: deactivating the corresponding prompt.\n")
		return
	}

	if p.Positional.Prompt == "" && p.Options.Enable {
		p.console.Shell.Multiline = true
		conf.Prompts[cc.Name].Multiline = true
		return
	}

	p.console.Shell.MultilinePrompt = p.Positional.Prompt
	conf.Prompts[cc.Name].MultilinePrompt = p.Positional.Prompt

	return
}
