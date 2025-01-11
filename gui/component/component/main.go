package main

import (
	"flag"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/op"
	"widget.mleku.dev/text"
	"widget.mleku.dev/material"
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

	router := NewRouter()
	router.Register(0, NewAppBarPage(&router))
	router.Register(1, NewNavDrawerPage(&router))
	router.Register(2, NewTextFieldPage(&router))
	router.Register(3, New(&router))
	router.Register(4, NewDiscloserPage(&router))
	router.Register(5, NewAboutPage(&router))

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
