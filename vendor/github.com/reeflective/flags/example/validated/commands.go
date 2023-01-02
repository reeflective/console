package validated

import (
	"fmt"
)

// Commands is a struct acting as a group of commands.
// All of these make use of tag directives to provide
// validation to their positional args or flags.
type Commands struct {
	ValidatedArgs  `command:"valid-positionals" desc:"Positional arguments validated with struct tag directives"`
	ValidatedFlags `command:"valid-flags" desc:"Flags validated with struct tag directives"`
}

// ValidatedArgs is a command whose positionals are being validated.
type ValidatedArgs struct {
	Args struct {
		IP     string   `description:"An IPv4 address" validate:"ipv4"`
		Emails []string `description:"A list of email addresses" required:"1-2" validate:"email"`
	} `positional-args:"yes"`
}

func (c *ValidatedArgs) Execute(args []string) error {
	fmt.Printf("IP (string):        %v\n", c.Args.IP)
	fmt.Printf("Emails ([]string):   %v\n", c.Args.Emails)

	return nil
}

// ValidatedArgs is a command whose flags' arguments are being validated.
type ValidatedFlags struct {
	Path         string            `short:"p" long:"path" description:"A valid path on your system" complete:"Files" validate:"file"`
	OptionalPath string            `short:"d" long:"opt-dir" description:"A validated directory on your system, with a default value" optional-value:"/home/user" complete:"Files" validate:"dir"`
	Files        []string          `short:"f" long:"files" desc:"A list of files, each validated" complete:"Files" validate:"file"`
	Elems        map[string]string `short:"e" long:"elems" description:"A map[string]string flag, can be repeated or used with comma-separated items" choice:"user:host machine:testing another:target"`
}

func (c *ValidatedFlags) Execute(args []string) error {
	fmt.Printf("Path (string):               %v\n", c.Path)
	fmt.Printf("OptionalPath (string):       %v\n", c.OptionalPath)
	fmt.Printf("Files ([]string):            %v\n", c.Files)
	fmt.Printf("Elems (map[string]string):   %v\n", c.Elems)

	return nil
}
