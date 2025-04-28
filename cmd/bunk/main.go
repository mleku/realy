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

	"realy.lol/cmd/bunk/pages"
	page "realy.lol/cmd/bunk/router"
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
	r := page.NewRouter()
	r.Register(0, pages.NewSetup(r))
	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			r.Layout(gtx, th)
			e.Frame(gtx.Ops)
		}
	}
}
