package gomod

type Module struct {
	Path    string
	Version string
}

type GoMod struct {
	Module  Module
	Require []Require
	Exclude []Module
	Replace []Replace
}

type Require struct {
	Path     string
	Version  string
	Indirect bool
}

type Replace struct {
	Old Module
	New Module
}
