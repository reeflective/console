package gonsole

import (
	"strings"

	"github.com/maxlandon/readline"
)

func (c *CommandCompleter) completeExpansionVariables(lastWord string) (last string, completions []*readline.CompletionGroup) {
	cc := c.console.CurrentMenu()

	// Check if last input is made of several different variables
	allVars := strings.Split(lastWord, "/")
	lastVar := allVars[len(allVars)-1]

	for exp, completer := range cc.expansionComps {

		for _, grp := range completer() {
			var suggs []string
			var evaluated = map[string]string{}

			escape := ""
			if exp == '%' {
				escape = "\\"
			}
			for _, v := range grp.Suggestions {
				if strings.HasPrefix(string(exp)+v, lastVar) {
					suggs = append(suggs, escape+string(exp)+v+"/")
					evaluated[escape+string(exp)+v+"/"] = grp.Descriptions[v]
					continue
				}
				if _, exists := grp.Aliases[v]; exists {
					suggs = append(suggs, escape+string(exp)+v+"/")
					evaluated[escape+string(exp)+v+"/"] = grp.Descriptions[v]
					continue
				}
			}
			grp.Suggestions = suggs
			grp.Descriptions = evaluated
			completions = append(completions, grp)
		}
	}

	return lastVar, completions
}

// ParseExpansionVariables - This function ca be used if you need to have access to a path in which your expansion variables have been evaluated.
func (c *Console) ParseExpansionVariables(args []string) (processed []string, err error) {

	if len(c.CurrentMenu().expansionComps) == 0 {
		return args, nil
	}

	for _, arg := range args {
		for exp, completer := range c.CurrentMenu().expansionComps {

			// Anywhere a $ is assigned means there is an env variable
			if strings.Contains(arg, string(exp)) {
				// if strings.Contains(arg, string(exp)) || strings.Contains(arg, "~") {

				//Split in case env is embedded in path
				envArgs := strings.Split(arg, "/")

				// If its not a path
				if len(envArgs) == 1 {
					processed = append(processed, handleCuratedVar(arg, exp, completer()))
				}

				// If len of the env var split is > 1, its a path
				if len(envArgs) > 1 {
					processed = append(processed, handleEmbeddedVar(arg, exp, completer()))
				}
			} else if arg != "" && arg != " " {
				// Else, if arg is not an environment variable, return it as is
				processed = append(processed, arg)
			}
		}
	}

	return
}

// handleCuratedVar - Replace an environment variable alone and without any undesired characters attached
func handleCuratedVar(arg string, exp rune, grps []*readline.CompletionGroup) (value string) {
	if strings.HasPrefix(arg, string(exp)) && arg != "" && arg != string(exp) {
		envVar := strings.TrimPrefix(arg, string(exp))
		// var expValue string
		// var found bool
		for _, grp := range grps {
			val, ok := grp.Descriptions[envVar]
			_, exists := grp.Aliases[envVar]
			if !ok && !exists {
				continue
			} else if !ok && exists {
				return val
			}
			return val
			// expValue = val
		}
		// if found {
		//         return expValue
		// }
		return envVar
	}
	// if arg != "" && arg == "~" {
	//         return clientEnv["HOME"]
	// }

	return arg
}

// handleEmbeddedVar - Replace an environment variable that is in the middle of a path, or other one-string combination
func handleEmbeddedVar(arg string, exp rune, grps []*readline.CompletionGroup) (value string) {

	envArgs := strings.Split(arg, "/")
	var path []string

	for _, arg := range envArgs {
		if strings.HasPrefix(arg, string(exp)) && arg != "" && arg != string(exp) {
			envVar := strings.TrimPrefix(arg, string(exp))
			var expValue string
			var found bool
			for _, grp := range grps {
				val, ok := grp.Descriptions[envVar]
				_, exists := grp.Aliases[envVar]
				if !ok && !exists {
					continue
				} else if !ok && exists {
					found = true
					expValue = val
					break
				}
				found = true
				expValue = val
			}
			// Err will be caught when command is ran anyway, or completion will stop...
			if !found {
				path = append(path, arg)
			} else {
				path = append(path, expValue)
			}

			// } else if arg != "" && arg == "~" {
			//         path = append(path, clientEnv["HOME"])
		} else if arg != " " && arg != "" {
			path = append(path, arg)
		}
	}

	return strings.Join(path, "/")
}

// parseTokens - Parse and process any special tokens that are not treated by environment-like parsers.
func (c *Console) parseTokens(sanitized []string) (parsed []string, err error) {

	// PATH SPACE TOKENS
	// Catch \ tokens, which have been introduced in paths where some directories have spaces in name.
	// For each of these splits, we concatenate them with the next string.
	// This will also inspect commands/options/arguments, but there is no reason why a backlash should be present in them.
	var pathAdjusted []string
	var roll bool
	var arg string
	for i := range sanitized {
		if strings.HasSuffix(sanitized[i], "\\") {
			// If we find a suffix, replace with a space. Go on with next input
			arg += strings.TrimSuffix(sanitized[i], "\\") + " "
			roll = true
		} else if roll {
			// No suffix but part of previous input. Add it and go on.
			arg += sanitized[i]
			pathAdjusted = append(pathAdjusted, arg)
			arg = ""
			roll = false
		} else {
			// Default, we add our path and go on.
			pathAdjusted = append(pathAdjusted, sanitized[i])
		}
	}
	parsed = pathAdjusted

	// Add new function here, act on parsed []string from now on, not sanitized
	return
}
