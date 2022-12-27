package console

import (
	"fmt"
	"oh-my-posh/engine"
	"oh-my-posh/platform"
	"oh-my-posh/properties"
)

// Prompt wraps an oh-my-posh prompt engine, so as to be able
// to be configured/enhanced and used the same way oh-my-posh is.
// Some methods have been added for ordering the application to
// to recompute prompts, print logs in sync with them, etc.
type Prompt struct {
	*engine.Engine
	console *Console
}

// Segment represents a type able to render itself as a prompt segment string.
// Any number of segments can be registered to the prompt engine, and those
// segments can then be used by declaration in the prompt configuration file.
type Segment interface {
	Enabled() bool
	Template() string
}

// AddSegment enables to register a prompt segment to the prompt engine.
// This segment can then be configured and used in the prompt configuration file.
func (p *Prompt) AddSegment(name string, prompt Segment) {
	if p.Engine == nil {
		p.Engine = newDefaultEngine()
	}

	if prompt == nil {
		return
	}

	segment := &segment{
		Segment: prompt,
	}

	p.Engine.AddSegment(engine.SegmentType(name), segment)
}

// LogTransient prints a string message (a log, or more broadly, an
// asynchronous event) without bothering the user, and by "pushing"
// the prompt below the message.
//
// If this function is called while a command is running, the console
// will simply print the log below the current line, and will not print
// the prompt. In any other case this function will work normally.
func (c *Console) LogTransient(msg string, args ...interface{}) {
	if c.isExecuting {
		fmt.Printf(msg, args...)
	} else {
		c.shell.LogTransient(msg, args...)
	}
}

// Log - A simple function to print a message and redisplay the prompt below it.
// As with LogTransient, if this function is called while a command is running,
// the console will simply print the log below the current line, and will not
// print the prompt. In any other case this function will work normally.
func (c *Console) Log(msg string, args ...interface{}) {
	if c.isExecuting {
		fmt.Printf(msg, args...)
	} else {
		c.shell.Log(msg, args...)
	}
}

// segment encapsulates a user-defined prompt segment and redeclares
// the methods required to be considered a valid prompt by oh-my-posh.
type segment struct {
	props properties.Properties
	env   platform.Environment
	Segment
}

// Init implements the engine.SegmentWriter interface.
func (s *segment) Init(props properties.Properties, env platform.Environment) {
	s.props = props
	s.env = env
}

// newPrompt initializes a prompt system/engine for the given menu,
// loading any configuration that is relevant to it.
func newPrompt(console *Console) *Prompt {
	flags := &platform.Flags{
		Shell: "plain",
	}

	p := &Prompt{
		Engine:  engine.New(flags),
		console: console,
	}

	return p
}

// makes a prompt engine with default/builtin configuration.
func newDefaultEngine() *engine.Engine {
	flags := &platform.Flags{
		Shell: "plain",
	}

	return engine.New(flags)
}
