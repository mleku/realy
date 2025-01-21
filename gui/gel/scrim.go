package gel

import (
	"widget.mleku.dev"
	col "realy.lol/gui/color"
)

// Scrim implments a clickable translucent overlay. It can animate appearing
// and disappearing as a fade-in, fade-out transition from zero opacity
// to a fixed maximum opacity.
type Scrim struct {
	// FinalAlpha is the final opacity of the scrim on a scale from 0 to 255.
	FinalAlpha uint8
	widget.Clickable
}

// Layout draws the scrim using the provided animation. If the animation indicates
// that the scrim is not visible, this is a no-op.
func (s *Scrim) Layout(g Gx, th *Theme, c *Palette, anim *VisibilityAnimation) Dim {
	return s.Clickable.Layout(g, func(g Gx) Dim {
		if !anim.Visible() {
			return Dim{}
		}
		g.Constraints.Min = g.Constraints.Max
		currentAlpha := s.FinalAlpha
		if anim.Animating() {
			revealed := anim.Revealed(g)
			currentAlpha = uint8(float32(s.FinalAlpha) * revealed)
		}
		color := c.GetColor(col.DocBg).NRGBA(currentAlpha)
		// th.Fg
		// color.A = currentAlpha
		fill := WithAlpha(color, currentAlpha)
		paintRect(g, g.Constraints.Max, fill)
		return Dim{Size: g.Constraints.Max}
	})
}

// ScrimState defines persistent state for a scrim.
type ScrimState struct {
	widget.Clickable
	VisibilityAnimation
}

// ScrimStyle defines how to lay out a scrim.
type ScrimStyle struct {
	*ScrimState
	Color      NRGBA
	FinalAlpha uint8
}

// NewScrim allocates a ScrimStyle.
// Alpha is the final alpha of a fully "appeared" scrim.
func NewScrim(th *Theme, c *Palette, scrim *ScrimState, alpha uint8) ScrimStyle {
	return ScrimStyle{
		ScrimState: scrim,
		Color:      c.GetColor(col.DocBg).NRGBA(),
		// th.Fg,
		FinalAlpha: alpha,
	}
}

func (scrim ScrimStyle) Layout(g Gx) Dim {
	return scrim.Clickable.Layout(g, func(g Gx) Dim {
		if !scrim.Visible() {
			return Dim{}
		}
		g.Constraints.Min = g.Constraints.Max
		alpha := scrim.FinalAlpha
		if scrim.Animating() {
			alpha = uint8(float32(scrim.FinalAlpha) * scrim.Revealed(g))
		}
		return Rect{
			Color: WithAlpha(scrim.Color, alpha),
			Size:  g.Constraints.Max,
		}.Layout(g)
	})
}
