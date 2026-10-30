package main

import (
	"bytes"
	"errors"
	"log"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/egoist/mygo"
)

// restartableCapture is true: the snipping overlay gives no word when the
// user cancels it, so pressing the shortcut again starts a fresh capture
// instead of waiting for the old one.
const restartableCapture = true

var (
	procGetForegroundWindow      = user32.NewProc("GetForegroundWindow")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procOpenProcess              = kernel32.NewProc("OpenProcess")
	procCloseHandle              = kernel32.NewProc("CloseHandle")
	procQueryFullProcessImage    = kernel32.NewProc("QueryFullProcessImageNameW")
)

const processQueryLimitedInformation = 0x1000

// snip opens the Windows snipping overlay, which puts the selection on the
// clipboard, and waits for a new image there. It returns nil when the user
// cancels (the overlay closes without a new image), when stale reports that
// a newer capture took over, or when nothing arrives within two minutes.
func snip(stale func() bool) ([]byte, error) {
	before := mygo.Clipboard.ReadImage()
	time.Sleep(150 * time.Millisecond) // let the editor disappear first
	if err := mygo.Shell.OpenExternal("ms-screenclip:"); err != nil {
		return nil, errors.New("couldn't open the Windows snipping tool: " + err.Error())
	}
	var seen bool      // the overlay has been in front
	var gone time.Time // when it left the front
	deadline := time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) {
		time.Sleep(200 * time.Millisecond)
		if stale() {
			return nil, nil
		}
		if img := mygo.Clipboard.ReadImage(); len(img) > 0 && !bytes.Equal(img, before) {
			return img, nil
		}
		switch exe := foregroundExe(); {
		case isSnipper(exe):
			if !seen {
				log.Println("capture: overlay is", exe)
			}
			seen, gone = true, time.Time{}
		case seen && gone.IsZero():
			gone = time.Now()
		case seen && time.Since(gone) > 1500*time.Millisecond:
			// The overlay closed and no image followed: Esc or its close button.
			return nil, nil
		}
	}
	return nil, nil
}

// isSnipper reports whether exe runs the snipping overlay: Snipping Tool on
// Windows 11, the screen clipping host before it.
func isSnipper(exe string) bool {
	return exe == "snippingtool.exe" || exe == "screenclippinghost.exe"
}

// foregroundExe returns the lower-case file name of the program whose
// window is in front, or "".
func foregroundExe() string {
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return ""
	}
	var pid uint32
	procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	h, _, _ := procOpenProcess.Call(processQueryLimitedInformation, 0, uintptr(pid))
	if h == 0 {
		return ""
	}
	defer procCloseHandle.Call(h)
	buf := make([]uint16, 512)
	n := uint32(len(buf))
	if ok, _, _ := procQueryFullProcessImage.Call(h, 0, uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&n))); ok == 0 {
		return ""
	}
	return strings.ToLower(filepath.Base(syscall.UTF16ToString(buf[:n])))
}
