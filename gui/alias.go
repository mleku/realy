package gui

import (
	_l "gioui.org/layout"
	_u "gioui.org/unit"
	_i "image"
	_w "gioui.org/widget"
	_m "gioui.org/widget/material"
	_c "image/color"
	_p "gioui.org/op/paint"
)

// types

type (
	Gx          = _l.Context
	Dim         = _l.Dimensions
	Widget      = _l.Widget
	Axis        = _l.Axis
	Alignment   = _l.Alignment
	List        = _l.List
	Constraints = _l.Constraints
	Spacing     = _l.Spacing
	Dp          = _u.Dp
	Sp          = _u.Sp
	Point       = _i.Point
	Enum        = _w.Enum
	FlexChild   = _l.FlexChild
	Flex        = _l.Flex
	Inset       = _l.Inset
	Theme       = _m.Theme
	Stack       = _l.Stack
	Rectangle   = _i.Rectangle
	NRGBA       = _c.NRGBA
	Spacer      = _l.Spacer
)

// constants

var (
	Horizontal   = _l.Horizontal
	Vertical     = _l.Vertical
	Start        = _l.Start
	End          = _l.End
	Middle       = _l.Middle
	Baseline     = _l.Baseline
	SpaceSides   = _l.SpaceSides
	SpaceStart   = _l.SpaceStart
	SpaceEvenly  = _l.SpaceEvenly
	SpaceAround  = _l.SpaceAround
	SpaceBetween = _l.SpaceBetween
	SpaceEnd     = _l.SpaceEnd
)

// functions

var (
	Exact        = _l.Exact
	Pt           = _i.Pt
	Rigid        = _l.Rigid
	Flexed       = _l.Flexed
	UniformInset = _l.UniformInset
	Body1        = _m.Body1
	Editor       = _m.Editor
	RadioButton  = _m.RadioButton
	Slider       = _m.Slider
	Expanded     = _l.Expanded
	Stacked      = _l.Stacked
	FillShape    = _p.FillShape
	Fill         = _p.Fill
)
