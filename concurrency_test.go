package console

import (
	"errors"
	"sync"
	"testing"

	"github.com/spf13/cobra"
)

// TestConcurrentStateAccess stresses the console's shared state (filters, the
// menus map, and per-menu interrupt handlers) from many goroutines at once.
//
// It is meant to be run with the race detector (`go test -race`). Before the
// locking fixes, these paths mutated maps/slices under a read lock (or no lock
// at all), which the detector flags and which can panic on concurrent map
// writes in production.
func TestConcurrentStateAccess(t *testing.T) {
	c := New("test")

	// Give the active menu a small command tree so that ActiveFiltersFor has
	// something to recurse over while filters are being mutated concurrently.
	menu := c.ActiveMenu()
	menu.SetCommands(func() *cobra.Command {
		root := &cobra.Command{Use: "root"}
		child := &cobra.Command{
			Use:         "child",
			Annotations: map[string]string{CommandFilterKey: "filterA,filterB"},
			Run:         func(*cobra.Command, []string) {},
		}
		root.AddCommand(child)
		return root
	})
	menu.resetPreRun()

	errInt := errors.New("interrupt")

	const workers = 64

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(i int) {
			defer wg.Done()

			// Filters: concurrent writers (Hide/Show) and readers (ActiveFiltersFor).
			c.HideCommands("filterA", "filterB")
			c.ShowCommands("filterA")

			m := c.ActiveMenu()
			for _, cmd := range m.Command.Commands() {
				_ = m.ActiveFiltersFor(cmd)
			}

			// Menus map: concurrent creation and lookup.
			_ = c.NewMenu("menu")
			_ = c.Menu("menu")
			_ = c.ActiveMenu()

			// Interrupt handlers map: concurrent writers.
			m.AddInterrupt(errInt, func(*Console) {})
			m.DelInterrupt(errInt)
		}(i)
	}

	wg.Wait()
}
