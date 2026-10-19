package main

import (
	"bytes"
	"errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/egoist/mygo"
)

const shortcut = "Ctrl+Shift+2"
const shortcutLabel = "Ctrl+Shift+2"

func installShortcut() {
	if err := mygo.GlobalShortcut.Register(shortcut, capture); err != nil {
		log.Println("shortcut:", err)
	}
}

// snipTools are area screenshot tools, tried in order; the first one that is
// installed takes the capture. {out} is replaced by a temporary file.
var snipTools = [][]string{
	{"sh", "-c", `region=$(slurp) && grim -g "$region" "$1"`, "sh", "{out}"}, // wlroots (Sway, Hyprland)
	{"gnome-screenshot", "-a", "-f", "{out}"},
	{"spectacle", "-r", "-b", "-n", "-o", "{out}"},
	{"flameshot", "gui", "-p", "{out}"},
	{"maim", "-s", "{out}"},
	{"scrot", "-s", "-o", "{out}"},
}

// need says which program has to be installed for each tool.
var need = []string{"slurp", "gnome-screenshot", "spectacle", "flameshot", "maim", "scrot"}

// snip runs the first installed area screenshot tool. It returns nil when
// the user cancels.
func snip() ([]byte, error) {
	time.Sleep(150 * time.Millisecond)
	out := filepath.Join(os.TempDir(), "cuteimg-capture.png")
	for i, tool := range snipTools {
		if _, err := exec.LookPath(need[i]); err != nil {
			continue
		}
		_ = os.Remove(out)
		args := make([]string, len(tool))
		for j, a := range tool {
			args[j] = string(bytes.ReplaceAll([]byte(a), []byte("{out}"), []byte(out)))
		}
		log.Println("capture: using", need[i])
		_ = exec.Command(args[0], args[1:]...).Run() // a canceled selection exits non-zero
		img, err := os.ReadFile(out)
		_ = os.Remove(out)
		if err != nil || len(img) == 0 {
			return nil, nil
		}
		return img, nil
	}
	return nil, errors.New("install one of: grim and slurp, gnome-screenshot, spectacle, flameshot, maim or scrot")
}
