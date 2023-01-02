package main

type Module string

func (m Module) Enabled() bool {
	return string(m) != ""
}

func (m Module) Template() string {
	return string(m)
}

var module Module = Module("exploit/multi/handler")

// Add prompt segment
// menu.Prompt().AddSegment("module", module)

// Load a custom prompt configuration
// menu.Prompt().LoadConfig("/home/user/code/github.com/reeflective/oh-my-posh/test-prompt.omp.json")
