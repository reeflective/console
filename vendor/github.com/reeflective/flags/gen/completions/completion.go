package completions

import (
	"errors"
	"reflect"
	"strings"

	"github.com/reeflective/flags/internal/tag"
	comp "github.com/rsteube/carapace"
)

// Completer represents a type that is able to return some completions based on the current carapace Context.
// Please see https://rsteube.github.io/carapace/carapace.html for using the carapace library completers,
// or https://github.com/reeflective/flags/wiki/Completions for an overview of completion features/use.
type Completer interface {
	Complete(ctx comp.Context) comp.Action
}

// compDirective identifies one of reflags' builtin completer functions.
type compDirective int

const (
	// Public directives =========================================================.

	// compError indicates an error occurred and completions should handled accordingly.
	compError compDirective = 1 << iota

	// compNoSpace indicates that the shell should not add a space after
	// the completion even if there is a single completion provided.
	compNoSpace

	// compNoFiles forbids file completion when no other comps are available.
	compNoFiles

	// compFilterExt only complete files that are part of the given extensions.
	compFilterExt

	// compFilterDirs only complete files within a given set of directories.
	compFilterDirs

	// compFiles completes all files found in the current filesystem context.
	compFiles

	// compDirs completes all directories in the current filesystem context.
	compDirs

	// Internal directives (must be below) =======================================.

	// shellCompDirectiveDefault indicates to let the shell perform its default
	// behavior after completions have been provided.
	// This one must be last to avoid messing up the iota count.
	shellCompDirectiveDefault compDirective = 0
)

var errCommandNotFound = errors.New("command not found")

const (
	completeTagName     = "complete"
	completeTagMaxParts = 2
)

func getCompletionAction(name, value, desc string) comp.Action {
	var action comp.Action

	switch strings.ToLower(name) {
	case "nospace":
		return action.NoSpace()
	case "nofiles":
	case "filterext":
		filterExts := strings.Split(value, ",")
		action = comp.ActionFiles(filterExts...).Tag("filtered extensions").NoSpace('/')
	case "filterdirs":
		action = comp.ActionDirectories().NoSpace('/').Tag("filtered directories") // TODO change this
	case "files":
		files := strings.Split(value, ",")
		action = comp.ActionFiles(files...).NoSpace('/')
	case "dirs":
		action = comp.ActionDirectories().NoSpace('/')

	// Should normally not be used often
	case "default":
		return action
	}

	return action
}

// typeCompleterAlt checksw for completer implementations on the type, checks
// if the implementations are on the type of its elements (if slice/map), and
// returns the results.
func typeCompleter(val reflect.Value) (comp.CompletionCallback, bool, bool) {
	isRepeatable := false
	itemsImplement := false

	var completer comp.CompletionCallback

	// Always check that the type itself does implement, even if
	// it's a list of type X that implements the completer as well.
	// If yes, we return this implementation, since it has priority.
	if val.Type().Kind() == reflect.Slice {
		isRepeatable = true

		i := val.Interface()
		if impl, ok := i.(Completer); ok {
			completer = impl.Complete
		} else if val.CanAddr() {
			if impl, ok := val.Addr().Interface().(Completer); ok {
				completer = impl.Complete
			}
		}

		// Else we reassign the value to the list type.
		val = reflect.New(val.Type().Elem())
	}

	// If we did NOT find an implementation on the compound type,
	// check for one on the items.
	if completer == nil {
		i := val.Interface()
		if impl, ok := i.(Completer); ok && impl != nil {
			itemsImplement = true
			completer = impl.Complete
		} else if val.CanAddr() {
			isRepeatable = true
			if impl, ok := val.Addr().Interface().(Completer); ok && impl != nil {
				itemsImplement = true
				completer = impl.Complete
			}
		}
	}

	return completer, isRepeatable, itemsImplement
}

// taggedCompletions builds a list of completion actions with struct tag specs.
func taggedCompletions(tag tag.MultiTag) (comp.CompletionCallback, bool) {
	compTag := tag.GetMany(completeTagName)
	description, _ := tag.Get("description")
	desc, _ := tag.Get("desc")

	if description == "" {
		description = desc
	}

	if len(compTag) == 0 {
		return nil, false
	}

	// We might have several tags, so several actions.
	actions := make([]comp.Action, 0)

	// ---- Example spec ----
	// Args struct {
	//     File string complete:"files,xml"
	//     Remote string complete:"files"
	//     Delete []string complete:"FilterExt,json,go,yaml"
	//     Local []string complete:"FilterDirs,/home/user"
	// }
	for _, tag := range compTag {
		if tag == "" || strings.TrimSpace(tag) == "" {
			continue
		}

		items := strings.SplitAfterN(tag, ",", completeTagMaxParts)

		name, value := strings.TrimSuffix(items[0], ","), ""

		if len(items) > 1 {
			value = strings.TrimSuffix(items[1], ",")
		}

		// build the completion action
		tagAction := getCompletionAction(name, value, description)
		actions = append(actions, tagAction)
	}

	// To be called when completion is needed, merging everything.
	callback := func(ctx comp.Context) comp.Action {
		return comp.Batch(actions...).ToA()
	}

	return callback, true
}

func hintCompletions(tag tag.MultiTag) (comp.CompletionCallback, bool) {
	description, _ := tag.Get("description")
	desc, _ := tag.Get("desc")

	if description == "" {
		description = desc
	}

	if description == "" {
		return nil, false
	}

	callback := func(comp.Context) comp.Action {
		return comp.Action{}.Usage(desc)
	}

	return callback, true
}

// choiceCompletions builds completions from field tag choices.
func choiceCompletions(tag tag.MultiTag, val reflect.Value) comp.CompletionCallback {
	choices := tag.GetMany("choice")

	if len(choices) == 0 {
		return nil
	}

	var allChoices []string

	flagIsList := val.Kind() == reflect.Slice || val.Kind() == reflect.Map

	if flagIsList {
		for _, choice := range choices {
			allChoices = append(allChoices, strings.Split(choice, " ")...)
		}
	} else {
		allChoices = choices
	}

	callback := func(ctx comp.Context) comp.Action {
		return comp.ActionValues(allChoices...)
	}

	return callback
}
