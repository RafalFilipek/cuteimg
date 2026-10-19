//go:build !windows

package main

import "github.com/egoist/mygo"

// raise needs nothing beyond Focus outside Windows.
func raise(*mygo.Window) {}
