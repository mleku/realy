package gui

type Layouter interface {
	Layout(g Gx, w Widget) (d Dim)
}
type Wigeter interface {
	Layout(g Gx) (d Dim)
}
