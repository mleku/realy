package main

import (
	"image"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/op"
	"gioui.org/op/clip"
	"widget.mleku.dev/text"
	"widget.mleku.dev/material"
	"realy.lol/gui/colorpicker"
)

func main() {
	go func() {
		w := new(app.Window)
		if err := loop(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

var white = NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}

func loop(w *app.Window) error {
	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
	background := white
	current := NRGBA{R: 255, G: 128, B: 75, A: 255}
	picker := colorpicker.State{}
	picker.SetColor(current)
	muxState := colorpicker.NewMuxState(
		[]colorpicker.MuxOption{
			{
				Label: "current",
				Value: &current,
			},
			{
				Label: "background",
				Value: &th.Palette.Bg,
			},
			{
				Label: "foreground",
				Value: &th.Palette.Fg,
			},
		}...)
	background = *muxState.Color()
	var ops op.Ops
	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			g := app.NewContext(&ops, e)
			if muxState.Update(g) {
				background = *muxState.Color()
			}
			if picker.Update(g) {
				current = picker.Color()
				background = *muxState.Color()
			}
			Flex{Axis: Vertical}.Layout(g,
				Rigid(func(g Gx) Dim {
					return colorpicker.PickerStyle{
						Label:         "Current",
						Theme:         th,
						State:         &picker,
						MonospaceFace: "Go Mono",
					}.Layout(g)
				}),
				Rigid(func(g Gx) Dim {
					return Flex{}.Layout(g,
						Rigid(func(g Gx) Dim {
							return colorpicker.Mux(th, &muxState, "Display Right:").Layout(g)
						}),
						Flexed(1, func(g Gx) Dim {
							size := g.Constraints.Max
							FillShape(g.Ops, background, clip.Rect(image.Rectangle{Max: size}).Op())
							return Dim{Size: size}
						}),
					)
				}),
			)
			e.Frame(g.Ops)
		}
	}
}
