package main

import (
	"flag"
	"log"
	"os"

	"gio.mleku.dev/gio/app"
	"gio.mleku.dev/gio/font/gofont"
	"gio.mleku.dev/gio/layout"
	"gio.mleku.dev/gio/op"
	"gio.mleku.dev/gio/text"
	"gio.mleku.dev/gio/widget/material"

	page "realy.mleku.dev/gui/pages"

	"realy.mleku.dev/gui/pages/appbar"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

func main() {
	flag.Parse()
	go func() {
		w := new(app.Window)
		if err := loop(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func loop(w *app.Window) error {
	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
	var ops op.Ops

	router := page.NewRouter()
	router.Register(0, appbar.New(&router))
	// router.Register(1, navdrawer.New(&router))
	// router.Register(2, textfield.New(&router))
	// router.Register(3, menu.New(&router))
	// router.Register(4, discloser.New(&router))
	// router.Register(5, about.New(&router))

	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			router.Layout(gtx, th)
			e.Frame(gtx.Ops)
		}
	}
}
