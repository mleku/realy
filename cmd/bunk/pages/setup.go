package pages

import (
	"gio.mleku.dev/gio-x/component"
	"gio.mleku.dev/gio/layout"
	"gio.mleku.dev/gio/widget"
	"gio.mleku.dev/gio/widget/material"

	"realy.lol/cmd/bunk/router"
)

type Page struct {
	widget.List
	*router.R
}

func (p Page) Actions() (actions []component.AppBarAction) {
	return
}

func (p Page) Overflow() (actions []component.OverflowAction) {
	return
}

func (p Page) Layout(gtx layout.Context, th *material.Theme) (d layout.Dimensions) {
	return
}

func (p Page) NavItem() (navItem component.NavItem) {
	return
}

// NewSetup constructs a Page with the provided router.
func NewSetup(router *router.R) *Page {
	return &Page{
		R: router,
	}
}
