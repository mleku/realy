package main

import (
	"image/color"

	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"realy.lol/gui/component"
)

// AppBarPage holds the state for a page demonstrating the features of
// the AppBar component.
type AppBarPage struct {
	heartBtn, plusBtn, contextBtn          widget.Clickable
	exampleOverflowState, red, green, blue widget.Clickable
	bottomBar, customNavIcon               widget.Bool
	favorited                              bool
	widget.List
	*Router
}

// NewAppBarPage constructs a AppBarPage with the provided router.
func NewAppBarPage(router *Router) *AppBarPage {
	return &AppBarPage{
		Router: router,
	}
}

var _ Page = &AppBarPage{}

func (p *AppBarPage) Actions() []component.AppBarAction {
	return []component.AppBarAction{
		{
			OverflowAction: component.OverflowAction{
				Name: "Favorite",
				Tag:  &p.heartBtn,
			},
			Layout: func(gtx layout.Context, bg, fg color.NRGBA) layout.Dimensions {
				if p.heartBtn.Clicked(gtx) {
					p.favorited = !p.favorited
				}
				btn := component.SimpleIconButton(bg, fg, &p.heartBtn, HeartIcon)
				btn.Background = bg
				if p.favorited {
					btn.Color = color.NRGBA{R: 200, A: 255}
				} else {
					btn.Color = fg
				}
				return btn.Layout(gtx)
			},
		},
		component.SimpleIconAction(&p.plusBtn, PlusIcon,
			component.OverflowAction{
				Name: "Create",
				Tag:  &p.plusBtn,
			},
		),
	}
}

func (p *AppBarPage) Overflow() []component.OverflowAction {
	return []component.OverflowAction{
		{
			Name: "Example 1",
			Tag:  &p.exampleOverflowState,
		},
		{
			Name: "Example 2",
			Tag:  &p.exampleOverflowState,
		},
	}
}

func (p *AppBarPage) NavItem() component.NavItem {
	return component.NavItem{
		Name: "App Bar Features",
		Icon: HomeIcon,
	}
}

const (
	settingNameColumnWidth    = .3
	settingDetailsColumnWidth = 1 - settingNameColumnWidth
)

func (p *AppBarPage) Layout(gtx C, th *material.Theme) D {
	p.List.Axis = layout.Vertical
	return material.List(th, &p.List).Layout(gtx, 1, func(gtx C, _ int) D {
		return layout.Flex{
			Alignment: layout.Middle,
			Axis:      layout.Vertical,
		}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return DefaultInset.Layout(gtx, material.Body1(th, `The app bar widget provides a consistent interface element for triggering navigation and page-specific actions.

The controls below allow you to see the various features available in our App Bar implementation.`).Layout)
			}),
			layout.Rigid(func(gtx C) D {
				return DetailRow{}.Layout(gtx, material.Body1(th, "Contextual App Bar").Layout, func(gtx C) D {
					if p.contextBtn.Clicked(gtx) {
						p.Router.AppBar.SetContextualActions(
							[]component.AppBarAction{
								component.SimpleIconAction(&p.red, HeartIcon,
									component.OverflowAction{
										Name: "House",
										Tag:  &p.red,
									},
								),
							},
							[]component.OverflowAction{
								{
									Name: "foo",
									Tag:  &p.blue,
								},
								{
									Name: "bar",
									Tag:  &p.green,
								},
							},
						)
						p.Router.AppBar.ToggleContextual(gtx.Now, "Contextual Title")
					}
					return material.Button(th, &p.contextBtn, "Trigger").Layout(gtx)
				})
			}),
			layout.Rigid(func(gtx C) D {
				return DetailRow{}.Layout(gtx,
					material.Body1(th, "Bottom App Bar").Layout,
					func(gtx C) D {
						if p.bottomBar.Update(gtx) {
							if p.bottomBar.Value {
								p.Router.ModalNavDrawer.Anchor = component.Bottom
								p.Router.AppBar.Anchor = component.Bottom
							} else {
								p.Router.ModalNavDrawer.Anchor = component.Top
								p.Router.AppBar.Anchor = component.Top
							}
							p.Router.BottomBar = p.bottomBar.Value
						}

						return material.Switch(th, &p.bottomBar, "Use Bottom App Bar").Layout(gtx)
					})
			}),
			layout.Rigid(func(gtx C) D {
				return DetailRow{}.Layout(gtx,
					material.Body1(th, "Custom Navigation Icon").Layout,
					func(gtx C) D {
						if p.customNavIcon.Update(gtx) {
							if p.customNavIcon.Value {
								p.Router.AppBar.NavigationIcon = HomeIcon
							} else {
								p.Router.AppBar.NavigationIcon = MenuIcon
							}
						}
						return material.Switch(th, &p.customNavIcon, "Use Custom Navigation Icon").Layout(gtx)
					})
			}),
			layout.Rigid(func(gtx C) D {
				return DetailRow{}.Layout(gtx,
					material.Body1(th, "Animated Resize").Layout,
					material.Body2(th, "Resize the width of your screen to see app bar actions collapse into or emerge from the overflow menu (as size permits).").Layout,
				)
			}),
			layout.Rigid(func(gtx C) D {
				return DetailRow{}.Layout(gtx,
					material.Body1(th, "Custom Action Buttons").Layout,
					material.Body2(th, "Click the heart action to see custom button behavior.").Layout)
			}),
		)
	})
}
