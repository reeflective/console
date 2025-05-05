package completion

import (
	"fmt"
	"os"

	"github.com/carapace-sh/carapace/pkg/style"
	"github.com/carapace-sh/carapace/pkg/xdg"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// DefaultStyleConfig sets some default styles for completion.
func DefaultStyleConfig() {
	// If carapace config file is found, just return.
	if dir, err := xdg.UserConfigDir(); err == nil {
		_, err := os.Stat(fmt.Sprintf("%v/carapace/styles.json", dir))
		if err == nil {
			return
		}
	}

	// Overwrite all default styles for color
	for i := 1; i < 13; i++ {
		styleStr := fmt.Sprintf("carapace.Highlight%d", i)
		style.Set(styleStr, "bright-white")
	}

	// Overwrite all default styles for flags
	style.Set("carapace.FlagArg", "bright-white")
	style.Set("carapace.FlagMultiArg", "bright-white")
	style.Set("carapace.FlagNoArg", "bright-white")
	style.Set("carapace.FlagOptArg", "bright-white")
}

// ResetFlagsDefaults resets all flags to their default values.
//
// Slice flags accumulate per execution (and do not reset),
//
//	so we must reset them manually.
//
// Example:
//
//	Given cmd.Flags().StringSlice("comment", nil, "")
//	If you run a command with --comment "a" --comment "b" you will get
//	the expected [a, b] slice.
//
//	If you run a command again with no --comment flags, you will get
//	[a, b] again instead of an empty slice.
//
//	If you run the command again with --comment "c" --comment "d" flags,
//	you will get [a, b, c, d] instead of just [c, d].
func ResetFlagsDefaults(target *cobra.Command) {
	target.Flags().VisitAll(func(flag *pflag.Flag) {
		flag.Changed = false
		switch value := flag.Value.(type) {
		case pflag.SliceValue:
			var res []string

			if len(flag.DefValue) > 0 && flag.DefValue != "[]" {
				res = append(res, flag.DefValue)
			}

			value.Replace(res)

		default:
			flag.Value.Set(flag.DefValue)
		}
	})
}
