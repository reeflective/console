package main

type Module string

func (m Module) Enabled() bool {
	return string(m) != ""
}

func (m Module) Template() string {
	return string(m)
}

var module = Module("exploit/multi/handler")
