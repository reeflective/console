package readline

/*
   console - Closed-loop console application for cobra commands
   Copyright (C) 2023 Reeflective

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

import (
	"fmt"
	"sort"
	"strings"

	"github.com/reeflective/readline"
	"github.com/reeflective/readline/inputrc"
	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

const (
	printOn  = "on"
	printOff = "off"
)

// listVars prints the readline global option variables in human-readable format.
func listVars(sh *readline.Shell, buf *cfgBuilder, cmd *cobra.Command) {
	var vars map[string]interface{}

	// Apply other filters to our current list of vars.
	if cmd.Flags().Changed("changed") {
		vars = cfgChanged.Vars
	} else {
		vars = sh.Config.Vars
	}

	if len(vars) == 0 {
		return
	}

	var variables = make([]string, len(sh.Config.Vars))

	for variable := range sh.Config.Vars {
		variables = append(variables, variable)
	}

	sort.Strings(variables)

	fmt.Fprintln(buf)
	fmt.Fprintln(buf, "======= Global Variables =========")
	fmt.Fprintln(buf)

	for _, variable := range variables {
		value := sh.Config.Vars[variable]
		if value == nil || variable == "" {
			continue
		}

		fmt.Fprintf(buf, "%s is set to `%v'\n", variable, value)
	}
}

// listVarsRC returns the readline global options, split according to which are
// supported by which library, and output in .inputrc compliant format.
func listVarsRC(sh *readline.Shell, buf *cfgBuilder, cmd *cobra.Command) {
	// Apply other filters to our current list of vars.
	var vars map[string]interface{}

	// Apply other filters to our current list of vars.
	if cmd.Flags().Changed("changed") {
		vars = cfgChanged.Vars
	} else {
		vars = sh.Config.Vars
	}

	if len(vars) == 0 {
		return
	}

	// Include print all legacy options.
	// Filter them in a separate groups only if NOT being used with --app/--lib
	if !cmd.Flags().Changed("app") && !cmd.Flags().Changed("lib") {
		var legacy []string
		for variable := range filterLegacyVars(vars) {
			legacy = append(legacy, variable)
		}

		sort.Strings(legacy)

		fmt.Fprintln(buf, "# General/legacy Options (generated from reeflective/readline)")

		for _, variable := range legacy {
			value := sh.Config.Vars[variable]
			var printVal string

			if on, ok := value.(bool); ok {
				if on {
					printVal = "on"
				} else {
					printVal = "off"
				}
			} else {
				printVal = fmt.Sprintf("%v", value)
			}

			fmt.Fprintf(buf, "set %s %s\n", variable, printVal)
		}

		// Now we print the App/lib specific.
		var reef []string

		for variable := range filterAppLibVars(vars) {
			reef = append(reef, variable)
		}

		sort.Strings(reef)

		fmt.Fprintln(buf)
		fmt.Fprintln(buf, "# reeflective/readline specific options (generated)")
		fmt.Fprintln(buf, "# The following block is not implemented in GNU C Readline.")
		buf.newCond("reeflective")

		for _, variable := range reef {
			value := sh.Config.Vars[variable]
			var printVal string

			if on, ok := value.(bool); ok {
				if on {
					printVal = printOn
				} else {
					printVal = printOff
				}
			} else {
				printVal = fmt.Sprintf("%v", value)
			}

			fmt.Fprintf(buf, "set %s %s\n", variable, printVal)
		}

		buf.endCond()

		return
	}

	fmt.Fprintln(buf, "# General options (legacy and reeflective)")

	var all []string
	for variable := range vars {
		all = append(all, variable)
	}
	sort.Strings(all)

	for _, variable := range all {
		value := sh.Config.Vars[variable]
		var printVal string

		if on, ok := value.(bool); ok {
			if on {
				printVal = "on"
			} else {
				printVal = "off"
			}
		} else {
			printVal = fmt.Sprintf("%v", value)
		}

		fmt.Fprintf(buf, "set %s %s\n", variable, printVal)
	}
}

func listBinds(sh *readline.Shell, buf *cfgBuilder, _ *cobra.Command, _ string) {
}

func listBindsRC(sh *readline.Shell, buf *cfgBuilder, _ *cobra.Command, _ string) {
}

func listMacros(sh *readline.Shell, buf *cfgBuilder, _ *cobra.Command, _ string) {
}

func listMacrosRC(sh *readline.Shell, buf *cfgBuilder, _ *cobra.Command, _ string) {
}

func bindsQuery(sh *readline.Shell, cmd *cobra.Command, keymap string) {
	binds := sh.Config.Binds[keymap]
	if binds == nil {
		return
	}

	command, _ := cmd.Flags().GetString("query")

	// Make a list of all sequences bound to each command.
	cmdBinds := make([]string, 0)

	for key, bind := range binds {
		if bind.Action != command {
			continue
		}

		cmdBinds = append(cmdBinds, inputrc.Escape(key))
	}

	sort.Strings(cmdBinds)

	switch {
	case len(cmdBinds) == 0:
	case len(cmdBinds) > 5:
		var firstBinds []string

		for i := 0; i < 5; i++ {
			firstBinds = append(firstBinds, "\""+cmdBinds[i]+"\"")
		}

		bindsStr := strings.Join(firstBinds, ", ")
		fmt.Printf("%s can be found on %s ...\n", command, bindsStr)

	default:
		var firstBinds []string

		for _, bind := range cmdBinds {
			firstBinds = append(firstBinds, "\""+bind+"\"")
		}

		bindsStr := strings.Join(firstBinds, ", ")
		fmt.Printf("%s can be found on %s\n", command, bindsStr)
	}
}

func bindsQueryRC(sh *readline.Shell, cmd *cobra.Command, keymap string, indent bool) {
	var commands []string

	for command := range sh.Keymap.Commands() {
		commands = append(commands, command)
	}

	sort.Strings(commands)

	binds := sh.Config.Binds[keymap]
	if binds == nil {
		return
	}

	// Make a list of all sequences bound to each command.
	allBinds := make(map[string][]string)

	for _, command := range commands {
		for key, bind := range binds {
			if bind.Action != command {
				continue
			}

			commandBinds := allBinds[command]
			commandBinds = append(commandBinds, inputrc.Escape(key))
			allBinds[command] = commandBinds
		}
	}

	if indent {
		fmt.Print("    ")
	}
	fmt.Println("# Command binds (autogenerated from reeflective/readline)")
	if indent {
		fmt.Print("    ")
	}
	fmt.Fprintf(cmd.OutOrStdout(), "set keymap %s\n\n", keymap)

	printBindsInputrc(commands, allBinds, indent)
}

func printBindsInputrc(commands []string, all map[string][]string, indent bool) {
	for _, command := range commands {
		commandBinds := all[command]
		sort.Strings(commandBinds)

		if len(commandBinds) > 0 {
			for _, bind := range commandBinds {
				if indent {
					fmt.Print("    ")
				}
				fmt.Printf("\"%s\": %s\n", bind, command)
			}
		}
	}
}

func macrosQuery(sh *readline.Shell, _ *cobra.Command, keymap string) {
	binds := sh.Config.Binds[keymap]
	if len(binds) == 0 {
		return
	}

	var macroBinds []string

	for keys, bind := range binds {
		if bind.Macro {
			macroBinds = append(macroBinds, inputrc.Escape(keys))
		}
	}

	if len(macroBinds) == 0 {
		return
	}

	sort.Strings(macroBinds)

	fmt.Println()
	fmt.Printf("=== Macros (%s)===\n", sh.Keymap.Main())
	fmt.Println()

	for _, key := range macroBinds {
		action := inputrc.Escape(binds[inputrc.Unescape(key)].Action)
		fmt.Printf("%s outputs %s\n", key, action)
	}
}

func macrosQueryRC(sh *readline.Shell, _ *cobra.Command, keymap string, indent bool) {
	binds := sh.Config.Binds[keymap]
	if len(binds) == 0 {
		return
	}
	var macroBinds []string

	for keys, bind := range binds {
		if bind.Macro {
			macroBinds = append(macroBinds, inputrc.Escape(keys))
		}
	}

	sort.Strings(macroBinds)

	fmt.Println()
	if indent {
		fmt.Printf("    ")
	}
	fmt.Println("# Macros (autogenerated from reeflective/readline)")

	for _, key := range macroBinds {
		action := inputrc.Escape(binds[inputrc.Unescape(key)].Action)
		if indent {
			fmt.Printf("    ")
		}
		fmt.Printf("\"%s\": \"%s\"\n", key, action)
	}

	fmt.Println()
}

// getChangedBinds returns the list of changed binds for a given keymap.
func getChangedBinds(keymap string) map[string]inputrc.Bind {
	return nil
}

// Filters out all configuration variables that have not been changed.
func filterChangedBinds(keymap string, binds map[string]inputrc.Bind) map[string]inputrc.Bind {
	if binds == nil {
		return cfgChanged.Binds[keymap]
	}

	changedBinds := cfgChanged.Binds[keymap]
	sequences := maps.Keys(changedBinds)

	for name, bind := range binds {
		if slices.Contains(sequences, name) && !bind.Macro {
			changedBinds[name] = bind
		}
	}

	return changedBinds
}

// getChangedBinds returns the list of changed binds for a given keymap.
func getChangedMacros(keymap string) map[string]inputrc.Bind {
	return nil
}
