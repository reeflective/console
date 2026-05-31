package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"

	"github.com/reeflective/console"
)

// featureGroupID groups the commands that demonstrate the readline hint and
// async-completion features.
const featureGroupID = "readline"

// setupReadlineHints registers a passive hint provider on the shell. The
// provider is recomputed from the current input line on every refresh and its
// result is shown below the input, in the dedicated "provided" hint lane.
//
// Here it resolves the command being typed and shows its short description.
// Because this lane is independent from completion hints (set by the completion
// engine) and from transient/async status messages (see the `notify`/`hint`
// commands), all three can be displayed at once without clobbering each other.
func setupReadlineHints(app *console.Console) {
	dim := func(format string, args ...any) []rune {
		return []rune("\x1b[2;3m" + fmt.Sprintf(format, args...) + "\x1b[0m")
	}

	app.Shell().Hint.SetProvider(func(line []rune, _ int) []rune {
		fields := strings.Fields(string(line))
		if len(fields) == 0 {
			return dim("type a command — try 'notify', 'hint set ...', or 'scan <Tab>'")
		}

		menu := app.ActiveMenu()
		if menu == nil || menu.Command == nil {
			return nil
		}

		// Find resolves the deepest command matched by the words typed so far.
		cmd, _, err := menu.Find(fields)
		if err != nil || cmd == nil || cmd == menu.Command {
			return nil
		}

		return dim("%s — %s", cmd.CommandPath(), cmd.Short)
	})
}

// readlineFeatureCommands builds the commands demonstrating the hint lanes and
// async completion regeneration. They are added to the main menu.
func readlineFeatureCommands(app *console.Console) []*cobra.Command {
	return []*cobra.Command{
		notifyCommand(app),
		hintCommand(app),
		scanCommand(app),
	}
}

// notifyCommand demonstrates ASYNC status updates in the transient hint lane.
// It starts a background job that pushes status messages from another goroutine
// with Hint.SetTransient; the shell repaints on its own (no keystroke), thanks
// to the async-refresh wake.
func notifyCommand(app *console.Console) *cobra.Command {
	return &cobra.Command{
		Use:     "notify",
		Short:   "Async status updates shown in the hint lane (transient hint + wake)",
		GroupID: featureGroupID,
		Run: func(_ *cobra.Command, _ []string) {
			hint := app.Shell().Hint
			stages := []string{
				"\x1b[33m⠋ connecting…\x1b[0m",
				"\x1b[33m⠙ authenticating…\x1b[0m",
				"\x1b[33m⠹ transferring…\x1b[0m",
				"\x1b[32m✓ transfer complete\x1b[0m",
			}

			go func() {
				for _, stage := range stages {
					time.Sleep(1200 * time.Millisecond)
					hint.SetTransient(stage)
				}

				time.Sleep(1500 * time.Millisecond)
				hint.ClearTransient()
			}()

			fmt.Println("Background job started — watch the hint line below the prompt update on its own (no keystroke needed).")
		},
	}
}

// hintCommand demonstrates SYNCHRONOUS use of the transient hint lane: setting a
// sticky status message that persists across keystrokes (unlike a completion
// hint) until it is cleared or replaced.
func hintCommand(app *console.Console) *cobra.Command {
	hint := &cobra.Command{
		Use:     "hint",
		Short:   "Set or clear a sticky transient hint immediately (non-async)",
		GroupID: featureGroupID,
	}

	hint.AddCommand(&cobra.Command{
		Use:   "set MESSAGE...",
		Short: "Set the transient hint lane to a message (persists until cleared)",
		Args:  cobra.MinimumNArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			app.Shell().Hint.SetTransient("\x1b[36m" + strings.Join(args, " ") + "\x1b[0m")
		},
	})

	hint.AddCommand(&cobra.Command{
		Use:   "clear",
		Short: "Clear the transient hint lane",
		Run: func(_ *cobra.Command, _ []string) {
			app.Shell().Hint.ClearTransient()
		},
	})

	return hint
}

// hostDiscovery is a process-wide singleton: the console rebuilds its command
// tree (and thus re-runs scanCommand) on each completion, so the discovery state
// must persist across those rebuilds rather than being recreated each time.
//
// Seeded with two known hosts so the menu opens and stays open — a single
// candidate would be auto-accepted, closing the menu before any async result
// could be shown.
var hostDiscovery = &discovery{base: []string{"localhost", "gateway"}}

// scanCommand demonstrates ASYNC completions. Its argument completer returns a
// set of hosts that a background "discovery" grows over time; each time a host
// is found, the goroutine calls Shell().RefreshCompletions(), which rebuilds the
// already-open completion menu in place — so hosts appear live while the menu
// stays open, with no keystroke from the user.
func scanCommand(app *console.Console) *cobra.Command {
	scan := &cobra.Command{
		Use:     "scan [HOST]",
		Short:   "Async completions — press Tab after 'scan ' and watch hosts appear live",
		GroupID: featureGroupID,
		Args:    cobra.MaximumNArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("Usage: scan HOST   (press Tab after 'scan ' and watch the menu fill in)")
				return
			}

			fmt.Println("Scanning host:", args[0])
		},
	}

	carapace.Gen(scan).PositionalCompletion(
		carapace.ActionCallback(func(_ carapace.Context) carapace.Action {
			hostDiscovery.start(app)
			return carapace.ActionValues(hostDiscovery.snapshot()...)
		}),
	)

	return scan
}

// discovery simulates an asynchronous completion producer: a background routine
// appends "discovered" hosts to a cache and asks the shell to regenerate the
// open menu in place.
type discovery struct {
	mu      sync.Mutex
	base    []string
	found   []string
	running bool
}

// snapshot returns the current known + discovered hosts.
func (d *discovery) snapshot() []string {
	d.mu.Lock()
	defer d.mu.Unlock()

	return append(append([]string{}, d.base...), d.found...)
}

// start kicks off one discovery run if none is in progress. Each newly found
// host triggers an in-place regeneration of the open completion menu.
func (d *discovery) start(app *console.Console) {
	d.mu.Lock()
	if d.running {
		d.mu.Unlock()
		return
	}

	d.running = true
	d.found = nil
	d.mu.Unlock()

	go func() {
		for i := 1; i <= 8; i++ {
			time.Sleep(900 * time.Millisecond)

			d.mu.Lock()
			d.found = append(d.found, fmt.Sprintf("10.0.0.%d", i))
			d.mu.Unlock()

			// Rebuild the open menu in place with the newly discovered host.
			app.Shell().RefreshCompletions()
		}

		d.mu.Lock()
		d.running = false
		d.mu.Unlock()
	}()
}
