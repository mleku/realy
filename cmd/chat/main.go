package main

import (
	"gioui.org/app"
	"gioui.org/op"
	"gioui.org/font/gofont"
	"gioui.org/text"
	"gioui.org/widget/material"
	"os"
	"flag"
	"realy.lol/gui/color"
)

func main() {
	flag.Parse()
	go func() {
		w := new(app.Window)
		w.Option(app.MinSize(320, 320))
		w.Option(app.Size(720, 1280))
		w.Option(app.Decorated(false))
		w.Option(app.Title("chat.realy.lol"))
		th := material.NewTheme()
		th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
		col := &color.Palette{Theme: color.GetDefaultTheme()}
		col.Mode = color.Night
		th.Fg = col.GetColor(color.DocText).NRGBA()
		th.Bg = col.GetColor(color.DocBg).NRGBA()
		th.ContrastFg = col.GetColor(color.DocText).NRGBA()
		th.ContrastBg = col.GetColor(color.DocBgHighlight).NRGBA()
		r := new(Root)
		r.Palette = col
		r.th = th
		r.Init()
		if err := loop(w, r); chk.E(err) {
			os.Exit(1)
		}
		os.Exit(0)
	}()
	app.Main()
}

func loop(w *app.Window, r *Root) error {
	var ops op.Ops
	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			g := app.NewContext(&ops, e)
			r.Size = g.Constraints.Max
			r.Layout(g)
			e.Frame(g.Ops)
		}
	}
}
