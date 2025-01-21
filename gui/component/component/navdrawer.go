package main

import (
	"gioui.org/layout"
	"widget.mleku.dev"
	"widget.mleku.dev/material"
	"realy.lol/gui/component"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

// NavDrawerPage holds the state for a page demonstrating the features of
// the NavDrawer component.
type NavDrawerPage struct {
	nonModalDrawer widget.Bool
	widget.List
	*Router
}

// NewNavDrawerPage constructs a NavDrawerPage with the provided router.
func NewNavDrawerPage(router *Router) *NavDrawerPage {
	return &NavDrawerPage{
		Router: router,
	}
}

var _ Page = &NavDrawerPage{}

func (p *NavDrawerPage) Actions() []component.AppBarAction {
	return []component.AppBarAction{}
}

func (p *NavDrawerPage) Overflow() []component.OverflowAction {
	return []component.OverflowAction{}
}

func (p *NavDrawerPage) NavItem() component.NavItem {
	return component.NavItem{
		Name: "Nav Drawer Features",
		Icon: SettingsIcon,
	}
}

func (p *NavDrawerPage) Layout(gtx C, th *material.Theme) D {
	p.List.Axis = layout.Vertical
	return material.List(th, &p.List).Layout(gtx, 1, func(gtx C, _ int) D {
		return layout.Flex{
			Alignment: layout.Middle,
			Axis:      layout.Vertical,
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return DefaultInset.Layout(gtx, material.Body1(th, `The nav drawer widget provides a consistent interface element for navigation.

The controls below allow you to see the various features available in our Navigation Drawer implementation.`).Layout)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return DetailRow{}.Layout(gtx,
					material.Body1(th, "Use non-modal drawer").Layout,
					func(gtx C) D {
						if p.nonModalDrawer.Update(gtx) {
							p.Router.NonModalDrawer = p.nonModalDrawer.Value
							if p.nonModalDrawer.Value {
								p.Router.NavAnim.Appear(gtx.Now)
							} else {
								p.Router.NavAnim.Disappear(gtx.Now)
							}
						}
						return material.Switch(th, &p.nonModalDrawer, "Use Non-Modal Navigation Drawer").Layout(gtx)
					})
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return DetailRow{}.Layout(gtx,
					material.Body1(th, "Drag to Close").Layout,
					material.Body2(th, "You can close the modal nav drawer by dragging it to the left.").Layout)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return DetailRow{}.Layout(gtx,
					material.Body1(th, "Touch Scrim to Close").Layout,
					material.Body2(th, "You can close the modal nav drawer touching anywhere in the translucent scrim to the right.").Layout)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return DetailRow{}.Layout(gtx,
					material.Body1(th, "Bottom content anchoring").Layout,
					material.Body2(th, "If you toggle support for the bottom app bar in the App Bar settings, nav drawer content will anchor to the bottom of the drawer area instead of the top.").Layout)
			}),
		)
	})
}
