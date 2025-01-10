package gel

import (
	"image"
	"time"

	"gioui.org/f32"
	"gioui.org/gesture"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"

	col "realy.lol/gui/color"
)

// Sheet implements the standard side sheet described here:
// https://material.io/components/sheets-side#usage
type Sheet struct{}

// NewSheet returns a new sheet
func NewSheet() Sheet {
	return Sheet{}
}

// Layout renders the provided widget on a background. The background will use
// the maximum space available.
func (s Sheet) Layout(g Gx, th *Theme, c *Palette, anim *VisibilityAnimation, w Widget) Dim {
	revealed := -1 + anim.Revealed(g)
	finalOffset := int(revealed * (float32(g.Constraints.Max.X)))
	revealedWidth := finalOffset + g.Constraints.Max.X
	defer op.Offset(image.Point{X: finalOffset}).Push(g.Ops).Pop()
	// lay out background
	paintRect(g, g.Constraints.Max, c.GetColor(col.PanelBg).NRGBA())
	// th.Bg)

	// lay out sheet contents
	dims := w(g)

	return Dim{
		Size: image.Point{
			X: int(revealedWidth),
			Y: g.Constraints.Max.Y,
		},
		Baseline: dims.Baseline,
	}
}

// ModalSheet implements the Modal Side Sheet component
// specified at https://material.io/components/sheets-side#modal-side-sheet
type ModalSheet struct {
	// MaxWidth constrains the maximum amount of horizontal screen real-estate
	// covered by the drawer. If the screen is narrower than this value, the
	// width will be inferred by reserving space for the scrim and using the
	// leftover area for the drawer. Values between 200 and 400 Dp are recommended.
	//
	// The default value used by NewModalNav is 400 Dp.
	MaxWidth Dp

	Modal *ModalLayer

	drag gesture.Drag

	// animation state
	dragging    bool
	dragStarted f32.Point
	dragOffset  int

	Sheet
}

// NewModalSheet creates a modal sheet that can render a widget on the modal layer.
func NewModalSheet(m *ModalLayer) *ModalSheet {
	s := &ModalSheet{
		MaxWidth: unit.Dp(320),
		Modal:    m,
		Sheet:    NewSheet(),
	}
	return s
}

// updateDragState ensures that a partially-dragged sheet
// snaps back into place when released and otherwise chooses
// when the sheet has been dragged far enough to close.
func (s *ModalSheet) updateDragState(g Gx, anim *VisibilityAnimation) {
	if s.dragOffset != 0 && !s.dragging && !anim.Animating() {
		if s.dragOffset < 2 {
			s.dragOffset = 0
		} else {
			s.dragOffset /= 2
		}
	} else if s.dragging && int(s.dragOffset) > g.Constraints.Max.X/10 {
		anim.Disappear(g.Now)
	}
}

// LayoutModal requests that the sheet prepare the associated ModalLayer to
// render itself (rather than another modal widget).
func (s *ModalSheet) LayoutModal(contents func(g Gx, th *Theme, c *Palette, anim *VisibilityAnimation) Dim) {
	s.Modal.Widget = func(g Gx, th *Theme, c *Palette, anim *VisibilityAnimation) Dim {
		s.updateDragState(g, anim)
		if !anim.Visible() {
			return Dim{}
		}
		for {
			event, ok := s.drag.Update(g.Metric, g.Source, gesture.Horizontal)
			if !ok {
				break
			}
			switch event.Kind {
			case pointer.Press:
				s.dragStarted = event.Position
				s.dragOffset = 0
				s.dragging = true
			case pointer.Drag:
				newOffset := int(s.dragStarted.X - event.Position.X)
				if newOffset > s.dragOffset {
					s.dragOffset = newOffset
				}
			case pointer.Release:
				fallthrough
			case pointer.Cancel:
				s.dragging = false
			}
		}
		for {
			// Beneath sheet content, listen for tap events. This prevents taps in the
			// empty sheet area from passing downward to the scrim underneath it.
			_, ok := g.Event(pointer.Filter{
				Target: s,
				Kinds:  pointer.Press | pointer.Release,
			})
			if !ok {
				break
			}
		}
		// Ensure any transformation is undone on return.
		defer op.Offset(image.Point{}).Push(g.Ops).Pop()
		if s.dragOffset != 0 || anim.Animating() {
			s.drawerTransform(g, anim).Add(g.Ops)
			g.Execute(op.InvalidateCmd{})
		}
		g.Constraints.Max.X = s.sheetWidth(g)

		// Beneath sheet content, listen for tap events. This prevents taps in the
		// empty sheet area from passing downward to the scrim underneath it.
		pr := clip.Rect(image.Rectangle{Max: g.Constraints.Max})
		defer pr.Push(g.Ops).Pop()
		event.Op(g.Ops, s)
		// lay out widget
		dims := s.Sheet.Layout(g, th, c, anim, func(g Gx) Dim {
			return contents(g, th, c, anim)
		})

		// On top of sheet content, listen for drag events to close the sheet.
		defer pointer.PassOp{}.Push(g.Ops).Pop()
		defer clip.Rect(image.Rectangle{Max: g.Constraints.Max}).Push(g.Ops).Pop()
		s.drag.Add(g.Ops)

		return dims
	}
}

// drawerTransform returns the current offset transformation
// of the sheet taking both drag and animation progress
// into account.
func (s ModalSheet) drawerTransform(g Gx, anim *VisibilityAnimation) op.TransformOp {
	finalOffset := -s.dragOffset
	return op.Offset(image.Point{X: finalOffset})
}

// sheetWidth returns the width of the sheet taking both the dimensions
// of the modal layer and the MaxWidth field into account.
func (s ModalSheet) sheetWidth(g Gx) int {
	scrimWidth := g.Dp(unit.Dp(56))
	withScrim := g.Constraints.Max.X - scrimWidth
	max := g.Dp(s.MaxWidth)
	return min(withScrim, max)
}

// ToggleVisibility triggers the appearance or disappearance of the
// ModalSheet.
func (s *ModalSheet) ToggleVisibility(when time.Time) {
	s.Modal.ToggleVisibility(when)
}

// Appear triggers the appearance of the ModalSheet.
func (s *ModalSheet) Appear(when time.Time) {
	s.Modal.Appear(when)
}

// Disappear triggers the appearance of the ModalSheet.
func (s *ModalSheet) Disappear(when time.Time) {
	s.Modal.Disappear(when)
}
