package main

import (
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

type Module struct {
	Type string
	Path string
}

func (m Module) Enabled() bool {
	return string(m.Path) != ""
}

func (m Module) Template() string {
	return string(m.Path)
}

func (m Module) Init(props properties.Properties, env platform.Environment) {
}

var module = Module{"scan", "protocol/tcp"}
