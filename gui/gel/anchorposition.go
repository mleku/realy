package gel

// VerticalAnchorPosition indicates the anchor position for the content
// of a component. Conventionally, this is use by AppBars and NavDrawers
// to decide how to allocate internal spacing and in which direction to
// animate certain actions.
type VerticalAnchorPosition uint

const (
	Top VerticalAnchorPosition = iota
	Bottom
)
