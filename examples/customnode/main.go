package main

import (
	"fmt"
	"log"

	"rgui"
	"rgui/draw"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type progressBar struct {
	value *rgui.Signal[float32]
}

func (p progressBar) Measure(available rl.Vector2) rl.Vector2 {
	return rl.Vector2{X: min(240, available.X), Y: 20}
}

func (p progressBar) Primitives(ctx rgui.PaintContext) []draw.Primitive {
	fill := ctx.Bounds
	fill.Width *= min(1, max(0, p.value.Get()))
	return []draw.Primitive{
		draw.RectPrimitive{Bounds: ctx.Bounds, Radius: 6, Style: draw.Style{Fill: rl.DarkGray}},
		draw.RectPrimitive{Bounds: fill, Radius: 6, Style: draw.Style{Fill: rl.SkyBlue}},
	}
}

func view() rgui.Node {
	value := rgui.NewSignal(float32(0.35))
	return rgui.StackPanel(rgui.Vertical, 12,
		rgui.Text(func() string { return fmt.Sprintf("Progress: %.0f%%", value.Get()*100) }),
		rgui.WithLayout(progressBar{value: value}, rgui.Box{HorizontalAlignment: rgui.Center}),
		rgui.Button("Advance", func() { value.Set(min(1, value.Get()+0.1)) }),
	)
}

func main() {
	if err := rgui.Run(rgui.WindowConfig{Title: "rgui custom node"}, view); err != nil {
		log.Fatal(err)
	}
}
