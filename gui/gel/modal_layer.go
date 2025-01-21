package gel

import (
	"time"

	"gioui.org/layout"
	"gioui.org/op"
)

// ModalLayer is a widget drawn on top of the normal UI that can be populated
// by other material components with dismissble modal dialogs. For instance,
// the App Bar can render its overflow menu within the modal layer, and the
// modal navigation drawer is entirely within the modal layer.
type ModalLayer struct {
	VisibilityAnimation
	Scrim
	Widget func(g Gx, th *Theme, c *Palette, anim *VisibilityAnimation) Dim
}

const defaultModalAnimationDuration = time.Millisecond * 250

// NewModal creates an initializes a modal layer.
func NewModal() *ModalLayer {
	m := ModalLayer{}
	m.VisibilityAnimation.State = Invisible
	m.VisibilityAnimation.Duration = defaultModalAnimationDuration
	m.Scrim.FinalAlpha = 192 // default
	return &m
}

// Layout renders the modal layer. Unless a modal widget has been triggered,
// this will do nothing.
func (m *ModalLayer) Layout(g Gx, th *Theme, c *Palette) Dim {
	if !m.Visible() {
		return Dim{}
	}
	if m.Scrim.Clicked(g) {
		m.Disappear(g.Now)
	}
	scrimDims := m.Scrim.Layout(g, th, c, &m.VisibilityAnimation)
	if m.Widget != nil {
		_ = m.Widget(g, th, c, &m.VisibilityAnimation)
	}
	return scrimDims
}

// ModalState defines persistent state for a modal.
type ModalState struct {
	ScrimState
	// content is the content widget to layout atop a scrim.
	// This is specified as a field because where the content is defined
	// is not where it is invoked.
	// Thus, the content widget becomes the state of the modal.
	content layout.Widget
}

// ModalStyle describes how to lay out a modal.
// Modal content is layed centered atop a clickable scrim.
type ModalStyle struct {
	*ModalState
	Scrim ScrimStyle
}

// Modal lays out a content widget atop a clickable scrim.
// Clicking the scrim dismisses the modal.
func Modal(th *Theme, c *Palette, modal *ModalState) ModalStyle {
	return ModalStyle{
		ModalState: modal,
		Scrim:      NewScrim(th, c, &modal.ScrimState, 250),
	}
}

// Layout the scrim and content. The content is only laid out once
// the scrim is fully animated in, and is hidden on the first frame
// of the scrim's fade-out animation.
func (m ModalStyle) Layout(g Gx) Dim {
	if m.content == nil || !m.Visible() {
		return Dim{}
	}
	if m.Clicked(g) {
		m.Disappear(g.Now)
	}
	macro := op.Record(g.Ops)
	dims := layout.Stack{}.Layout(
		g,
		layout.Expanded(func(g Gx) Dim {
			return m.Scrim.Layout(g)
		}),
		layout.Expanded(func(g Gx) Dim {
			if m.Scrim.Visible() && !m.Scrim.Animating() {
				return m.content(g)
			}
			return Dim{}
		}),
	)
	op.Defer(g.Ops, macro.Stop())
	return dims
}

// Show widget w in the modal, starting animation at now.
func (m *ModalState) Show(now time.Time, w layout.Widget) {
	m.content = w
	m.Appear(now)
}
