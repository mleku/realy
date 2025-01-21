package gui

import (
	"gioui.org/op"
)

func GetDim(g Gx, w Widget) (d Dim) {
	child := op.Record(g.Ops)
	defer child.Stop()
	d = w(g)
	return
}
