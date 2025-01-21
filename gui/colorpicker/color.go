// Package colorpicker provides simple widgets for selecting an RGBA color and
// for choosing one of a set of colors.
//
// The PickerStyle type can be used to render a colorpicker (the state will be
// stored in a State). Colorpickers allow choosing specific RGBA values with
// sliders or providing an RGB hex code.
//
// The MuxStyle type can be used to render a color multiplexer (the state will
// be stored in a MuxState). Color multiplexers provide a choice from among a
// set of colors.
package colorpicker

import (
	"encoding/hex"
	"strconv"
	"strings"

	"gioui.org/font"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"widget.mleku.dev"
)

// MuxState holds the state of a color multiplexer. A color multiplexer allows
// choosing from among a set of colors.
type MuxState struct {
	Enum
	Options        map[string]*NRGBA
	OrderedOptions []string
}

// NewMuxState creates a MuxState that will provide choices between the
// MuxOptions given as parameters.
func NewMuxState(options ...MuxOption) MuxState {
	keys := make([]string, 0, len(options))
	mapped := make(map[string]*NRGBA)
	for _, opt := range options {
		keys = append(keys, opt.Label)
		mapped[opt.Label] = opt.Value
	}
	state := MuxState{
		Options:        mapped,
		OrderedOptions: keys,
	}
	if len(keys) > 0 {
		state.Enum.Value = keys[0]
	}
	return state
}

// MuxOption is one choice for the value of a color multiplexer.
type MuxOption struct {
	Label string
	Value *NRGBA
}

// Color returns the currently-selected color.
func (m MuxState) Color() *NRGBA {
	return m.Options[m.Enum.Value]
}

// MuxStyle renders a MuxState as a material design themed widget.
type MuxStyle struct {
	*MuxState
	Theme *Theme
	Label string
}

// Mux creates a MuxStyle from a theme and a state.
func Mux(theme *Theme, state *MuxState, label string) MuxStyle {
	return MuxStyle{
		Theme:    theme,
		MuxState: state,
		Label:    label,
	}
}

// Layout renders the MuxStyle into the provided context.
func (m MuxStyle) Layout(g Gx) Dim {
	g.Constraints.Min.Y = 0
	var children []FlexChild
	inset := UniformInset(Dp(2))
	children = append(children, Rigid(func(g Gx) Dim {
		return inset.Layout(g, func(g Gx) Dim {
			return Body1(m.Theme, m.Label).Layout(g)
		})
	}))
	for i := range m.OrderedOptions {
		opt := m.OrderedOptions[i]
		children = append(children, Rigid(func(g Gx) Dim {
			return inset.Layout(g, func(g Gx) Dim {
				return m.layoutOption(g, opt)
			})
		}))
	}
	return Flex{Axis: Vertical}.Layout(g, children...)
}

func (m MuxStyle) layoutOption(g Gx, option string) Dim {
	return Flex{Alignment: Middle}.Layout(g,
		Rigid(func(g Gx) Dim {
			return RadioButton(m.Theme, &m.Enum, option, option).Layout(g)
		}),
		Flexed(1, func(g Gx) Dim {
			return Inset{Left: Dp(8)}.Layout(g, func(g Gx) Dim {
				c := m.Options[option]
				if c == nil {
					return Dim{}
				}
				return borderedSquare(g, *c)
			})
		}),
	)
}

func borderedSquare(g Gx, c NRGBA) Dim {
	dims := square(g, Dp(20), NRGBA{A: 255})

	off := g.Dp(Dp(1))
	defer op.Offset(Pt(off, off)).Push(g.Ops).Pop()
	square(g, Dp(18), c)
	return dims
}

func square(g Gx, sizeDp Dp, col NRGBA) Dim {
	return rect(g, sizeDp, sizeDp, col)
}

func rect(g Gx, width, height Dp, col NRGBA) Dim {
	w, h := g.Dp(width), g.Dp(height)
	return rectAbs(g, w, h, col)
}

