package gui

type Layouter interface {
	Layout(g Gx, w Widget) (d Dim)
}
