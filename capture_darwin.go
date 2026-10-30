package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const defaultShortcut = "Cmd+Shift+2"
const capsName = "Cmd"

// restartableCapture is false: the capture tool here ends on its own when
// the user cancels, and a second one must not start meanwhile.
const restartableCapture = false

// snip runs the interactive macOS screenshot tool. It returns nil when the
// user presses Esc.
func snip(stale func() bool) ([]byte, error) {
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
