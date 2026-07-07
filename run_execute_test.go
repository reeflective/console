package console

import (
	"context"
	"testing"

	"github.com/spf13/cobra"
)

func TestConsoleRunMenuCommandRunsPreparedMenu(t *testing.T) {
	c := New("test")
	menu := c.ActiveMenu()

	var ran bool
	root := &cobra.Command{Use: "root"}
	root.AddCommand(&cobra.Command{
		Use: "run",
		Run: func(*cobra.Command, []string) {
			ran = true
		},
	})
	menu.Command = root

	if err := c.RunMenuCommand(context.Background(), menu, []string{"run"}, false); err != nil {
		t.Fatal(err)
	}
	if !ran {
		t.Fatal("RunMenuCommand did not run the target command")
	}
	if c.isExecuting.Load() {
		t.Fatal("RunMenuCommand left the console marked as executing")
	}
}
