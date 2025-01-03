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
	// Mock content
	mock := chat.Relay{
		Chats: []chat.ChatButton{
			{"@DoctorBreen", "A Message From Our Benefactors", nil},
			{"@GMan", "Rise and Shine. Mister Freeman", nil},
			{"@DoctorKleiner", "She's debeaked and completely harmless", nil},
			{"@Barney", "I still have nightmares about that cat", nil},
			{"@Alyx", "Man of few words, aren't you? :blush:", nil},
			{"#general", "CPs at Delta Niner", nil},
			{"#combine", "Alert! Alert! Unrest structure detected in sector Bravo Six", nil},
		},
	}
	s.Root.Panel.S = append(s.Root.Panel.S, mock)
	go func() {
		w := new(app.Window)
		w.Option(app.MinSize(320, 320))
		w.Option(app.Size(720, 1280))
		w.Option(app.Decorated(true))
		w.Option(app.Title("nostr realy chat"))
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
