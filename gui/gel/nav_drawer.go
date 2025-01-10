package gel

import (
	"image"
	"image/color"
	"time"

	"gioui.org/font"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/widget"
	"gioui.org/widget/material"

	col "realy.lol/gui/color"
)

var (
	hoverOverlayAlpha    uint8 = 48
	selectedOverlayAlpha uint8 = 96
)

type NavItem struct {
	// Tag is an externally-provided identifier for the view that this item should
	// navigate to. Its value is opaque to navigation elements.
	Tag  interface{}
	Name string

	// Icon, if set, renders the provided icon to the left of the item's name.
	// Material specifies that either all navigation items should have an icon, or
	// none should. As such, if this field is nil, the Name will be aligned all the
	// way to the left. A mixture of icon and non-icon items will be misaligned.
	// Users should either set icons for all elements or none.
	Icon *widget.Icon
}

// renderNavItem holds both basic nav item state and the interaction state for
// that item.
type renderNavItem struct {
	NavItem
	hovering bool
	selected bool
	widget.Clickable
	*AlphaPalette
}

func (n *renderNavItem) Clicked(g Gx) bool {
	return n.Clickable.Clicked(g)
}

func (n *renderNavItem) Layout(g Gx, th *Theme, c *Palette) Dim {
	for {
		event, ok := g.Event(pointer.Filter{
			Target: n,
			Kinds:  pointer.Enter | pointer.Leave,
		})
		if !ok {
			break
		}
		switch ev := event.(type) {
		case pointer.Event:
			switch ev.Kind {
			case pointer.Enter:
				n.hovering = true
			case pointer.Leave, pointer.Cancel:
				n.hovering = false
			}
		}
	}
	defer pointer.PassOp{}.Push(g.Ops).Pop()
	defer clip.Rect(image.Rectangle{
		Max: g.Constraints.Max,
	}).Push(g.Ops).Pop()
	event.Op(g.Ops, n)
	return layout.Inset{
		Top:    Dp(4),
		Bottom: Dp(4),
		Left:   Dp(8),
		Right:  Dp(8),
	}.Layout(g, func(g Gx) Dim {
		return material.Clickable(g, &n.Clickable, func(g Gx) Dim {
			return layout.Stack{}.Layout(g,
				layout.Expanded(func(g Gx) Dim { return n.layoutBackground(g, th, c) }),
				layout.Stacked(func(g Gx) Dim { return n.layoutContent(g, th, c) }),
			)
		})
	})
}

func (n *renderNavItem) layoutContent(g Gx, th *Theme, c *Palette) Dim {
	g.Constraints.Min = g.Constraints.Max
	contentColor := c.GetColor(col.PanelText).NRGBA()
	// th.Palette.Fg
	if n.selected {
		contentColor = c.GetColor(col.PanelBgDim).NRGBA()
		// th.Palette.ContrastBg
	}
	return layout.Inset{
		Left:  Dp(8),
		Right: Dp(8),
	}.Layout(g, func(g Gx) Dim {
		return layout.Flex{Alignment: layout.Middle}.Layout(g,
			layout.Rigid(func(g Gx) Dim {
				if n.NavItem.Icon == nil {
					return Dim{}
				}
				return layout.Inset{Right: Dp(40)}.Layout(g,
					func(g Gx) Dim {
						iconSize := g.Dp(Dp(24))
						g.Constraints = layout.Exact(image.Pt(iconSize, iconSize))
						return n.NavItem.Icon.Layout(g, contentColor)
					})
			}),
			layout.Rigid(func(g Gx) Dim {
				l := material.Label(th, Sp(14), n.Name)
				l.Color = contentColor
				l.Font.Weight = font.Bold
				return layout.Center.Layout(g, l.Layout)
			}),
		)
	})
}

func (n *renderNavItem) layoutBackground(g Gx, th *Theme, c *Palette) Dim {
	if !n.selected && !n.hovering {
		return Dim{}
	}
	var fill color.NRGBA
	if n.hovering {
		fill = c.GetColor(col.PanelBg).NRGBA(n.AlphaPalette.Hover)
		// WithAlpha(th.Palette.Fg, n.AlphaPalette.Hover)
	} else if n.selected {
		fill = c.GetColor(col.DocBgHighlight).NRGBA(n.AlphaPalette.Selected)
		// WithAlpha(th.Palette.ContrastBg, n.AlphaPalette.Selected)
	}
	rr := g.Dp(Dp(4))
	defer clip.RRect{
		Rect: image.Rectangle{
			Max: g.Constraints.Max,
		},
		NE: rr,
		SE: rr,
		NW: rr,
		SW: rr,
	}.Push(g.Ops).Pop()
	paintRect(g, g.Constraints.Max, fill)
	return Dim{Size: g.Constraints.Max}
}

// NavDrawer implements the Material Design Navigation Drawer
// described here: https://material.io/components/navigation-drawer
type NavDrawer struct {
	AlphaPalette

	Title    string
	Subtitle string

	// Anchor indicates whether content in the nav drawer should be anchored to
	// the upper or lower edge of the drawer. This value should match the anchor
	// of an app bar if an app bar is used in conjunction with this nav drawer.
	Anchor VerticalAnchorPosition

	selectedItem    int
	selectedChanged bool // selected item changed during the last frame
	items           []renderNavItem

	navList layout.List
}

// NewNav configures a navigation drawer
func NewNav(title, subtitle string) NavDrawer {
	m := NavDrawer{
		Title:    title,
		Subtitle: subtitle,
		AlphaPalette: AlphaPalette{
			Hover:    hoverOverlayAlpha,
			Selected: selectedOverlayAlpha,
		},
	}
	return m
}

