package main

import (
	"io"
	"strings"

	"gioui.org/io/clipboard"
	"gioui.org/layout"
	"widget.mleku.dev"
	"widget.mleku.dev/material"
	"realy.lol/gui/component"
)

// AboutPage holds the state for a page demonstrating the features of
// the AppBar component.
type AboutPage struct {
	eliasCopyButton, chrisCopyButtonGH, chrisCopyButtonLP widget.Clickable
	widget.List
	*Router
}

// NewAboutPage constructs a AboutPage with the provided router.
func NewAboutPage(router *Router) *AboutPage {
	return &AboutPage{
		Router: router,
	}
}

var _ Page = &AboutPage{}

func (p *AboutPage) Actions() []component.AppBarAction {
	return []component.AppBarAction{}
}

func (p *AboutPage) Overflow() []component.OverflowAction {
	return []component.OverflowAction{}
}

func (p *AboutPage) NavItem() component.NavItem {
	return component.NavItem{
		Name: "About this library",
		Icon: OtherIcon,
	}
}

const (
	sponsorEliasURL          = "https://github.com/sponsors/eliasnaur"
	sponsorChrisURLGitHub    = "https://github.com/sponsors/whereswaldon"
	sponsorChrisURLLiberapay = "https://liberapay.com/whereswaldon/"
)

func (p *AboutPage) Layout(gtx C, th *material.Theme) D {
	p.List.Axis = layout.Vertical
	return material.List(th, &p.List).Layout(gtx, 1, func(gtx C, _ int) D {
		return layout.Flex{
			Alignment: layout.Middle,
			Axis:      layout.Vertical,
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return DefaultInset.Layout(gtx, material.Body1(th, `This library implements material design components from https://material.io using https://gioui.org.

If you like this library and work like it, please consider sponsoring Elias and/or Chris!`).Layout)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return DetailRow{}.Layout(gtx,
					material.Body1(th, "Elias Naur can be sponsored on GitHub at "+sponsorEliasURL).Layout,
					func(gtx C) D {
						if p.eliasCopyButton.Clicked(gtx) {
							gtx.Execute(clipboard.WriteCmd{
								Data: io.NopCloser(strings.NewReader(sponsorEliasURL)),
							})
						}
						return material.Button(th, &p.eliasCopyButton, "Copy Sponsorship URL").Layout(gtx)
					})
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return DetailRow{}.Layout(gtx,
					material.Body1(th, "Chris Waldon can be sponsored on GitHub at "+sponsorChrisURLGitHub+" and on Liberapay at "+sponsorChrisURLLiberapay).Layout,

					func(gtx C) D {
						if p.chrisCopyButtonGH.Clicked(gtx) {
							gtx.Execute(clipboard.WriteCmd{
								Data: io.NopCloser(strings.NewReader(sponsorChrisURLGitHub)),
							})
						}
						if p.chrisCopyButtonLP.Clicked(gtx) {
							gtx.Execute(clipboard.WriteCmd{
								Data: io.NopCloser(strings.NewReader(sponsorChrisURLLiberapay)),
							})
						}
						return DefaultInset.Layout(gtx, func(gtx C) D {
							return layout.Flex{}.Layout(gtx,
								layout.Flexed(.5, material.Button(th, &p.chrisCopyButtonGH, "Copy GitHub URL").Layout),
								layout.Flexed(.5, material.Button(th, &p.chrisCopyButtonLP, "Copy Liberapay URL").Layout),
							)
						})
					})
			}),
		)
	})
}
