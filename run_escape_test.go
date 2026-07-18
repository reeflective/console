package console

import (
	"context"
	"reflect"
	"testing"

	"github.com/spf13/cobra"
)

// TestRunCommandLineEscapeMode verifies that Console.SetEscapeMode flows all the
// way through to the argument vector a command actually receives.
func TestRunCommandLineEscapeMode(t *testing.T) {
	tests := []struct {
		name string
		mode EscapeMode
		line string
		want []string
	}{
		{"shell default eats backslashes", EscapeShell, `run C:\Windows\Temp`, []string{`C:WindowsTemp`}},
		{"literal preserves backslashes", EscapeLiteral, `run C:\Windows\Temp`, []string{`C:\Windows\Temp`}},
		{"literal preserves trailing backslash", EscapeLiteral, `run C:\Windows\Temp\`, []string{`C:\Windows\Temp\`}},
		{"literal still groups quotes", EscapeLiteral, `run "a b" C:\x`, []string{"a b", `C:\x`}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := New("test")
			c.SetEscapeMode(tc.mode)
			menu := c.ActiveMenu()

			var got []string
			root := &cobra.Command{Use: "root"}
			root.AddCommand(&cobra.Command{
				Use: "run",
				Run: func(_ *cobra.Command, args []string) {
					got = args
				},
			})
			menu.Command = root

			if err := menu.RunCommandLine(context.Background(), tc.line); err != nil {
				t.Fatalf("RunCommandLine(%q): %v", tc.line, err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("RunCommandLine(%q) args = %q, want %q", tc.line, got, tc.want)
			}
		})
	}
}
