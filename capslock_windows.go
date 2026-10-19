package main

import (
	"log"
	"runtime"
	"syscall"
	"unsafe"
)

// Windows has no CapsLock modifier for RegisterHotKey, so CapsLock+C comes
// from a low-level keyboard hook. CapsLock held with another key works as a
// modifier; tapped alone it still toggles Caps Lock.

var (
	user32              = syscall.NewLazyDLL("user32.dll")
	kernel32            = syscall.NewLazyDLL("kernel32.dll")
	procSetWindowsHook  = user32.NewProc("SetWindowsHookExW")
	procCallNextHook    = user32.NewProc("CallNextHookEx")
	procGetMessage      = user32.NewProc("GetMessageW")
	procKeybdEvent      = user32.NewProc("keybd_event")
	procGetModuleHandle = kernel32.NewProc("GetModuleHandleW")
)

const (
	whKeyboardLL   = 13
	wmKeyDown      = 0x0100
	wmKeyUp        = 0x0101
	wmSysKeyDown   = 0x0104
	wmSysKeyUp     = 0x0105
	vkCapital      = 0x14
	vkC            = 0x43
	llkhfInjected  = 0x10
	keyeventfKeyUp = 0x0002
)

type kbdLLHookStruct struct {
	VkCode, ScanCode, Flags, Time uint32
	ExtraInfo                     uintptr
}

var capsDown, capsUsed bool

func keyboardHook(nCode int, wParam, lParam uintptr) uintptr {
	if nCode == 0 {
		k := (*kbdLLHookStruct)(unsafe.Pointer(lParam))
		down := wParam == wmKeyDown || wParam == wmSysKeyDown
		up := wParam == wmKeyUp || wParam == wmSysKeyUp
		if k.Flags&llkhfInjected == 0 {
			switch {
			case k.VkCode == vkCapital && down:
				if !capsDown {
					capsDown, capsUsed = true, false
				}
				return 1
			case k.VkCode == vkCapital && up:
				capsDown = false
				if !capsUsed {
					go toggleCapsLock()
				}
				return 1
			case capsDown && k.VkCode == vkC:
				if down && !capsUsed {
					log.Println("CapsLock+C")
					capture()
				}
				capsUsed = true
				return 1
			case capsDown && down:
				capsUsed = true
			}
		}
	}
	r, _, _ := procCallNextHook.Call(0, uintptr(nCode), wParam, lParam)
	return r
}

// toggleCapsLock replays a plain CapsLock tap, which the hook lets through
// because it is injected.
func toggleCapsLock() {
	procKeybdEvent.Call(vkCapital, 0x3A, 0, 0)
	procKeybdEvent.Call(vkCapital, 0x3A, keyeventfKeyUp, 0)
}

// installCapsShortcut runs the hook on its own thread with the message loop
// a low-level hook needs.
func installCapsShortcut() {
	go func() {
		runtime.LockOSThread()
		mod, _, _ := procGetModuleHandle.Call(0)
		h, _, err := procSetWindowsHook.Call(whKeyboardLL, syscall.NewCallback(keyboardHook), mod, 0)
		if h == 0 {
			log.Println("keyboard hook:", err)
			return
		}
		log.Println("CapsLock+C ready")
		var msg [64]byte
		for {
			r, _, _ := procGetMessage.Call(uintptr(unsafe.Pointer(&msg[0])), 0, 0, 0)
			if int32(r) <= 0 {
				return
			}
		}
	}()
}
