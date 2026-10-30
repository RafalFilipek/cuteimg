package main

import (
	"time"

	"github.com/egoist/mygo"
)

var (
	procSetForeground = user32.NewProc("SetForegroundWindow")
	procBringToTop    = user32.NewProc("BringWindowToTop")
	procShowWindow    = user32.NewProc("ShowWindow")
	procIsIconic      = user32.NewProc("IsIconic")
)

const (
	vkMenu    = 0x12
	swShow    = 5
	swRestore = 9
)

// raise brings w to the front. Windows lets only the app the user last
// interacted with take the foreground, and after a capture that is the
// snipping tool, so cuteimg would only flash in the taskbar. A synthetic Alt
// tap counts as input to this app and lifts that lock.
func raise(w *mygo.Window) {
	hwnd := w.NativeHandle()
	if hwnd == 0 {
		return
	}
	procKeybdEvent.Call(vkMenu, 0, 0, 0)
	procKeybdEvent.Call(vkMenu, 0, keyeventfKeyUp, 0)
	// SW_RESTORE only when minimized: on a maximized window it would undo
	// the maximize.
	if iconic, _, _ := procIsIconic.Call(hwnd); iconic != 0 {
		procShowWindow.Call(hwnd, swRestore)
	} else {
		procShowWindow.Call(hwnd, swShow)
	}
	procBringToTop.Call(hwnd)
	procSetForeground.Call(hwnd)
	// Topmost for a moment puts it above the snipping tool's leftovers.
	w.SetAlwaysOnTop(true)
	time.AfterFunc(300*time.Millisecond, func() { w.SetAlwaysOnTop(false) })
}
