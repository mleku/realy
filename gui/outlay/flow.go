package outlay

import (
	"gioui.org/op"
)

// inf is an infinite axis constraint.
const inf = 1e6

// FlowElement lays out the ith element of a Grid.
type FlowElement func(gtx Gx, i int) Dim

// Flow lays out at most Num elements along the main axis.
// The number of cross axis elements depend on the total number of elements.
type Flow struct {
	Num       int
	Axis      Axis
	Alignment Alignment
	list      List
}

// FlowWrap lays out as many elements as possible along the main axis
// before wrapping to the cross axis.
type FlowWrap struct {
	Axis      Axis
	Alignment Alignment
}

type wrapData struct {
	dims Dim
	call op.CallOp
}

func (g FlowWrap) Layout(gtx Gx, num int, el FlowElement) Dim {
	defer op.TransformOp{}.Push(gtx.Ops).Pop()

	csMax := gtx.Constraints.Max
	var mainSize, crossSize, mainPos, crossPos, base, firstBase int
	gtx.Constraints.Min = Point{}
	mainCs := axisMain(g.Axis, csMax)
	crossCs := axisCross(g.Axis, gtx.Constraints.Max)

	var els []wrapData
	for i := 0; i < num; i++ {
		macro := op.Record(gtx.Ops)
		dims, okMain, okCross := g.place(gtx, i, el)
		if i == 0 {
			firstBase = dims.Size.Y - dims.Baseline
		}
		call := macro.Stop()
		if !okMain && !okCross {
			break
		}
		main := axisMain(g.Axis, dims.Size)
		cross := axisCross(g.Axis, dims.Size)
		if okMain {
			els = append(els, wrapData{dims, call})

			mainCs := axisMain(g.Axis, gtx.Constraints.Max)
			gtx.Constraints.Max = axisPoint(g.Axis, mainCs-main, crossCs)

			mainPos += main
			crossPos = max(crossPos, cross)
			base = max(base, dims.Baseline)
			continue
		}
		// okCross
		mainSize = max(mainSize, mainPos)
		crossSize += crossPos
		g.placeAll(gtx.Ops, els, crossPos, base)
		els = append(els[:0], wrapData{dims, call})

		gtx.Constraints.Max = axisPoint(g.Axis, mainCs-main, crossCs-crossPos)
		mainPos = main
		crossPos = cross
		base = dims.Baseline
	}
	mainSize = max(mainSize, mainPos)
	crossSize += crossPos
	g.placeAll(gtx.Ops, els, crossPos, base)
	sz := axisPoint(g.Axis, mainSize, crossSize)
	return Dim{Size: sz, Baseline: sz.Y - firstBase}
}

func (g FlowWrap) place(gtx Gx, i int, el FlowElement) (dims Dim, okMain, okCross bool) {
	cs := gtx.Constraints
	if g.Axis == Horizontal {
		gtx.Constraints.Max.X = inf
	} else {
		gtx.Constraints.Max.Y = inf
	}
	dims = el(gtx, i)
	okMain = dims.Size.X <= cs.Max.X
	okCross = dims.Size.Y <= cs.Max.Y
	if g.Axis == Vertical {
		okMain, okCross = okCross, okMain
	}
	return
}

func (g FlowWrap) placeAll(ops *op.Ops, els []wrapData, crossMax, baseMax int) {
	var mainPos int
	var pt Point
	for i, el := range els {
		cross := axisCross(g.Axis, el.dims.Size)
		switch g.Alignment {
		case Start:
			cross = 0
		case End:
			cross = crossMax - cross
		case Middle:
			cross = (crossMax - cross) / 2
		case Baseline:
			if g.Axis == Horizontal {
				cross = baseMax - el.dims.Baseline
			} else {
				cross = 0
			}
		}
		if cross == 0 {
			el.call.Add(ops)
		} else {
			pt = axisPoint(g.Axis, 0, cross)
			op.Offset(pt).Add(ops)
			el.call.Add(ops)
			op.Offset(pt.Mul(-1)).Add(ops)
		}
		if i == len(els)-1 {
			pt = axisPoint(g.Axis, -mainPos, crossMax)
		} else {
			main := axisMain(g.Axis, el.dims.Size)
			pt = axisPoint(g.Axis, main, 0)
			mainPos += main
		}
		op.Offset(pt).Add(ops)
	}
}

func (g *Flow) Layout(gtx Gx, num int, el FlowElement) Dim {
	if g.Num == 0 {
		return Dim{Size: gtx.Constraints.Min}
	}
	if g.Axis == g.list.Axis {
		if g.Axis == Horizontal {
			g.list.Axis = Vertical
		} else {
			g.list.Axis = Horizontal
		}
		g.list.Alignment = g.Alignment
	}
	csMax := gtx.Constraints.Max
	return g.list.Layout(gtx, (num+g.Num-1)/g.Num, func(gtx Gx, idx int) Dim {
		if g.Axis == Horizontal {
			gtx.Constraints.Max.Y = inf
		} else {
			gtx.Constraints.Max.X = inf
		}
		gtx.Constraints.Min = Point{}
		var mainMax, crossMax int
		left := axisMain(g.Axis, csMax)
		i := idx * g.Num
		n := min(num, i+g.Num)
		for ; i < n; i++ {
			dims := el(gtx, i)
			main := axisMain(g.Axis, dims.Size)
			crossMax = max(crossMax, axisCross(g.Axis, dims.Size))
			left -= main
			if left <= 0 {
				mainMax = axisMain(g.Axis, csMax)
				break
			}
			pt := axisPoint(g.Axis, main, 0)
			op.Offset(pt).Add(gtx.Ops)
			mainMax += main
		}
		return Dim{Size: axisPoint(g.Axis, mainMax, crossMax)}
	})
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func axisPoint(a Axis, main, cross int) Point {
	if a == Horizontal {
		return Point{main, cross}
	} else {
		return Point{cross, main}
	}
}

func axisMain(a Axis, sz Point) int {
	if a == Horizontal {
		return sz.X
	} else {
		return sz.Y
	}
}

func axisCross(a Axis, sz Point) int {
	if a == Horizontal {
		return sz.Y
	} else {
		return sz.X
	}
}
