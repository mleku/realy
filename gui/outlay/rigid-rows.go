package outlay

import (
	"gioui.org/op"
)

// RigidRows lays out a sequence of rigid widgets along Axis until it runs out of out space.
// It then makes a new row/column on the cross axis and fills that with widgets until it
// runs out there, repeating this process until all widgets are placed.
type RigidRows struct {
	Axis         Axis
	Alignment    Alignment
	Spacing      Spacing
	CrossSpacing Spacing
	CrossAlign   Alignment
}

// Layout children in rows/columns.
func (m RigidRows) Layout(g Gx, children ...Widget) Dim {
	converted := m.Axis.Convert(g.Constraints.Max)
	col := []FlexChild{}
	row := make([]FlexChild, 0, len(children))

	spine := Flex{
		Spacing:   m.CrossSpacing,
		Alignment: m.CrossAlign,
	}
	if m.Axis == Horizontal {
		spine.Axis = Vertical
	} else {
		spine.Axis = Horizontal
	}

	for len(children) > 0 {
		remaining := converted.X
		var i int
		for i < len(children) {
			max := converted
			newGtx := g
			newGtx.Constraints.Max = m.Axis.Convert(max)
			macro := op.Record(newGtx.Ops)
			dims := children[i](newGtx)
			call := macro.Stop()
			convertedDims := m.Axis.Convert(dims.Size)
			if remaining-convertedDims.X >= 0 {
				row = append(row, Rigid(func(g Gx) Dim {
					call.Add(g.Ops)
					return dims
				}))
				remaining -= convertedDims.X
				i++
			} else if convertedDims.X > remaining && len(row) == 0 {
				// This is the first item on the row, but it doesn't fit. We have to
				// place it here anyway.
				row = append(row, Rigid(func(g Gx) Dim {
					call.Add(g.Ops)
					return dims
				}))
				i++
				break
			} else {
				break
			}
		}
		children = children[i:]
		rowRef := row
		col = append(col, Rigid(func(g Gx) Dim {
			return Flex{
				Axis:      m.Axis,
				Spacing:   m.Spacing,
				Alignment: m.Alignment,
			}.Layout(g, rowRef...)
		}))
		// Preserve the elements already in the slice so that our closure's references to them
		// remain valid.
		row = row[len(row):]
	}
	return spine.Layout(g, col...)
}
