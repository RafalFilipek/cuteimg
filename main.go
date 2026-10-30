package main

import (
	_ "embed"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/plugins/updater"
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

// Captured tells the editor that a fresh screen capture is waiting; the
// editor fetches it with TakeCapture.
var Captured = mygo.NewEvent[bool]("captured")

var (
	capturing atomic.Bool
	// captureGen numbers captures, so a capture started while another one
	// still waits takes over and the older one gives up.
	captureGen atomic.Int64
	// reshow brings the editor back when a capture started from it is
	// canceled.
	reshow atomic.Bool
)

// pending holds the latest capture until the editor takes it. A window that
// was just shown can miss the event or not repaint until it is clicked, so
// the editor also asks for it whenever it gets focus or becomes visible.
var (
	pendingMu sync.Mutex
	pending   []byte
)

// TakeCapture returns the capture waiting for the editor (PNG), once, or nothing.
func (Editor) TakeCapture() []byte {
	pendingMu.Lock()
	defer pendingMu.Unlock()
	img := pending
	pending = nil
	return img
}

// capture hides the editor, lets the user select an area of the screen with
// the system's own tool, then shows the result in the editor.
func capture() {
	if !capturing.CompareAndSwap(false, true) && !restartableCapture {
		return
	}
	gen := captureGen.Add(1)
	stale := func() bool { return captureGen.Load() != gen }
	go func() {
		log.Println("capture: start")
		if win != nil && !win.IsDestroyed() && win.IsVisible() {
			reshow.Store(true)
			win.Hide()
		}
		img, err := snip(stale)
		if stale() {
			return // a newer capture took over
		}
		defer capturing.Store(false)
		switch {
		case err != nil:
			log.Println("capture:", err)
			showEditor()
			mygo.Dialog.Error("Screen capture failed", err.Error())
		case len(img) == 0:
			log.Println("capture: canceled")
			if reshow.Swap(false) {
				showEditor()
			}
		default:
			reshow.Store(false)
			log.Printf("capture: got %d bytes", len(img))
			pendingMu.Lock()
			pending = img
			pendingMu.Unlock()
			showEditor()
			// Give the webview a moment to become visible before nudging it.
			time.Sleep(150 * time.Millisecond)
			_ = Captured.Emit(win, true)
		}
	}()
}

func showEditor() {
	if win == nil || win.IsDestroyed() {
		openEditor()
		return
	}
	win.Show()
	// Restore only a minimized window: on a maximized one it would undo the
	// maximize.
	if win.IsMinimized() {
		win.Restore()
	}
	raise(win)
	win.Focus()
}

var quitting bool

var tray *mygo.Tray

// updateTray shows the current shortcut in the tray's menu and tooltip.
func updateTray() {
	if tray == nil {
		return
	}
	label := "Capture area"
	tip := "cuteimg"
	if current != "" {
		label += "\t" + current
		tip += ": " + current + " captures the screen"
	}
	tray.SetToolTip(tip)
	tray.SetMenu(mygo.NewMenu([]*mygo.MenuItem{
		{Label: label, Click: func(*mygo.MenuItem, *mygo.Window) { capture() }},
		{Label: "Open editor", Click: func(*mygo.MenuItem, *mygo.Window) { showEditor() }},
		updater.MenuItem(),
		mygo.Separator(),
		{Label: "Quit cuteimg", Click: func(*mygo.MenuItem, *mygo.Window) { quitting = true; mygo.App.Quit() }},
	}))
}

func openEditor() {
	opts := mygo.WindowOptions{
		Title:           "cuteimg",
		Width:           1280,
		Height:          820,
		MinWidth:        980,
		MinHeight:       480,
		BackgroundColor: "light-dark(#eef0f4, #13151b)", // the editor's --bg
		StateKey:        "main",
		URL:             "/",
	}
	// The toolbar is the title bar: it drags the window, and holds the
	// window buttons (see titlebar_*.go and src/main.ts).
	titleBar(&opts)
	win = mygo.NewWindow(opts)
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
	mygo.Use(updater.Plugin)
	mygo.Bind(Editor{})

	mygo.App.OnSecondInstance(func([]string, string) { showEditor() })
	mygo.App.OnActivate(func(bool) { showEditor() })
	mygo.App.OnBeforeQuit(func(*mygo.QuitEvent) { quitting = true })
	mygo.App.OnWindowAllClosed(func() {})

	mygo.App.WhenReady(func() {
		openEditor()
		installHook()
		var err error
		tray, err = mygo.NewTray(mygo.TrayOptions{Icon: trayIcon, ToolTip: "cuteimg"})
		if err != nil {
			log.Println("tray:", err)
		} else {
			// A left click opens the editor; a right click shows the menu.
			tray.OnClick(showEditor)
		}
		startShortcut()
		updateTray()
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
