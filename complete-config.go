package gonsole

import (
	"fmt"
	"sort"
	"strings"

	"github.com/maxlandon/readline"
)

var (
	promptEffectsDesc = map[string]string{
		"{blink}": "blinking", // blinking
		"{bold}":  "bold text",
		"{dim}":   "obscured text",
		"{fr}":    "fore red",
		"{g}":     "fore green",
		"{b}":     "fore blue",
		"{y}":     "fore yellow",
		"{fw}":    "fore white",
		"{bdg}":   "back dark gray",
		"{br}":    "back red",
		"{bg}":    "back green",
		"{by}":    "back yellow",
		"{blb}":   "back light blue",
		"{reset}": "reset effects",
		// Custom colors
		"{ly}":   "light yellow",
		"{lb}":   "light blue (VSCode keyword)", // like VSCode var keyword
		"{db}":   "dark blue",
		"{bddg}": "back dark dark gray",
	}
)

// PromptItems - Queries the console context prompt for all its callbacks and passes them as completions.
func (c *CommandCompleter) PromptItems(lastWord string) (prefix string, comps []*readline.CompletionGroup) {

	cc := c.console.current
	serverPromptItems := cc.Prompt.Callbacks
	promptEffects := cc.Prompt.Colors

	// Items
	sComp := &readline.CompletionGroup{
		Name:         fmt.Sprintf("%s prompt items", cc.Name),
		Descriptions: map[string]string{},
		DisplayType:  readline.TabDisplayMap,
	}

	var keys []string
	for item := range serverPromptItems {
		keys = append(keys, item)
	}
	sort.Strings(keys)
	for _, item := range keys {
		if strings.HasPrefix(item, lastWord) {
			sComp.Suggestions = append(sComp.Suggestions, item)
		}
	}
	comps = append(comps, sComp)

	// Colors & effects
	cComp := &readline.CompletionGroup{
		Name:         "colors/effects",
		Descriptions: map[string]string{},
		DisplayType:  readline.TabDisplayList,
	}

	var colorKeys []string
	for item := range promptEffects {
		colorKeys = append(colorKeys, item)
	}
	sort.Strings(colorKeys)
	for _, item := range colorKeys {
		if strings.HasPrefix(item, lastWord) {
			desc, ok := promptEffectsDesc[item]
			if ok {
				cComp.Suggestions = append(cComp.Suggestions, item)
				cComp.Descriptions[item] = readline.Dim(desc)
			} else {
				cComp.Suggestions = append(cComp.Suggestions, item)
			}
		}
	}
	comps = append(comps, cComp)

	return
}
