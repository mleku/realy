package component

import (
	"image"
	"image/color"
	"strconv"
	"time"

	"gioui.org/f32"
	"gioui.org/gesture"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// TextField implements the Material Design Text Field
// described here: https://material.io/components/text-fields
type TextField struct {
	// Editor contains the edit buffer.
	widget.Editor
	// click detects when the mouse pointer clicks or hovers
	// within the textfield.
	click gesture.Click

	// Helper text to give additional context to a field.
	Helper string
	// CharLimit specifies the maximum number of characters the text input
	// will allow. Zero means "no limit".
	CharLimit uint
	// Prefix appears before the content of the text input.
	Prefix layout.Widget
	// Suffix appears after the content of the text input.
	Suffix layout.Widget

	// Animation state.
	state
	Label  Label
	Border Border
	helper helper
	anim   *Progress

	// errored tracks whether the input is in an errored state.
	// This is orthogonal to the other states: the input can be both errored
	// and inactive for example.
	errored bool
}

// Validator validates text and returns a string describing the error.
// Error is displayed as helper text.
type Validator = func(string) string

type Label struct {
	TextSize unit.Sp
	Inset    layout.Inset
	Smallest layout.Dimensions
}

type Border struct {
	Thickness unit.Dp
	Color     color.NRGBA
}

type helper struct {
	Color color.NRGBA
	Text  string
}

type state int

const (
	inactive state = iota
	hovered
	activated
	focused
)

// IsActive if input is in an active state (Active, Focused or Errored).
func (in TextField) IsActive() bool {
	return in.state >= activated
}

// IsInvalid if input is in an error state, usually when the validator returns
// an error.
func (in *TextField) IsInvalid() bool {
	return in.errored
}

// SetError puts the input into an errored state with the specified error text.
func (in *TextField) SetError(err string) {
	in.errored = true
	in.helper.Text = err
}

// ClearError clears any errored status.
func (in *TextField) ClearError() {
	in.errored = false
	in.helper.Text = in.Helper
}

// Clear the input text and reset any error status.
func (in *TextField) Clear() {
	in.Editor.SetText("")
	in.ClearError()
}

// TextTooLong returns whether the current editor text exceeds the set character
// limit.
func (in *TextField) TextTooLong() bool {
	return !(in.CharLimit == 0 || uint(len(in.Editor.Text())) < in.CharLimit)
}

func (in *TextField) Update(g C, th *material.Theme, hint string) {
	disabled := g.Source == (input.Source{})
	for {
		ev, ok := in.click.Update(g.Source)
		if !ok {
			break
		}
		switch ev.Kind {
		case gesture.KindPress:
			g.Execute(key.FocusCmd{Tag: &in.Editor})
		}
	}
	in.state = inactive
	if in.click.Hovered() && !disabled {
		in.state = hovered
	}
	hasContents := in.Editor.Len() > 0
	if hasContents {
		in.state = activated
	}
	if g.Source.Focused(&in.Editor) && !disabled {
		in.state = focused
	}
	const (
		duration = time.Millisecond * 100
	)
	if in.anim == nil {
		in.anim = &Progress{}
	}
	if in.state == activated || hasContents {
		in.anim.Start(g.Now, Forward, 0)
	}
	if in.state == focused && !hasContents && !in.anim.Started() {
		in.anim.Start(g.Now, Forward, duration)
	}
	if in.state == inactive && !hasContents && in.anim.Finished() {
		in.anim.Start(g.Now, Reverse, duration)
	}
	if in.anim.Started() {
		g.Execute(op.InvalidateCmd{})
	}
	in.anim.Update(g.Now)
	var (
		// Text size transitions.
		textNormal = th.TextSize
		textSmall  = th.TextSize * 0.8
		// Border color transitions.
		borderColor        = WithAlpha(th.Palette.Fg, 64)
		borderColorHovered = WithAlpha(th.Palette.Fg, 216)
		borderColorActive  = th.Palette.Fg
		// TODO: derive from Theme.Error or Theme.Danger
		dangerColor = color.NRGBA{R: 200, A: 255}
		// Border thickness transitions.
		borderThickness       = unit.Dp(1)
		borderThicknessActive = unit.Dp(2.0)
	)
	in.Label.TextSize = unit.Sp(lerp(float32(textSmall), float32(textNormal), 1.0-in.anim.Progress()))
	switch in.state {
	case inactive:
		in.Border.Thickness = borderThickness
		in.Border.Color = borderColor
		in.helper.Color = borderColor
	case hovered, activated:
		in.Border.Thickness = borderThickness
		in.Border.Color = borderColorHovered
		in.helper.Color = borderColorHovered
	case focused:
		in.Border.Thickness = borderThicknessActive
		in.Border.Color = borderColorActive
		in.helper.Color = borderColorHovered
	}
	if in.IsInvalid() {
		in.Border.Color = dangerColor
		in.helper.Color = dangerColor
	}
	// Calculate the dimensions of the smallest label size and store the
	// result for use in clipping.
	// Hack: Reset min constraint to 0 to avoid min == max.
	g.Constraints.Min.X = 0
	macro := op.Record(g.Ops)
	var spacing unit.Dp
	if len(hint) > 0 {
		spacing = 4
	}
	in.Label.Smallest = layout.Inset{
		Left:  spacing,
		Right: spacing,
	}.Layout(g, func(g C) D {
		return material.Label(th, textSmall, hint).Layout(g)
	})
	macro.Stop()
	labelTopInsetNormal := float32(in.Label.Smallest.Size.Y) - float32(in.Label.Smallest.Size.Y/4)
	topInsetDP := unit.Dp(labelTopInsetNormal / g.Metric.PxPerDp)
	topInsetActiveDP := (topInsetDP / 2 * -1) - unit.Dp(in.Border.Thickness)
	in.Label.Inset = layout.Inset{
		Top:  unit.Dp(lerp(float32(topInsetDP), float32(topInsetActiveDP), in.anim.Progress())),
		Left: unit.Dp(10),
	}
}

func (in *TextField) Layout(g C, th *material.Theme, hint string) D {
	in.Update(g, th, hint)
	// Offset accounts for label height, which sticks above the border dimensions.
	defer op.Offset(image.Pt(0, in.Label.Smallest.Size.Y/2)).Push(g.Ops).Pop()
	in.Label.Inset.Layout(
		g,
		func(g C) D {
			return layout.Inset{
				Left:  unit.Dp(4),
				Right: unit.Dp(4),
			}.Layout(g, func(g C) D {
				label := material.Label(th, unit.Sp(in.Label.TextSize), hint)
				label.Color = in.Border.Color
				return label.Layout(g)
			})
		})

	dims := layout.Flex{Axis: layout.Vertical}.Layout(
		g,
		layout.Rigid(func(g C) D {
			return layout.Stack{}.Layout(
				g,
				layout.Expanded(func(g C) D {
					cornerRadius := unit.Dp(th.TextSize) * 6 / 5
					dimsFunc := func(g C) D {
						return D{Size: image.Point{
							X: g.Constraints.Max.X,
							Y: g.Constraints.Min.Y,
						}}
					}
					b := widget.Border{
						Color:        in.Border.Color,
						Width:        in.Border.Thickness,
						CornerRadius: cornerRadius,
					}
					if g.Source.Focused(&in.Editor) || in.Editor.Len() > 0 {
						visibleBorder := clip.Path{}
						visibleBorder.Begin(g.Ops)
						// Move from the origin to the beginning of the
						visibleBorder.LineTo(f32.Point{
							Y: float32(g.Constraints.Min.Y),
						})
						visibleBorder.LineTo(f32.Point{
							X: float32(g.Constraints.Max.X),
							Y: float32(g.Constraints.Min.Y),
						})
						visibleBorder.LineTo(f32.Point{
							X: float32(g.Constraints.Max.X),
						})
						labelStartX := float32(g.Dp(in.Label.Inset.Left))
						labelEndX := labelStartX + float32(in.Label.Smallest.Size.X)
						labelEndY := float32(in.Label.Smallest.Size.Y)
						visibleBorder.LineTo(f32.Point{
							X: labelEndX,
						})
						visibleBorder.LineTo(f32.Point{
							X: labelEndX,
							Y: labelEndY,
						})
						visibleBorder.LineTo(f32.Point{
							X: labelStartX,
							Y: labelEndY,
						})
						visibleBorder.LineTo(f32.Point{
							X: labelStartX,
						})
						visibleBorder.LineTo(f32.Point{})
						visibleBorder.Close()
						defer clip.Outline{
							Path: visibleBorder.End(),
						}.Op().Push(g.Ops).Pop()
					}
					return b.Layout(g, dimsFunc)
				}),
				layout.Stacked(func(g C) D {
					return layout.UniformInset(unit.Dp(12)).Layout(
						g,
						func(g C) D {
							g.Constraints.Min.X = g.Constraints.Max.X
							return layout.Flex{
								Axis:      layout.Horizontal,
								Alignment: layout.Middle,
							}.Layout(
								g,
								layout.Rigid(func(g C) D {
									if in.IsActive() && in.Prefix != nil {
										return in.Prefix(g)
									}
									return D{}
								}),
								layout.Flexed(1, func(g C) D {
									return material.Editor(th, &in.Editor, "").Layout(g)
								}),
								layout.Rigid(func(g C) D {
									if in.IsActive() && in.Suffix != nil {
										return in.Suffix(g)
									}
									return D{}
								}),
							)
						},
					)
				}),
				layout.Expanded(func(g C) D {
					defer pointer.PassOp{}.Push(g.Ops).Pop()
					defer clip.Rect(image.Rectangle{
						Max: g.Constraints.Min,
					}).Push(g.Ops).Pop()
					in.click.Add(g.Ops)
					return D{}
				}),
			)
		}),
		layout.Rigid(func(g C) D {
			return layout.Flex{
				Axis:      layout.Horizontal,
				Alignment: layout.Middle,
				Spacing:   layout.SpaceBetween,
			}.Layout(
				g,
				layout.Rigid(func(g C) D {
					if in.helper.Text == "" {
						return D{}
					}
					return layout.Inset{
						Top:  unit.Dp(4),
						Left: unit.Dp(10),
					}.Layout(
						g,
						func(g C) D {
							h := material.Label(th, unit.Sp(12), in.helper.Text)
							h.Color = in.helper.Color
							return h.Layout(g)
						},
					)
				}),
				layout.Rigid(func(g C) D {
					if in.CharLimit == 0 {
						return D{}
					}
					return layout.Inset{
						Top:   unit.Dp(4),
						Right: unit.Dp(10),
					}.Layout(
						g,
						func(g C) D {
							count := material.Label(
								th,
								unit.Sp(12),
								strconv.Itoa(in.Editor.Len())+"/"+strconv.Itoa(int(in.CharLimit)),
							)
							count.Color = in.helper.Color
							return count.Layout(g)
						},
					)
				}),
			)
		}),
	)
	return D{
		Size: image.Point{
			X: dims.Size.X,
			Y: dims.Size.Y + in.Label.Smallest.Size.Y/2,
		},
		Baseline: dims.Baseline,
	}
}

// interpolate linearly between two values based on progress.
//
// Progress is expected to be [0, 1]. Values greater than 1 will therefore be
// become a coeficient.
//
// For example, 2.5 is 250% progress.
func lerp(start, end, progress float32) float32 {
	return start + (end-start)*progress
}
