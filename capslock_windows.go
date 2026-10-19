package main

import (
	"errors"
	"log"
	"runtime"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"
)

// Windows has no CapsLock modifier for RegisterHotKey, so CapsLock
// shortcuts such as CapsLock+C come from a low-level keyboard hook. The hook
// takes only the key of the shortcut while CapsLock is held and lets
// everything else through, CapsLock itself included, so tools that use
// CapsLock as a hyper key (Raycast, PowerToys) keep working.

var (
	user32              = syscall.NewLazyDLL("user32.dll")
	kernel32            = syscall.NewLazyDLL("kernel32.dll")
	procSetWindowsHook  = user32.NewProc("SetWindowsHookExW")
	procCallNextHook    = user32.NewProc("CallNextHookEx")
	procGetMessage      = user32.NewProc("GetMessageW")
	procKeybdEvent      = user32.NewProc("keybd_event")
	procGetModuleHandle = kernel32.NewProc("GetModuleHandleW")
	procGetKeyState     = user32.NewProc("GetKeyState")
)

const (
	whKeyboardLL   = 13
	wmKeyDown      = 0x0100
	wmKeyUp        = 0x0101
	wmSysKeyDown   = 0x0104
	wmSysKeyUp     = 0x0105
	vkCapital      = 0x14
	llkhfInjected  = 0x10
	keyeventfKeyUp = 0x0002
)

type kbdLLHookStruct struct {
	VkCode, ScanCode, Flags, Time uint32
	ExtraInfo                     uintptr
}

var capsDown, capsUsed, capsWasOn bool

func capsLockOn() bool {
	r, _, _ := procGetKeyState.Call(vkCapital)
	return r&1 != 0
}

// capsKey is the virtual key that triggers the shortcut together with
// CapsLock, or 0 when no CapsLock shortcut is active.
var capsKey atomic.Uint32

// setCapsKey activates CapsLock+key: a letter, a digit, F1 to F12 or Space.
func setCapsKey(key string) error {
	var vk uint32
	switch k := strings.ToUpper(key); {
	case len(k) == 1 && (k[0] >= 'A' && k[0] <= 'Z' || k[0] >= '0' && k[0] <= '9'):
		vk = uint32(k[0])
	case k == "SPACE":
		vk = 0x20
	case len(k) >= 2 && len(k) <= 3 && k[0] == 'F':
		var n uint32
		for _, c := range k[1:] {
			if c < '0' || c > '9' {
				n = 0
				break
			}
			n = n*10 + uint32(c-'0')
		}
		if n >= 1 && n <= 12 {
			vk = 0x6F + n
		}
	}
	if vk == 0 {
		return errors.New("unsupported key " + key)
	}
	capsKey.Store(vk)
	return nil
}

func keyboardHook(nCode int, wParam, lParam uintptr) uintptr {
	if nCode == 0 {
		k := (*kbdLLHookStruct)(unsafe.Pointer(lParam))
		down := wParam == wmKeyDown || wParam == wmSysKeyDown
		up := wParam == wmKeyUp || wParam == wmSysKeyUp
		if key := capsKey.Load(); key != 0 && k.Flags&llkhfInjected == 0 {
			switch {
			case k.VkCode == vkCapital && down:
				if !capsDown {
					capsDown, capsUsed, capsWasOn = true, false, capsLockOn()
				}
			case k.VkCode == vkCapital && up:
				capsDown = false
				if capsUsed {
					go restoreCapsLock(capsWasOn)
				}
			case capsDown && k.VkCode == key:
				if down && !capsUsed {
					go onShortcut()
				}
				capsUsed = true
				return 1
			}
		}
	}
	r, _, _ := procCallNextHook.Call(0, uintptr(nCode), wParam, lParam)
	return r
}

// restoreCapsLock undoes the Caps Lock toggle that the shortcut's CapsLock
// press caused, when nothing else (such as a hyper key tool) swallowed it.
func restoreCapsLock(wasOn bool) {
	time.Sleep(30 * time.Millisecond)
	if capsLockOn() != wasOn {
		toggleCapsLock()
	}
}

// toggleCapsLock replays a plain CapsLock tap, which the hook ignores
// because it is injected.
func toggleCapsLock() {
	procKeybdEvent.Call(vkCapital, 0x3A, 0, 0)
	procKeybdEvent.Call(vkCapital, 0x3A, keyeventfKeyUp, 0)
}

// installCapsHook runs the hook on its own thread with the message loop
// a low-level hook needs.
func installCapsHook() {
	go func() {
		runtime.LockOSThread()
		mod, _, _ := procGetModuleHandle.Call(0)
		h, _, err := procSetWindowsHook.Call(whKeyboardLL, syscall.NewCallback(keyboardHook), mod, 0)
		if h == 0 {
			log.Println("keyboard hook:", err)
			return
		}
		var msg [64]byte
		for {
			r, _, _ := procGetMessage.Call(uintptr(unsafe.Pointer(&msg[0])), 0, 0, 0)
			if int32(r) <= 0 {
				return
			}
		}
	}()
}