// AddNavItem inserts a navigation target into the drawer. This should be
// invoked only from the layout thread to avoid nasty race conditions.
func (m *NavDrawer) AddNavItem(item NavItem) {
	m.items = append(m.items, renderNavItem{
		NavItem:      item,
		AlphaPalette: &m.AlphaPalette,
	})
	if len(m.items) == 1 {
		m.items[0].selected = true
	}
}

func (m *NavDrawer) Layout(g Gx, th *Theme, c *Palette, anim *VisibilityAnimation) Dim {
	sheet := NewSheet()
	return sheet.Layout(g, th, c, anim, func(g Gx) Dim {
		return m.LayoutContents(g, th, c, anim)
	})
}

func (m *NavDrawer) LayoutContents(g Gx, th *Theme, c *Palette, anim *VisibilityAnimation) Dim {
	if !anim.Visible() {
		return Dim{}
	}
	spacing := layout.SpaceEnd
	if m.Anchor == Bottom {
		spacing = layout.SpaceStart
	}

	layout.Flex{
		Spacing: spacing,
		Axis:    layout.Vertical,
	}.Layout(g,
		layout.Rigid(func(g Gx) Dim {
			return layout.Inset{
				Left:   Dp(16),
				Bottom: Dp(18),
			}.Layout(g, func(g Gx) Dim {
				return layout.Flex{Axis: layout.Vertical}.Layout(g,
					layout.Rigid(func(g Gx) Dim {
						g.Constraints.Max.Y = g.Dp(Dp(36))
						g.Constraints.Min = g.Constraints.Max
						title := material.Label(th, Sp(18), m.Title)
						title.Font.Weight = font.Bold
						return layout.SW.Layout(g, title.Layout)
					}),
					layout.Rigid(func(g Gx) Dim {
						g.Constraints.Max.Y = g.Dp(Dp(20))
						g.Constraints.Min = g.Constraints.Max
						return layout.SW.Layout(g, material.Label(th, Sp(12), m.Subtitle).Layout)
					}),
				)
			})
		}),
		layout.Flexed(1, func(g Gx) Dim {
			return m.layoutNavList(g, th, c, anim)
		}),
	)
	return Dim{Size: g.Constraints.Max}
}

func (m *NavDrawer) layoutNavList(g Gx, th *Theme, c *Palette, anim *VisibilityAnimation) Dim {
	m.selectedChanged = false
	g.Constraints.Min.Y = 0
	m.navList.Axis = layout.Vertical
	return m.navList.Layout(g, len(m.items), func(g Gx, index int) Dim {
		g.Constraints.Max.Y = g.Dp(Dp(48))
		g.Constraints.Min = g.Constraints.Max
		if m.items[index].Clicked(g) {
			m.changeSelected(index)
		}
		dimensions := m.items[index].Layout(g, th, c)
		return dimensions
	})
}

func (m *NavDrawer) UnselectNavDestination() {
	m.items[m.selectedItem].selected = false
	m.selectedChanged = false
}

func (m *NavDrawer) changeSelected(newIndex int) {
	if newIndex == m.selectedItem && m.items[m.selectedItem].selected {
		return
	}
	m.items[m.selectedItem].selected = false
	m.selectedItem = newIndex
	m.items[m.selectedItem].selected = true
	m.selectedChanged = true
}

// SetNavDestination changes the selected navigation item to the item with
// the provided tag. If the provided tag does not exist, it has no effect.
func (m *NavDrawer) SetNavDestination(tag interface{}) {
	for i, item := range m.items {
		if item.Tag == tag {
			m.changeSelected(i)
			break
		}
	}
}

// CurrentNavDestination returns the tag of the navigation destination
// selected in the drawer.
func (m *NavDrawer) CurrentNavDestination() interface{} {
	return m.items[m.selectedItem].Tag
}

// NavDestinationChanged returns whether the selected navigation destination
// has changed since the last frame.
func (m *NavDrawer) NavDestinationChanged() bool {
	return m.selectedChanged
}

// ModalNavDrawer implements the Material Design Modal Navigation Drawer
// described here: https://material.io/components/navigation-drawer
type ModalNavDrawer struct {
	NavDrawer *NavDrawer
	sheet     *ModalSheet
}

// NewModalNav configures a modal navigation drawer that will render itself into the provided ModalLayer
func NewModalNav(modal *ModalLayer, title, subtitle string) *ModalNavDrawer {
	nav := NewNav(title, subtitle)
	return ModalNavFrom(&nav, modal)
}

func ModalNavFrom(nav *NavDrawer, modal *ModalLayer) *ModalNavDrawer {
	m := &ModalNavDrawer{}
	modalSheet := NewModalSheet(modal)
	m.NavDrawer = nav
	m.sheet = modalSheet
	return m
}

func (m *ModalNavDrawer) Layout() Dim {
	m.sheet.LayoutModal(func(g Gx, th *Theme, c *Palette, anim *VisibilityAnimation) Dim {
		dims := m.NavDrawer.LayoutContents(g, th, c, anim)
		if m.NavDrawer.selectedChanged {
			anim.Disappear(g.Now)
		}
		return dims
	})
	return Dim{}
}

func (m *ModalNavDrawer) ToggleVisibility(when time.Time) {
	m.Layout()
	m.sheet.ToggleVisibility(when)
}

func (m *ModalNavDrawer) Appear(when time.Time) {
	m.Layout()
	m.sheet.Appear(when)
}

func (m *ModalNavDrawer) Disappear(when time.Time) {
	m.Layout()
	m.sheet.Disappear(when)
}

func paintRect(g Gx, size image.Point, fill color.NRGBA) {
	Rect{
		Color: fill,
		Size:  size,
	}.Layout(g)
}