func rectAbs(g Gx, w, h int, col NRGBA) Dim {
	size := Point{X: w, Y: h}
	bounds := Rectangle{Max: size}
	paint.FillShape(g.Ops, col, clip.Rect(bounds).Op())
	return Dim{Size: Pt(w, h)}
}

// State is the state of a colorpicker.
type State struct {
	R, G, B, A widget.Float
	widget.Editor

	changed bool
}

// SetColor changes the color represented by the colorpicker.
func (s *State) SetColor(col NRGBA) {
	s.R.Value = float32(col.R) / 255.0
	s.G.Value = float32(col.G) / 255.0
	s.B.Value = float32(col.B) / 255.0
	s.A.Value = float32(col.A) / 255.0
	s.updateEditor()
}

// Color returns the currently selected color.
func (s State) Color() NRGBA {
	return NRGBA{
		R: s.Red(),
		G: s.Green(),
		B: s.Blue(),
		A: s.Alpha(),
	}
}

// Red returns the red value of the currently selected color.
func (s State) Red() uint8 {
	return uint8(s.R.Value * 255)
}

// Green returns the green value of the currently selected color.
func (s State) Green() uint8 {
	return uint8(s.G.Value * 255)
}

// Blue returns the blue value of the currently selected color.
func (s State) Blue() uint8 {
	return uint8(s.B.Value * 255)
}

// Alpha returns the alpha value of the currently selected color.
func (s State) Alpha() uint8 {
	return uint8(s.A.Value * 255)
}

// Update handles all state updates from the underlying widgets.
func (s *State) Update(g Gx) bool {
	changed := false
	if s.R.Update(g) || s.G.Update(g) || s.B.Update(g) || s.A.Update(g) {
		s.updateEditor()
		changed = true
	}
	for {
		_, ok := s.Editor.Update(g)
		if !ok {
			break
		}
		out, err := hex.DecodeString(s.Editor.Text())
		if err == nil && len(out) == 3 {
			s.R.Value = float32(out[0]) / 255.0
			s.G.Value = float32(out[1]) / 255.0
			s.B.Value = float32(out[2]) / 255.0
			changed = true
		}
	}
	return changed
}

func (s *State) updateEditor() {
	s.Editor.SetText(hex.EncodeToString([]byte{s.Red(), s.Green(), s.Blue()}))
}

// PickerStyle renders a color picker using material widgets.
type PickerStyle struct {
	*State
	*Theme
	Label string
	// MonospaceFace selects the typeface to use for monospace text fields.
	// The zero value will use the generic family "monospace".
	MonospaceFace font.Typeface
}

// Picker creates a pickerstyle from a theme and a state.
func Picker(th *Theme, state *State, label string) PickerStyle {
	return PickerStyle{
		Theme: th,
		State: state,
		Label: label,
	}
}

// Layout renders the PickerStyle into the provided context.
func (p PickerStyle) Layout(g Gx) Dim {
	p.State.Update(g)

	// lay out the label and editor to compute their width
	leftSide := op.Record(g.Ops)
	leftSideDims := p.layoutLeftPane(g)
	layoutLeft := leftSide.Stop()

	// lay out the sliders in the remaining horizontal space
	rgtx := g
	rgtx.Constraints.Max.X -= leftSideDims.Size.X
	rightSide := op.Record(g.Ops)
	rightSideDims := p.layoutSliders(rgtx)
	layoutRight := rightSide.Stop()

	// compute the space beneath the editor that will not extend
	// past the sliders vertically
	margin := g.Dp(Dp(4))
	sampleWidth, sampleHeight := leftSideDims.Size.X, rightSideDims.Size.Y-leftSideDims.Size.Y

	// lay everything out for real, starting with the editor/label
	layoutLeft.Add(g.Ops)

	// offset downwards and lay out the color sample
	var stack op.TransformStack
	stack = op.Offset(Pt(margin, leftSideDims.Size.Y)).Push(g.Ops)
	rectAbs(g, sampleWidth-(2*margin), sampleHeight-(2*margin), p.State.Color())
	stack.Pop()

	// offset to the right to lay out the sliders
	defer op.Offset(Pt(leftSideDims.Size.X, 0)).Push(g.Ops).Pop()
	layoutRight.Add(g.Ops)

	return Dim{
		Size: Point{
			X: g.Constraints.Max.X,
			Y: rightSideDims.Size.Y,
		},
	}
}

