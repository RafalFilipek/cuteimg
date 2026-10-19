package main

import (
	"strings"

	"github.com/egoist/mygo"
)

const defaultShortcut = "CapsLock+C"
const capsSupported = true
const capsName = "CapsLock"

func installHook() { installCapsHook() }

// registerShortcut sends CapsLock combinations to the keyboard hook, which
// Windows needs for them, and the rest to RegisterHotKey.
func registerShortcut(acc string) error {
	if key, ok := strings.CutPrefix(strings.ToLower(acc), "capslock+"); ok {
		return setCapsKey(key)
	}
	return mygo.GlobalShortcut.Register(acc, onShortcut)
}

func unregisterShortcut() {
	capsKey.Store(0)
	mygo.GlobalShortcut.UnregisterAll()
}
