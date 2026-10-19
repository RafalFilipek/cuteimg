# cuteimg

Beautiful screenshots on Windows, macOS and Linux. Capture or paste a screenshot, put it on a gradient background, mark it up with arrows, rectangles and circles, pixelate anything private, then copy or save it. A CleanShot X–style editor built with [MyGo](https://mygo.egoist.dev).

![cuteimg editor](docs/screenshot.png)

## Features

- **Capture from anywhere** with a global shortcut. cuteimg hides, lets you select an area with the system's own tool, and opens the result in the editor.
- **Paste** an image with Ctrl+V, open a file, or drag one onto the window.
- **Backgrounds:** mesh gradients (Aurora, Lava, Candy, Lagoon, Dusk, Neon, Holo, Forest, Peach, Ice), linear gradients, a rainbow, white, a custom color or none, with optional film grain.
- **Frame:** padding, corner radius and drop shadow.
- **Annotations:** arrows, rectangles and ellipses in eight colors and three thicknesses. Hold Shift for a square, a circle or a 45° arrow. Select, move, resize and delete them.
- **Pixelate** an area to hide emails, names or tokens, with an adjustable block size.
- **Undo and redo** for every change.
- **Copy** the result to the clipboard (Ctrl+C) or **save** a PNG (Ctrl+S).
- Lives in the system tray, so closing the window keeps the shortcut working.

## Shortcuts

| Action | Windows | macOS | Linux |
|---|---|---|---|
| Capture an area (global) | CapsLock+C | ⌘⇧2 | Ctrl+Shift+2 |
| Paste an image | Ctrl+V | ⌘V | Ctrl+V |
| Select / Arrow / Rectangle / Ellipse / Pixelate | V / A / R / O / P | same | same |
| Undo / Redo | Ctrl+Z / Ctrl+Y | ⌘Z / ⌘⇧Z | Ctrl+Z / Ctrl+Y |
| Delete the selected shape | Delete | Delete | Delete |
| Copy the result | Ctrl+C | ⌘C | Ctrl+C |
| Save a PNG | Ctrl+S | ⌘S | Ctrl+S |

On Windows, CapsLock works as a modifier while you hold it. Tapping it on its own still toggles Caps Lock.

## How capture works

| System | Tool |
|---|---|
| Windows | The built-in Snipping overlay (`ms-screenclip:`). cuteimg picks up the selection from the clipboard. |
| macOS | `screencapture -i`. The first capture asks for the Screen Recording permission. |
| Linux | The first one installed of `grim` + `slurp`, `gnome-screenshot`, `spectacle`, `flameshot`, `maim` or `scrot`. |

If a capture fails, the log at `%APPDATA%\cuteimg\cuteimg.log` (Windows), `~/Library/Application Support/cuteimg/cuteimg.log` (macOS) or `~/.config/cuteimg/cuteimg.log` (Linux) says why.

## Install

Download the latest build from [Releases](https://github.com/RafalFilipek/cuteimg/releases):

- **Windows:** `cuteimg Setup <version> amd64.exe`, or the portable `cuteimg.exe`. The builds are not signed yet, so SmartScreen warns you: choose **More info**, then **Run anyway**.
- **macOS:** the `.dmg`. The app is not notarized yet, so open it the first time with right-click → **Open**.
- **Linux:** the `.deb`, or the `.tar.gz` with its `install.sh`.

## Build from source

You need [Go](https://go.dev) and [Bun](https://bun.sh).

```sh
bun install
bun run dev        # run with live reload
bun run build      # package for this system
bun run build -- -platform windows/amd64,linux/amd64,darwin/universal
```

MyGo needs no cgo, so any machine builds for every platform. The Windows installer needs NSIS (`apt install nsis` on Linux), and the macOS disk image needs a Mac.

## Project layout

| Path | What it is |
|---|---|
| `index.html` | The editor: canvas rendering, tools, backgrounds and export, in plain JavaScript |
| `src/main.ts` | Connects the editor to Go: copy, save and captured images |
| `main.go` | The app: window, tray, clipboard, save dialog |
| `capture_*.go` | Screen capture and the global shortcut for each system |
| `capslock_windows.go` | The CapsLock+C keyboard hook on Windows |

## Releases

Push a tag that matches `version` in `mygo.config.ts`, for example `git tag v0.1.0 && git push origin v0.1.0`. The release workflow builds every platform and attaches the files to a draft release.
