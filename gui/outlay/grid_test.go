package outlay

import (
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/io/system"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
)

func TestGridLockedRows(t *testing.T) {
	var grid Grid
	var ops op.Ops
	var g = Gx{
		Constraints: Exact(Pt(100, 100)),
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Source:      (&input.Router{}).Source(),
		Now:         time.Time{},
		Locale:      system.Locale{},
		Ops:         &ops,
	}

	highestX := 0
	highestY := 0

	dim := func(axis Axis, index, constraint int) int {
		return 10
	}
	layoutCell := func(gtx Gx, x, y int) Dim {
		if x > highestX {
			highestX = x
		}
		if y > highestY {
			highestY = y
		}
		return Dim{Size: Pt(10, 10)}
	}

	grid.Layout(g, 10, 10, dim, layoutCell)

	if highestX != 9 {
		t.Errorf("expected highest X index laid out to be %d, got %d", 9, highestX)
	}
	if highestY != 9 {
		t.Errorf("expected highest Y index laid out to be %d, got %d", 9, highestY)
	}

	highestX = 0
	highestY = 0
	grid.LockedRows = 3
	grid.Layout(g, 10, 10, dim, layoutCell)

	if highestX != 9 {
		t.Errorf("expected highest X index laid out to be %d, got %d", 9, highestX)
	}
	if highestY != 9 {
		t.Errorf("expected highest Y index laid out to be %d, got %d", 9, highestY)
	}
}

func TestGridSize(t *testing.T) {
	var grid Grid
	var ops op.Ops
	var g = Gx{
		Constraints: Constraints{
			Min: Pt(10, 10),
			Max: Pt(1000, 1000),
		},
		Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Source: (&input.Router{}).Source(),
		Now:    time.Time{},
		Locale: system.Locale{},
		Ops:    &ops,
	}

	dim := func(axis Axis, index, constraint int) int {
		return 10
	}
	layoutCell := func(gtx Gx, x, y int) Dim {
		return Dim{Size: Pt(10, 10)}
	}

	// Ensure the returned size is less than the maximum.
	dims := grid.Layout(g, 10, 10, dim, layoutCell)
	expected := Dim{Size: Pt(100, 100)}
	if dims != expected {
		t.Errorf("expected size %#+v, got %#+v", expected, dims)
	}

	// Ensure returned size respects maximum.
	g.Constraints.Max = Pt(50, 50)
	dims = grid.Layout(g, 10, 10, dim, layoutCell)
	expected = Dim{Size: Pt(50, 50)}
	if dims != expected {
		t.Errorf("expected size %#+v, got %#+v", expected, dims)
	}

	// Ensure returned size respects minimum.
	g.Constraints.Min = Pt(500, 500)
	g.Constraints.Max = Pt(1000, 1000)
	dims = grid.Layout(g, 10, 10, dim, layoutCell)
	expected = Dim{Size: Pt(500, 500)}
	if dims != expected {
		t.Errorf("expected size %#+v, got %#+v", expected, dims)
	}
}

func TestGridPointerEvents(t *testing.T) {
	var grid Grid
	var ops op.Ops
	router := &input.Router{}
	var gtx Gx = Gx{
		Constraints: Exact(Pt(100, 100)),
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Source:      router.Source(),
		Now:         time.Time{},
		Locale:      system.Locale{},
		Ops:         &ops,
	}

	sideSize := 100

	dim := func(axis Axis, index, constraint int) int {
		return sideSize
	}
	layoutCell := func(gtx Gx, x, y int) Dim {
		defer clip.Rect{Max: Pt(sideSize, sideSize)}.Push(gtx.Ops).Pop()
		event.Op(gtx.Ops, t)
		return Dim{Size: Pt(sideSize, sideSize)}
	}

	// Lay out the grid to establish its input handlers.
	grid.Layout(gtx, 1, 1, dim, layoutCell)
	router.Frame(gtx.Ops)

	// Drain the initial cancel event:
	_, _ = router.Event(pointer.Filter{
		Target: t,
		Kinds:  pointer.Press,
	})

	// Queue up a press.
	press := pointer.Event{
		Position: f32.Point{
			X: 50,
			Y: 50,
		},
		Kind: pointer.Press,
	}
	router.Queue(press)

	ev, ok := router.Event(pointer.Filter{
		Target: t,
		Kinds:  pointer.Press,
	})
	if !ok {
		t.Errorf("expected an event, got none")
	} else if ev.(pointer.Event).Kind != press.Kind {
		t.Errorf("expected %#+v, got %#+v", press, ev)
	}
}
