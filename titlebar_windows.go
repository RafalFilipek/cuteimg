package main

import "github.com/egoist/mygo"

// titleBar gives the window a hidden title bar: the toolbar fills it, and
// Windows keeps its own minimize, maximize and close buttons over the
// toolbar's right end, with snap layouts and a top edge that resizes the
// window. The page makes room for the buttons (src/main.ts).
func titleBar(o *mygo.WindowOptions) {
	o.TitleBarStyle = mygo.TitleBarHidden
	o.TitleBarHeight = toolbarHeight
}

// toolbarHeight is the height of the editor's toolbar, without its bottom
// border, in CSS pixels.
const toolbarHeight = 54
