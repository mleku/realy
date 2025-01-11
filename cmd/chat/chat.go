package main

import (
	"widget.mleku.dev/material"
	"realy.lol/gui/color"
	"realy.lol/gui"
)

type Chat struct {
	r      *Root
	Active st
}

func (c *Chat) Init(r *Root) *Chat {
	c.r = r
	return c
}

func (c *Chat) Layout(g Gx) Dim {
	return c.EmptyMessage(g)
}

func (c *Chat) EmptyMessage(g Gx) (d Dim) {
	l := material.Body1(c.r.th, "select a chat to start messaging")
	p := gui.Pill{c.r.Palette.GetColor(color.PanelBg).NRGBA(), l.TextSize, l.Layout}
	g.Constraints.Min.Y = 0
	d = gui.Centered{}.Layout(g, p.Layout)
	return d
}
