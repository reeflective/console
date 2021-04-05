package gonsole

import (
	"fmt"
	"strings"

	ansi "github.com/acarl005/stripansi"

	"github.com/maxlandon/readline"
)

// Prompt - Computes all prompts used on the shell for a given context.
type Prompt struct {
	Left  string // The leftmost prompt
	Right string // The rightmost prompt, currently same line as left.

	Callbacks map[string]func() string // A list of value callbacks to be used.
	Colors    map[string]string        // Users can also register colors

	Newline bool // If true, leaves a new line before showing command output.
}

// RefreshPromptLog - A simple function to print a string message (a log, or more broadly,
// an asynchronous event) without bothering the user, and by "pushing" the prompt below the message.
// If this function is called while a command is running, the console will simply print the log
// below the current line, and will not print the prompt. In any other case this function will work normally.
func (c *Console) RefreshPromptLog(log string) {
	if c.isExecuting {
		fmt.Print(log)
	} else {
		c.Shell.RefreshPromptLog(log)
	}
}

// Render - The core prompt computes all necessary values, forges a prompt string
// and returns it for being printed by the shell.
func (p *Prompt) Render() (prompt string) {

	// We need the terminal width: the prompt sometimes
	// makes use of both sides for different items.
	sWidth := readline.GetTermWidth()

	// Compute all prompt parts independently
	left, bWidth := p.computeCallbacks(p.Left)
	right, cWidth := p.computeCallbacks(p.Right)

	// Verify that the length of all combined prompt elements is not wider than
	// determined terminal width. If yes, truncate the prompt string accordingly.
	if bWidth+cWidth > sWidth {
		// m.Module = truncate()
	}

	// Get the empty part of the prompt and pad accordingly.
	pad := getPromptPad(sWidth, bWidth, cWidth)

	// Finally, forge the complete prompt string
	prompt = left + pad + right

	// Don't mess with input line colors
	prompt += readline.RESET

	return
}

// computeBase - Computes the base prompt (left-side) with potential custom prompt given.
// Returns the width of the computed string, for correct aggregation of all strings.
func (p *Prompt) computeCallbacks(raw string) (ps string, width int) {
	ps = raw

	// Compute callback values
	for ok, cb := range p.Callbacks {
		ps = strings.Replace(ps, ok, cb(), 1)
	}
	for tok, color := range p.Colors {
		ps = strings.Replace(ps, tok, color, -1)
	}

	width = getRealLength(ps)

	return
}

// getRealLength - Some strings will have ANSI escape codes, which might be wrongly
// interpreted as legitimate parts of the strings. This will bother if some prompt
// components depend on other's length, so we always pass the string in this for
// getting its real-printed length.
func getRealLength(s string) (l int) {
	return len(ansi.Strip(s))
}

func getPromptPad(total, base, context int) (pad string) {
	var padLength = total - base - context
	for i := 0; i < padLength; i++ {
		pad += " "
	}
	return
}

var (
	// defaultColorCallbacks - All colors and effects needed in the main menu
	defaultColorCallbacks = map[string]string{
		// Base readline colors
		"{blink}": "\033[5m", // blinking
		"{bold}":  readline.BOLD,
		"{dim}":   readline.DIM,
		"{fr}":    readline.RED,
		"{g}":     readline.GREEN,
		"{b}":     readline.BLUE,
		"{y}":     readline.YELLOW,
		"{fw}":    readline.FOREWHITE,
		"{bdg}":   readline.BACKDARKGRAY,
		"{br}":    readline.BACKRED,
		"{bg}":    readline.BACKGREEN,
		"{by}":    readline.BACKYELLOW,
		"{blb}":   readline.BACKLIGHTBLUE,
		"{reset}": readline.RESET,
		// Custom colors
		"{ly}":   "\033[38;5;187m",
		"{lb}":   "\033[38;5;117m", // like VSCode var keyword
		"{db}":   "\033[38;5;24m",
		"{bddg}": "\033[48;5;237m",
	}
)
