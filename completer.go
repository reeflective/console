package console

import (
	"strings"

	"github.com/carapace-sh/carapace"
	"github.com/carapace-sh/carapace/pkg/style"
	completer "github.com/carapace-sh/carapace/pkg/x"
	"github.com/reeflective/readline"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/reeflective/console/internal/completion"
	"github.com/reeflective/console/internal/line"
)

func (c *Console) complete(input []rune, pos int) readline.Completions {
	menu := c.activeMenu()

	// Ensure the carapace library is called so that the function
	// completer.Complete() variable is correctly initialized before use.
	carapace.Gen(menu.Command)

	// Split the line as shell words, only using
	// what the right buffer (up to the cursor)
	args, prefixComp, prefixLine := completion.SplitArgs(input, pos)
	resetCompletionFlagState(menu.Command, args)

	// Prepare arguments for the carapace completer
	// (we currently need those two dummies for avoiding a panic).
	args = append([]string{c.name, "_carapace"}, args...)

	// Call the completer with our current command context.
	completions, err := completer.Complete(menu.Command, args...)

	// The completions are never nil: fill out our own object
	// with everything it contains, regardless of errors.
	raw := make([]readline.Completion, len(completions.Values))

	for idx, val := range completions.Values {
		raw[idx] = readline.Completion{
			Value:       line.UnescapeValue(prefixComp, prefixLine, val.Value),
			Display:     val.Display,
			Description: val.Description,
			Style:       style.SGR(val.Style),
			Tag:         val.Tag,
		}

		if !completions.Nospace.Matches(val.Value) {
			raw[idx].Value = val.Value + " "
		}

		// Remove short/long flags grouping
		// join to single tag group for classic zsh side-by-side view
		switch val.Tag {
		case "shorthand flags", "longhand flags":
			raw[idx].Tag = "flags"
		}
	}

	// Assign both completions and command/flags/args usage strings.
	comps := readline.CompleteRaw(raw)
	comps = comps.Usage("%s", completions.Usage)
	comps = c.justifyCommandComps(comps)

	// If any errors arose from the completion call itself.
	if err != nil {
		comps = readline.CompleteMessage("failed to load config: " + err.Error())
	}

	// Completion status/errors
	for _, msg := range completions.Messages.Get() {
		comps = comps.Merge(readline.CompleteMessage(msg))
	}

	// Suffix matchers for the completions if any.
	suffixes, err := completions.Nospace.MarshalJSON()
	if len(suffixes) > 0 && err == nil {
		comps = comps.NoSpace([]rune(string(suffixes))...)
	}

	// If we have a quote/escape sequence unaccounted
	// for in our completions, add it to all of them.
	comps = comps.Prefix(prefixComp)
	comps.PREFIX = prefixLine

	// Finally, reset our command tree for the next call. Only the commands need
	// regenerating here: the prompt is already bound and no command output was
	// produced, so the full resetPreRun would just be wasted work per keystroke.
	// (resetCommands already re-hides filtered commands.)
	completer.ClearStorage()
	menu.resetCommands()

	return comps
}

func resetCompletionFlagState(root *cobra.Command, args []string) {
	if root == nil {
		return
	}

	target := findCompletionTarget(root, args)
	_ = target.LocalFlags()
	resetCompletionFlagDefaults(target)
	resetArgsLenAtDash(target)
}

func resetCompletionFlagDefaults(target *cobra.Command) {
	if target == nil {
		return
	}

	target.Flags().VisitAll(func(flag *pflag.Flag) {
		flag.Changed = false
		switch value := flag.Value.(type) {
		case pflag.SliceValue:
			var res []string
			if len(flag.DefValue) > 0 && flag.DefValue != "[]" {
				res = append(res, flag.DefValue)
			}

			_ = value.Replace(res)
		default:
			_ = flag.Value.Set(flag.DefValue)
		}
	})
}

func resetArgsLenAtDash(target *cobra.Command) {
	for cmd := target; cmd != nil; cmd = cmd.Parent() {
		resetFlagSetArgsLenAtDash(cmd.Flags(), cmd.DisplayName())
		resetFlagSetArgsLenAtDash(cmd.PersistentFlags(), cmd.DisplayName())
	}
}

func resetFlagSetArgsLenAtDash(fs *pflag.FlagSet, name string) {
	if fs == nil {
		return
	}

	fs.Init(name, pflag.ContinueOnError)
}

func findCompletionTarget(root *cobra.Command, args []string) *cobra.Command {
	cmd := root
	for _, arg := range args {
		if arg == "--" || strings.HasPrefix(arg, "-") {
			break
		}

		next := findSubcommand(cmd, arg)
		if next == nil {
			break
		}
		cmd = next
	}

	return cmd
}

func findSubcommand(cmd *cobra.Command, name string) *cobra.Command {
	if cmd == nil {
		return nil
	}

	for _, sub := range cmd.Commands() {
		if sub.Name() == name || sub.HasAlias(name) {
			return sub
		}
	}

	return nil
}

// justifyCommandComps justifies the descriptions for all commands in all groups
// to the same level, for prettiness. Also, removes any coloring from them, as currently,
// the carapace engine does add coloring to each group, and we don't want this.
func (c *Console) justifyCommandComps(comps readline.Completions) readline.Completions {
	justified := []string{}

	comps.EachValue(func(comp readline.Completion) readline.Completion {
		if !strings.HasSuffix(comp.Tag, "commands") {
			return comp
		}

		justified = append(justified, comp.Tag)
		comp.Style = "" // Remove command coloring

		return comp
	})

	if len(justified) > 0 {
		return comps.JustifyDescriptions(justified...)
	}

	return comps
}

// highlightSyntax - Entrypoint to all input syntax highlighting in the Wiregost console.
func (c *Console) highlightSyntax(input []rune) string {
	// Serve a memoized result when the input has not changed since the last
	// render. The cache is cleared whenever the command tree is regenerated,
	// so a stale tree can never produce a stale highlight.
	key := string(input)
	if cached := c.hlCache.Load(); cached != nil && cached.input == key {
		return cached.output
	}

	highlighted := c.computeHighlight(input)
	c.hlCache.Store(&highlightCache{input: key, output: highlighted})

	return highlighted
}

func (c *Console) computeHighlight(input []rune) string {
	// Split the line as shellwords
	args, unprocessed, err := line.Split(string(input), true)
	if err != nil {
		args = append(args, unprocessed)
	}

	done := make([]string, 0)          // List of processed words, append to
	remain := args                     // List of words to process, draw from
	trimmed := line.TrimSpaces(remain) // Match stuff against trimmed words

	// Highlight the root command when found.
	cmd, _, _ := c.activeMenu().Find(trimmed)
	if cmd != nil {
		done, remain = line.HighlightCommand(done, args, c.activeMenu().Command, c.cmdHighlight)
	}

	// Highlight command flags
	done, remain = line.HighlightCommandFlags(done, remain, c.flagHighlight)

	// Done with everything, add remainind, non-processed words
	done = append(done, remain...)

	// Join all words.
	highlighted := strings.Join(done, "")

	return highlighted
}