func (p PickerStyle) layoutLeftPane(g Gx) Dim {
	monospaceFace := p.MonospaceFace
	if len(p.MonospaceFace) == 0 {
		monospaceFace = "monospace"
	}
	g.Constraints.Min.X = 0
	inset := UniformInset(Dp(4))
	dims := Flex{Axis: Vertical}.Layout(g,
		Rigid(func(g Gx) Dim {
			return inset.Layout(g, func(g Gx) Dim {
				return Body1(p.Theme, p.Label).Layout(g)
			})
		}),
		Rigid(func(g Gx) Dim {
			return inset.Layout(g, func(g Gx) Dim {
				return Stack{}.Layout(g,
					Expanded(func(g Gx) Dim {
						return rectAbs(g, g.Constraints.Min.X, g.Constraints.Min.Y, NRGBA{R: 230, G: 230, B: 230, A: 255})
					}),
					Stacked(func(g Gx) Dim {
						return UniformInset(Dp(2)).Layout(g, func(g Gx) Dim {
							return Flex{Alignment: Baseline}.Layout(g,
								Rigid(func(g Gx) Dim {
									label := Body1(p.Theme, "#")
									label.Font.Typeface = monospaceFace
									return label.Layout(g)
								}),
								Rigid(func(g Gx) Dim {
									editor := Editor(p.Theme, &p.Editor, "rrggbb")
									editor.Font.Typeface = monospaceFace
									return editor.Layout(g)
								}),
							)
						})
					}),
				)
			})
		}),
	)
	return dims
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (p PickerStyle) layoutSliders(g Gx) Dim {
	return Flex{Axis: Vertical}.Layout(g,
		Rigid(func(g Gx) Dim {
			return p.layoutSlider(g, &p.R, "R:", valueString(p.Red()))
		}),
		Rigid(func(g Gx) Dim {
			return p.layoutSlider(g, &p.G, "G:", valueString(p.Green()))
		}),
		Rigid(func(g Gx) Dim {
			return p.layoutSlider(g, &p.B, "B:", valueString(p.Blue()))
		}),
		Rigid(func(g Gx) Dim {
			return p.layoutSlider(g, &p.A, "A:", valueString(p.Alpha()))
		}),
	)
}

func valueString(in uint8) string {
	s := strconv.Itoa(int(in))
	delta := 3 - len(s)
	if delta > 0 {
		s = strings.Repeat(" ", delta) + s
	}
	return s
}

func (p PickerStyle) layoutSlider(g Gx, value *widget.Float, label, valueStr string) Dim {
	monospaceFace := p.MonospaceFace
	if len(p.MonospaceFace) == 0 {
		monospaceFace = "monospace"
	}
	inset := UniformInset(Dp(2))
	layoutDims := Flex{Alignment: Middle}.Layout(g,
		Rigid(func(g Gx) Dim {
			return inset.Layout(g, func(g Gx) Dim {
				l := Body1(p.Theme, label)
				l.Font.Typeface = monospaceFace
				return l.Layout(g)
			})
		}),
		Flexed(1, func(g Gx) Dim {
			sliderDims := inset.Layout(g, Slider(p.Theme, value).Layout)
			return sliderDims
		}),
		Rigid(func(g Gx) Dim {
			return inset.Layout(g, func(g Gx) Dim {
				l := Body1(p.Theme, valueStr)
				l.Font.Typeface = monospaceFace
				return l.Layout(g)
			})
		}),
	)
	return layoutDims
}
