package console

import (
	"context"
	"reflect"
	"testing"

	"github.com/spf13/cobra"
)

func TestRunCommandArgsResetsFlagDefaults(t *testing.T) {
	c := New("test")
	menu := c.ActiveMenu()

	type runState struct {
		verbose        bool
		verboseChanged bool
		items          []string
		itemsChanged   bool
	}
	var states []runState

	root := &cobra.Command{Use: "root"}
	cmd := &cobra.Command{
		Use: "run",
		RunE: func(cmd *cobra.Command, _ []string) error {
			verbose, err := cmd.Flags().GetBool("verbose")
			if err != nil {
				return err
			}
			items, err := cmd.Flags().GetStringSlice("item")
			if err != nil {
				return err
			}

			states = append(states, runState{
				verbose:        verbose,
				verboseChanged: cmd.Flags().Changed("verbose"),
				items:          append([]string(nil), items...),
				itemsChanged:   cmd.Flags().Changed("item"),
			})

			return nil
		},
	}
	cmd.Flags().Bool("verbose", false, "")
	cmd.Flags().StringSlice("item", []string{"base"}, "")
	root.AddCommand(cmd)
	menu.SetCommands(func() *cobra.Command { return root })

	if err := menu.RunCommandArgs(context.Background(), []string{"run", "--verbose", "--item", "one", "--item", "two"}); err != nil {
		t.Fatal(err)
	}
	if err := menu.RunCommandArgs(context.Background(), []string{"run"}); err != nil {
		t.Fatal(err)
	}

	if len(states) != 2 {
		t.Fatalf("executed %d times, want 2", len(states))
	}
	if !states[0].verbose || !states[0].verboseChanged {
		t.Fatalf("first run verbose state = %+v, want true/changed", states[0])
	}
	if !reflect.DeepEqual(states[0].items, []string{"one", "two"}) || !states[0].itemsChanged {
		t.Fatalf("first run slice state = %+v, want [one two]/changed", states[0])
	}
	if states[1].verbose || states[1].verboseChanged {
		t.Fatalf("second run verbose state = %+v, want false/not changed", states[1])
	}
	if !reflect.DeepEqual(states[1].items, []string{"base"}) || states[1].itemsChanged {
		t.Fatalf("second run slice state = %+v, want [base]/not changed", states[1])
	}
}
