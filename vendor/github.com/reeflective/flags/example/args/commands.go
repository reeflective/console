package args

import (
	"fmt"
)

//
// This file contains all subcommands meant to demonstrate how to use positional arguments.
//

// MultipleListsArgs shows how to declare multiple lists as positional arguments,
// and how provided arguments are dispatched onto their slots.
type MultipleListsArgs struct {
	Args struct {
		// The Vuln positional slot is of type IP, which is an aliased []string.
		// This is so that we can implement a completer around this type.
		//
		// Notice, here, that there is no individual min/max requirement for
		// this slot, in additional to no global requirement on the struct itself.
		// Thus, this positional slot can be empty.
		Vuln IP `description:"Vulnerable IP addresses to check" validate:"ipv4"`

		// Other is a positional slot requiring at least one element, and at most two.
		// Consequently:
		// - If one argument is passed, it will be stored in this slot.
		// - If two arguments, one is stored here, and the first is stored in Vuln
		// - If three or more, two will be stored here, and the others in Vuln.
		Other []Host `description:"Other list of IP addresses" required:"1-2"`
	} `positional-args:"yes"`
}

// Execute - Note that the args []string parameter will ALWAYS be empty,
// since all positional slots of the commands are lists, and at least one
// of them has no maximum number of arguments.
func (c *MultipleListsArgs) Execute(args []string) error {
	fmt.Printf("Vuln (IP):        %v\n", c.Args.Vuln)
	fmt.Printf("Other ([]Host):   %v\n", c.Args.Other)

	return nil
}

// FirstListArgs shows how to use several positionals, of which the first is a list, but not the last.
type FirstListArgs struct {
	Args struct {
		Hosts  []Host `description:"A list of hosts with minimum and maximum requirements" required:"1-2"`
		Target Proxy  `description:"A single, required remaining argument" required:"1"`
	} `positional-args:"yes" required:"yes"`
}

// Execute - Since the positional arguments for this command all have a maximum allowed number
// of items, any word given in excess will be stored in the args []string parameter of this function.
func (c *FirstListArgs) Execute(args []string) error {
	fmt.Printf("Hosts ([]Host):   %v\n", c.Args.Hosts)
	fmt.Printf("Target (Proxy):   %v\n", c.Args.Target)

	if len(args) > 0 {
		fmt.Printf("Remaining args: %v\n", args)
	}

	return nil
}

// MultipleMinMaxArgs shows how to use multiple lists as positionals, with overlapping min/max requirements.
// Note that here, the two first slots (Hosts and Proxies), have "overlapping" requirements:
// - If 2 args are given, each will get one
// - If 3 args, Hosts will get 2, and proxies will get 1
// - If 4 args, each will get 2
//
// Since the IP slot is also a list, all arguments in excess (here, if more than 5 args) will be stored in it.
type MultipleMinMaxArgs struct {
	Args struct {
		Hosts     []Host  `description:"A list of hosts with minimum and maximum requirements" required:"1-2"`
		Proxies   []Proxy `description:"A list of proxies, with min/max requirements overlapping with Hosts" required:"1-2"`
		Addresses IP      `description:"A last list of IPs, which will store any words given in excess" required:"1"`
	} `positional-args:"yes" required:"yes"`
}

// Execute - Note that the args []string parameter will ALWAYS be empty,
// since all positional slots of the commands are lists, and at least one
// of them has no maximum number of arguments.
func (c *MultipleMinMaxArgs) Execute(args []string) error {
	fmt.Printf("Hosts ([]Host):      %v\n", c.Args.Hosts)
	fmt.Printf("Proxies ([]Proxy):   %v\n", c.Args.Proxies)
	fmt.Printf("Addresses (IP):      %v\n", c.Args.Addresses)

	return nil
}

// TagCompletedArgs shows how to specify completers with struct tags.
type TagCompletedArgs struct {
	Args struct {
		// Files accepts at most two values, and the completions for them will be restricted to files with a '.go' extension.
		// Since this slot has a min and max requirement value, once one argument is provided at the command-line, completions
		// will be proposed both for this slot and for the next one, up until the maximum requirements are fulfilled.
		Files []string `description:"A list of files with min/max requirements" required:"1-2" complete:"FilterExt,go"`

		// JSONConfig also completes files, but with a .json extension.
		JSONConfig string `description:"the target of your command (anything string-based)" required:"1" complete:"FilterExt,json"`
	} `positional-args:"yes" required:"yes"`
}

// Execute - Here, since the last positional slot is not a list,
// and the first one is a list but has a maximum number of arguments
// allowed, any arg in excess is stored in args []string parameter.
func (c *TagCompletedArgs) Execute(args []string) error {
	fmt.Printf("Files ([]string):       %v\n", c.Args.Files)
	fmt.Printf("JsonConfig (string):    %v\n", c.Args.JSONConfig)

	if len(args) > 0 {
		fmt.Printf("Remaining args: %v\n", args)
	}

	return nil
}
