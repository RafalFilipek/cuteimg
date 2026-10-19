package main

import (
	"bytes"
	"errors"
	"time"

	"github.com/egoist/mygo"
)

const shortcutLabel = "CapsLock+C"

func installShortcut() { installCapsShortcut() }

// snip opens the Windows snipping overlay, which puts the selection on the
// clipboard, and waits for a new image there. It returns nil when the user
// cancels or nothing arrives within two minutes.
func snip() ([]byte, error) {
	before := mygo.Clipboard.ReadImage()
	time.Sleep(150 * time.Millisecond) // let the editor disappear first
	if err := mygo.Shell.OpenExternal("ms-screenclip:"); err != nil {
		return nil, errors.New("couldn't open the Windows snipping tool: " + err.Error())
	}
	deadline := time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) {
		time.Sleep(250 * time.Millisecond)
		if img := mygo.Clipboard.ReadImage(); len(img) > 0 && !bytes.Equal(img, before) {
			return img, nil
		}
	}
	return nil, nil
}
