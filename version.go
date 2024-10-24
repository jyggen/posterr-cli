package main

import (
	"github.com/alecthomas/kong"
)

var version = ""

type VersionFlag string

func (v VersionFlag) Decode(_ *kong.DecodeContext) error { return nil }
func (v VersionFlag) IsBool() bool                       { return true }
func (v VersionFlag) BeforeApply(app *kong.Kong, _ kong.Vars) error {
	app.Printf("%s", version)
	app.Exit(0)
	return nil
}
