package main

import (
	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/widget/material"
	"realy.lol/chat"
	"realy.lol/gui/color"
	"os"
)

type State struct {
	*chat.Root
}

func (s *State) Init(th *Theme, col *color.Palette) *State {
	s.Root = new(chat.Root).Init(th, col)
	return s
}

func (s *State) Layout(g Gx) Dim {
	s.Size = g.Constraints.Max
	return s.Root.Layout(g)
}

func main() {
	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
	col := &color.Palette{Theme: color.GetDefaultTheme()}
	col.Mode = color.Night
	th.Fg = col.GetColor(color.DocText).NRGBA()
	th.Bg = col.GetColor(color.DocBg).NRGBA()
	th.ContrastFg = col.GetColor(color.DocTextDim).NRGBA()
	th.ContrastBg = col.GetColor(color.DocBgDim).NRGBA()
	s := new(State).Init(th, col)
	go func() {
		w := new(app.Window)
		w.Option(app.MinSize(320, 320))
		w.Option(app.Size(720, 1280))
		w.Option(app.Decorated(false))
		if err := loop(w, s); chk.E(err) {
			os.Exit(1)
		}
		os.Exit(0)
	}()
	app.Main()
}

func loop(w *app.Window, s *State) (err er) {
	var ops op.Ops
	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			g := app.NewContext(&ops, e)
			s.Layout(g)
			e.Frame(g.Ops)
		}
	}
}
