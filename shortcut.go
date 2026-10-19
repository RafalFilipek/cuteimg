package main

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/egoist/mygo"
)

// The capture shortcut can be changed in the editor's settings and is kept
// in settings.json next to the log.

type settings struct {
	Shortcut string `json:"shortcut"`
}

var (
	shortcutMu sync.Mutex
	current    string      // the active shortcut, "" when none could be registered
	testing    atomic.Bool // while true, the shortcut reports itself instead of capturing
)

// ShortcutPressed tells the editor that the shortcut works, while testing.
var ShortcutPressed = mygo.NewEvent[string]("shortcutPressed")

func onShortcut() {
	if testing.Load() {
		_ = ShortcutPressed.Emit(win, current)
		return
	}
	capture()
}

func settingsPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "cuteimg", "settings.json")
}

func loadShortcut() string {
	var s settings
	if b, err := os.ReadFile(settingsPath()); err == nil {
		_ = json.Unmarshal(b, &s)
	}
	if s.Shortcut == "" {
		return defaultShortcut
	}
	return s.Shortcut
}

func saveShortcut(acc string) {
	p := settingsPath()
	if p == "" {
		return
	}
	_ = os.MkdirAll(filepath.Dir(p), 0o755)
	b, _ := json.MarshalIndent(settings{Shortcut: acc}, "", "  ")
	_ = os.WriteFile(p, b, 0o644)
}

// useShortcut replaces the active shortcut with acc. When acc can't be
// registered, the previous one stays active and the error says why.
func useShortcut(acc string) error {
	acc = strings.TrimSpace(acc)
	if err := checkShortcut(acc); err != nil {
		return err
	}
	shortcutMu.Lock()
	defer shortcutMu.Unlock()
	unregisterShortcut()
	if err := registerShortcut(acc); err != nil {
		log.Println("shortcut:", acc, err)
		if current != "" && current != acc {
			_ = registerShortcut(current)
		}
		return errors.New(acc + " can't be used. Another app or the system probably uses it already. Try another combination.")
	}
	current = acc
	log.Println("shortcut:", acc)
	updateTray()
	return nil
}

// checkShortcut asks for a modifier that doesn't just type text.
func checkShortcut(acc string) error {
	parts := strings.Split(strings.ToLower(acc), "+")
	if len(parts) < 2 {
		return errors.New("Add a modifier such as Ctrl, Alt or " + capsName + " to the key.")
	}
	for _, p := range parts[:len(parts)-1] {
		switch p {
		case "ctrl", "control", "alt", "option", "cmd", "command", "super", "win", "meta", "cmdorctrl":
			return nil
		case "capslock":
			if !capsSupported {
				return errors.New("CapsLock shortcuts work only on Windows.")
			}
			return nil
		}
	}
	return errors.New("Shift alone isn't enough. Add Ctrl, Alt or " + capsName + ".")
}

// startShortcut registers the saved shortcut, or the default one when the
// saved one is no longer available.
func startShortcut() {
	if err := useShortcut(loadShortcut()); err != nil && current == "" {
		_ = useShortcut(defaultShortcut)
	}
}

// Shortcut returns the active capture shortcut, or "" when there is none.
func (Editor) Shortcut() string {
	shortcutMu.Lock()
	defer shortcutMu.Unlock()
	return current
}

// DefaultShortcut returns the capture shortcut cuteimg starts with.
func (Editor) DefaultShortcut() string { return defaultShortcut }

// SetShortcut makes acc, such as "Ctrl+Shift+2", the capture shortcut and
// saves it.
func (Editor) SetShortcut(acc string) error {
	if err := useShortcut(acc); err != nil {
		return err
	}
	saveShortcut(acc)
	return nil
}

// TestShortcut turns testing on or off: while on, pressing the shortcut sends
// shortcutPressed instead of capturing the screen.
func (Editor) TestShortcut(on bool) { testing.Store(on) }

// PauseShortcut turns the shortcut off while a new one is recorded, so that
// pressing keys doesn't trigger it.
func (Editor) PauseShortcut(on bool) {
	shortcutMu.Lock()
	defer shortcutMu.Unlock()
	if on {
		unregisterShortcut()
	} else if current != "" {
		_ = registerShortcut(current)
	}
}
