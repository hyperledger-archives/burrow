package main

import "testing"

func TestBurrow(t *testing.T) {
	app := burrow()
	// Basic smoke test for cli config
	app.Run([]string{"--version"})
	app.Run([]string{"spec"})
	app.Run([]string{"configure"})
	app.Run([]string{"serve"})
}
