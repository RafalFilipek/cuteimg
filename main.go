package main

import (
	_ "embed"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/egoist/mygo"
)

//go:embed tray.png
var trayIcon []byte

var win *mygo.Window

// Editor is called from the editor page: `mygo generate` turns its methods
// into typed TypeScript functions in src/mygo.ts.
type Editor struct{}

// Copy puts the finished PNG on the clipboard.
func (Editor) Copy(png []byte) error {
	return mygo.Clipboard.WriteImage(png)
}

// Save asks where to save the PNG and writes it. It returns the path, or ""
// when the user cancels.
func (Editor) Save(png []byte, name string) (string, error) {
	dir, _ := mygo.App.Path(mygo.PathPictures)
	path, err := mygo.Dialog.Save(mygo.SaveDialogOptions{
		Parent:      win,
		Title:       "Save image",
		DefaultPath: filepath.Join(dir, name),
		Filters:     []mygo.FileFilter{{Name: "PNG image", Extensions: []string{"png"}}},
	})
	if err != nil || path == "" {
		return "", err
	}
	if !strings.EqualFold(filepath.Ext(path), ".png") {
		path += ".png"
	}
	return path, os.WriteFile(path, png, 0o644)
}

// Capture starts a screen capture, the same as the global shortcut.
func (Editor) Capture() { capture() }

// Shortcut returns the label of the capture shortcut on this system.
func (Editor) Shortcut() string { return shortcutLabel }

// Captured carries a fresh screen capture (PNG) to the editor.
var Captured = mygo.NewEvent[[]byte]("captured")

var capturing atomic.Bool

// capture hides the editor, lets the user select an area of the screen with
// the system's own tool, then shows the result in the editor.
func capture() {
	if !capturing.CompareAndSwap(false, true) {
		return
	}
	go func() {
		defer capturing.Store(false)
		log.Println("capture: start")
		if win != nil {
			win.Hide()
		}
		img, err := snip()
		switch {
		case err != nil:
			log.Println("capture:", err)
			showEditor()
			mygo.Dialog.Error("Screen capture failed", err.Error())
		case len(img) == 0:
			log.Println("capture: canceled")
		default:
			log.Printf("capture: got %d bytes", len(img))
			showEditor()
			_ = Captured.Emit(win, img)
		}
	}()
}

func showEditor() {
	if win == nil || win.IsDestroyed() {
		openEditor()
		return
	}
	win.Show()
	win.Restore()
	win.Focus()
}

var quitting bool

func openEditor() {
	win = mygo.NewWindow(mygo.WindowOptions{
		Title:           "cuteimg",
		Width:           1280,
		Height:          820,
		MinWidth:        720,
		MinHeight:       480,
		BackgroundColor: "light-dark(#eef0f4, #13151b)", // the editor's --bg
		StateKey:        "main",
		URL:             "/",
	})
	// Closing the window keeps cuteimg in the tray, ready for the shortcut.
	win.OnClose(func(e *mygo.CloseEvent) {
		if !quitting {
			e.PreventDefault()
			win.Hide()
		}
	})
}

// openLog writes the log to cuteimg/cuteimg.log in the user's config
// directory (%APPDATA% on Windows), to see why a capture failed.
func openLog() {
	dir, err := os.UserConfigDir()
	if err != nil {
		return
	}
	dir = filepath.Join(dir, "cuteimg")
	_ = os.MkdirAll(dir, 0o755)
	if f, err := os.OpenFile(filepath.Join(dir, "cuteimg.log"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644); err == nil {
		log.SetOutput(f)
	}
}

func main() {
	if !mygo.App.RequestSingleInstanceLock() {
		return
	}
	openLog()
	mygo.Bind(Editor{})

	mygo.App.OnSecondInstance(func([]string, string) { showEditor() })
	mygo.App.OnActivate(func(bool) { showEditor() })
	mygo.App.OnBeforeQuit(func(*mygo.QuitEvent) { quitting = true })
	mygo.App.OnWindowAllClosed(func() {})

	mygo.App.WhenReady(func() {
		openEditor()
		installShortcut()
		_, err := mygo.NewTray(mygo.TrayOptions{
			Icon:    trayIcon,
			ToolTip: "cuteimg: " + shortcutLabel + " captures the screen",
			Menu: mygo.NewMenu([]*mygo.MenuItem{
				{Label: "Capture area\t" + shortcutLabel, Click: func(*mygo.MenuItem, *mygo.Window) { capture() }},
				{Label: "Open editor", Click: func(*mygo.MenuItem, *mygo.Window) { showEditor() }},
				mygo.Separator(),
				{Label: "Quit cuteimg", Click: func(*mygo.MenuItem, *mygo.Window) { quitting = true; mygo.App.Quit() }},
			}),
		})
		if err != nil {
			log.Println("tray:", err)
		}
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
