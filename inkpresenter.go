package rgui

import (
	"image/color"
	"rgui/draw"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// InkStroke stores points relative to the ink surface's content origin.
type InkStroke struct {
	Points []rl.Vector2
	Color  color.RGBA
	Width  float32
}

// InkPresenter displays strokes without editing them.
func InkPresenter(strokes *Signal[[]InkStroke]) Node {
	if strokes == nil {
		strokes = NewSignal([]InkStroke(nil))
	}
	return &inkPresenter{strokes: strokes}
}

type inkPresenter struct {
	BaseNode
	strokes *Signal[[]InkStroke]
}

func (p *inkPresenter) bind(pending subscriptions) *binding {
	return bindVisual(&p.BaseNode, p.strokes, pending)
}
func (*inkPresenter) Measure(ctx MeasureContext) rl.Vector2 {
	return rl.Vector2{X: min(280, ctx.Available.X), Y: min(180, ctx.Available.Y)}
}

func (p *inkPresenter) Primitives(PaintContext) []Node {
	strokes := p.strokes.Get()
	nodes := make([]Node, 0, len(strokes))
	for _, stroke := range strokes {
		if len(stroke.Points) != 0 && stroke.Width > 0 {
			nodes = append(nodes, inkPrimitive(p.Bounds(), stroke))
		}
	}
	return nodes
}

func inkPrimitive(bounds rl.Rectangle, stroke InkStroke) Node {
	points := make([]rl.Vector2, len(stroke.Points))
	for i, point := range stroke.Points {
		points[i] = rl.Vector2{X: bounds.X + point.X, Y: bounds.Y + point.Y}
	}
	return DrawNode(func(PaintContext) {
		if len(points) == 1 {
			draw.Circle(points[0], stroke.Width/2, draw.Style{Fill: stroke.Color})
		} else {
			draw.Polyline(points, false, draw.Stroke{Color: stroke.Color, Width: stroke.Width})
		}
	})
}
