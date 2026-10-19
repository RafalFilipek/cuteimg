//go:build !windows

package main

import "github.com/egoist/mygo"

const capsSupported = false

func installHook() {}

func registerShortcut(acc string) error { return mygo.GlobalShortcut.Register(acc, onShortcut) }

func unregisterShortcut() { mygo.GlobalShortcut.UnregisterAll() }
