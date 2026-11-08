package main

import (
	"log"

	"rgui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func view() rgui.Node {
	label := rgui.NewSignal("Four anchors")
	button := func(text string) rgui.Node {
		return rgui.Button(text, func() { label.Set(text) })
	}
	canvas := rgui.WithLayout(rgui.Canvas(
		rgui.CanvasItem{Child: button("Top left"), Left: rgui.Anchor(8), Top: rgui.Anchor(8)},
		rgui.CanvasItem{Child: button("Top right"), Right: rgui.Anchor(8), Top: rgui.Anchor(8)},
		rgui.CanvasItem{Child: button("Bottom left"), Left: rgui.Anchor(8), Bottom: rgui.Anchor(8)},
		rgui.CanvasItem{Child: button("Clipped"), Right: rgui.Anchor(-20), Bottom: rgui.Anchor(8)},
	), rgui.Box{ClipToBounds: true})
	grid := rgui.Grid(
		[]rgui.Track{rgui.Fixed(48), rgui.Star(1)},
		[]rgui.Track{rgui.Fixed(60), rgui.Auto(), rgui.Star(1)},
		rgui.Cell{Child: rgui.Text(label.Get), ColumnSpan: 3},
		rgui.Cell{Child: canvas, Row: 1, ColumnSpan: 3},
	)
	return rgui.Border(grid, rgui.BorderStyle{
		Background: rl.NewColor(34, 41, 54, 255), BorderColor: rl.SkyBlue,
		Thickness: 2, Radius: 12, Padding: rgui.UniformInsets(8), ClipToBounds: true,
	})
}

func main() {
	if err := rgui.Run(rgui.WindowConfig{Title: "rgui extended panels", Width: 600}, view); err != nil {
		log.Fatal(err)
	}
}
