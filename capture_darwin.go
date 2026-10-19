package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const defaultShortcut = "Cmd+Shift+2"
const capsName = "Cmd"

// snip runs the interactive macOS screenshot tool. It returns nil when the
// user presses Esc.
func snip() ([]byte, error) {
	time.Sleep(150 * time.Millisecond)
	out := filepath.Join(os.TempDir(), "cuteimg-capture.png")
	_ = os.Remove(out)
	if err := exec.Command("/usr/sbin/screencapture", "-i", "-x", out).Run(); err != nil {
		return nil, err
	}
	defer os.Remove(out)
	img, err := os.ReadFile(out)
	if os.IsNotExist(err) {
		return nil, nil
	}
	return img, err
}
