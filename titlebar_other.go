//go:build !windows

package main

import "github.com/egoist/mygo"

// titleBar makes the window frameless: the toolbar drags it and draws the
// window buttons (src/main.ts).
func titleBar(o *mygo.WindowOptions) { o.Frameless = true }
