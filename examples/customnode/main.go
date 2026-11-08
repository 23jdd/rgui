package main

import (
	"fmt"
	"log"

	"rgui"
	"rgui/draw"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type progressBar struct {
	rgui.BaseNode
	value *rgui.Signal[float32]
}

func (p *progressBar) Measure(ctx rgui.MeasureContext) rl.Vector2 {
	return rl.Vector2{X: min(240, ctx.Available.X), Y: 20}
}

func (p *progressBar) Primitives(_ rgui.PaintContext) []draw.Primitive {
	bounds := p.Bounds()
	fill := bounds
	fill.Width *= min(1, max(0, p.value.Get()))
	return []draw.Primitive{
		draw.RectPrimitive{Bounds: bounds, Radius: 6, Style: draw.Style{Fill: rl.DarkGray}},
		draw.RectPrimitive{Bounds: fill, Radius: 6, Style: draw.Style{Fill: rl.SkyBlue}},
	}
}

func view() rgui.Node {
	value := rgui.NewSignal(float32(0.35))
	return rgui.StackPanel(rgui.Vertical, 12,
		rgui.Text(func() string { return fmt.Sprintf("Progress: %.0f%%", value.Get()*100) }),
		rgui.WithLayout(&progressBar{value: value}, rgui.Box{HorizontalAlignment: rgui.Center}),
		rgui.Button("Advance", func() { value.Set(min(1, value.Get()+0.1)) }),
	)
}

func main() {
	if err := rgui.Run(rgui.WindowConfig{Title: "rgui custom node"}, view); err != nil {
		log.Fatal(err)
	}
}
