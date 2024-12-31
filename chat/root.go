package chat

import (
	"realy.lol/atomic"
	"realy.lol/gui/color"
	"sync"
)

// Root is a widget that has a sidebar menu that becomes the whole display when
// on a small display, on large, it has a second panel that becomes the main
// display, when in small, it replaces the menu with the view widget and a top
// bar that contains a back button added to the left of the view widget's
// header.
type Root struct {
	sync.Mutex
	th *Theme
	*color.Palette
	Size Point
	*Panel
	*Chat
	Pages  map[st]Widget
	Active atomic.String
}

func (r *Root) Init(th *Theme, col *color.Palette) *Root {
	r.th = th
	r.Palette = col
	r.Panel = new(Panel).Init(r)
	r.Chat = new(Chat).Init(r)
	return r
}

func (r *Root) Layout(g Gx) Dim {
	Fill(g.Ops, r.GetColor(color.DocBg).NRGBA())
	flex := Flex{Axis: Horizontal}
	if r.Size.X < 720 {
		// at small sizes, only render panel
		r.Panel.Small = true
		return r.Panel.Layout(g)
	}
	// // at larger sizes, rigid render panel at 360px wide
	r.Panel.Small = false
	return flex.Layout(g, Rigid(r.Panel.Layout), Rigid(r.Chat.Layout))

}
