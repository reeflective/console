package flags

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/reeflective/flags/internal/scan"
	"github.com/reeflective/flags/internal/tag"
)

// parseFlagTag now also handles some of the tags used in jessevdk/go-flags.
func parseFlagTag(field reflect.StructField, options opts) (*Flag, *tag.MultiTag, error) {
	flag := &Flag{}

	ignorePrefix := false
	flag.Name = camelToFlag(field.Name, options.FlagDivider)

	// Parse the struct tag
	flagTags, skip, err := getFlagTags(field, options)
	if err != nil {
		return nil, nil, err
	}

	if skip {
		return nil, nil, nil
	}

	// Parse all base struct tag flags attributes and populate the flag object.
	if skip, ignorePrefix = parseBaseAttributes(flagTags, flag, options); skip {
		return nil, flagTags, nil
	}

	flag.DefValue = flagTags.GetMany("default")
	setFlagChoices(flag, flagTags.GetMany("choice"))
	flag.OptionalValue = flagTags.GetMany("optional-value")

	if options.Prefix != "" && !ignorePrefix {
		flag.Name = options.Prefix + flag.Name
	}

	return flag, flagTags, nil
}

// getFlagTags tries to parse any struct tag we need, and tells the caller if
// we should actually build a flag object out of the struct field, or skip it.
func getFlagTags(field reflect.StructField, options opts) (*tag.MultiTag, bool, error) {
	flagTags, none, err := tag.GetFieldTag(field)
	if err != nil {
		return nil, true, fmt.Errorf("%w: %s", ErrTag, err.Error())
	}

	// If the global options specify that we must build a flag
	// out of each struct field, regardless of them being tagged.
	if options.ParseAll {
		return &flagTags, false, nil
	}

	// Else we skip this field only if there's not tag on it
	if none {
		return &flagTags, true, nil
	}

	return &flagTags, false, nil
}

// parseBaseAttributes checks which type of struct tags we found, parses them
// accordingly (legacy, or not), taking into account any global config settings.
func parseBaseAttributes(flagTags *tag.MultiTag, flag *Flag, options opts) (skip, ignorePrefix bool) {
	sflagsTag, _ := flagTags.Get(options.FlagTag)
	sflagValues := strings.Split(sflagsTag, ",")

	if sflagsTag != "" && len(sflagValues) > 0 {
		// Either we have found the legacy flags tag value.
		skip, ignorePrefix = parseflagsTag(sflagsTag, flag)
		if skip {
			return true, false
		}
	} else {
		// Or we try for the go-flags tags.
		parseGoFlagsTag(flagTags, flag)
	}

	// Descriptions
	if desc, isSet := flagTags.Get("desc"); isSet && desc != "" {
		flag.Usage = desc
	} else if desc, isSet := flagTags.Get("description"); isSet && desc != "" {
		flag.Usage = desc
	}

	// Requirements
	if required, _ := flagTags.Get("required"); !isStringFalsy(required) {
		flag.Required = true
	}

	return false, ignorePrefix
}

// parseflagsTag parses only the original tag values of this library flags.
func parseflagsTag(flagsTag string, flag *Flag) (skip, ignorePrefix bool) {
	values := strings.Split(flagsTag, ",")

	// Base / legacy flags tag
	switch fName := values[0]; fName {
	case "-":
		return true, ignorePrefix
	case "":
	default:
		fNameSplitted := strings.Split(fName, " ")
		if len(fNameSplitted) > 1 {
			fName = fNameSplitted[0]
			flag.Short = fNameSplitted[1]
		}

		if strings.HasPrefix(fName, "~") {
			flag.Name = fName[1:]
			ignorePrefix = true
		} else {
			flag.Name = fName
		}
	}

	flag.Hidden = hasOption(values[1:], "hidden")
	flag.Deprecated = hasOption(values[1:], "deprecated")

	return false, ignorePrefix
}

// parseGoFlagsTag parses only the tags used by jessevdk/go-flags.
func parseGoFlagsTag(flagTags *tag.MultiTag, flag *Flag) {
	if short, found := flagTags.Get("short"); found && short != "" {
		shortR, err := getShortName(short)
		if err == nil {
			flag.Short = string(shortR)
		}
		if long, found := flagTags.Get("long"); found && long != "" {
			flag.Name, _ = flagTags.Get("long")
		}
	} else if long, found := flagTags.Get("long"); found && long != "" {
		// Or we have only a short tag being specified.
		flag.Name = long
	}
}

func parseEnvTag(flagName string, field reflect.StructField, options opts) string {
	ignoreEnvPrefix := false
	envVar := flagToEnv(flagName, options.FlagDivider, options.EnvDivider)

	if envTags := strings.Split(field.Tag.Get(scan.DefaultEnvTag), ","); len(envTags) > 0 {
		switch envName := envTags[0]; envName {
		case "-":
			// if tag is `env:"-"` then won't fill flag from environment
			envVar = ""
		case "":
			// if tag is `env:""` then env var will be taken from flag name
		default:
			// if tag is `env:"NAME"` then env var is envPrefix_flagPrefix_NAME
			// if tag is `env:"~NAME"` then env var is NAME
			if strings.HasPrefix(envName, "~") {
				envVar = envName[1:]
				ignoreEnvPrefix = true
			} else {
				envVar = envName
				if options.Prefix != "" {
					envVar = flagToEnv(
						options.Prefix,
						options.FlagDivider,
						options.EnvDivider) + envVar
				}
			}
		}
	}

	if envVar != "" && options.EnvPrefix != "" && !ignoreEnvPrefix {
		envVar = options.EnvPrefix + envVar
	}

	return envVar
}

func setFlagChoices(flag *Flag, choices []string) {
	var allChoices []string

	for _, choice := range choices {
		allChoices = append(allChoices, strings.Split(choice, " ")...)
	}

	flag.Choices = allChoices
}

func hasOption(options []string, option string) bool {
	for _, opt := range options {
		if opt == option {
			return true
		}
	}

	return false
}
