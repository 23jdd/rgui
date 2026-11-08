package rgui

import (
	"image/color"
	"slices"

	"rgui/draw"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// InkCanvas records primary-pointer strokes and writes completed strokes to
// the supplied signal. Points are local to the drawing surface.
func InkCanvas(strokes *Signal[[]InkStroke], pen color.RGBA, width float32) Node {
	presenter := InkPresenter(strokes).(*inkPresenter)
	return &inkCanvas{inkPresenter: *presenter, pen: pen, width: max(1, width)}
}

type inkCanvas struct {
	inkPresenter
	pen    color.RGBA
	width  float32
	active []rl.Vector2
}

func (n *inkCanvas) point(position rl.Vector2) rl.Vector2 {
	bounds := n.Bounds()
	local := n.localPointer(position)
	return rl.Vector2{X: max(0, min(bounds.Width, local.X)), Y: max(0, min(bounds.Height, local.Y))}
}

func (n *inkCanvas) PointerDown(e *InputEvent) {
	if e.Source == n && e.Button == PrimaryButton && e.CapturePointer() {
		n.active = []rl.Vector2{n.point(e.Position)}
		n.InvalidateVisual()
		e.Handled = true
	}
}

func (n *inkCanvas) PointerMove(e *InputEvent) {
	if len(n.active) == 0 {
		return
	}
	if point := n.point(e.Position); point != n.active[len(n.active)-1] {
		n.active = append(n.active, point)
		n.InvalidateVisual()
	}
}

func (n *inkCanvas) PointerUp(e *InputEvent) {
	if e.Button == PrimaryButton && len(n.active) > 0 {
		n.PointerMove(e)
		strokes := append(slices.Clone(n.strokes.Get()), InkStroke{Points: n.active, Color: n.pen, Width: n.width})
		n.active = nil
		n.strokes.Set(strokes)
		n.InvalidateVisual()
		e.Handled = true
	}
}

func (n *inkCanvas) PointerCaptureLost(*InputEvent) {
	if len(n.active) > 0 {
		n.active = nil
		n.InvalidateVisual()
	}
}

func (n *inkCanvas) Primitives(ctx PaintContext) []Node {
	nodes := []Node{&RectNode{Rect: n.Bounds(), Style: draw.Style{Fill: rl.DarkGray}}}
	nodes = append(nodes, n.inkPresenter.Primitives(ctx)...)
	if len(n.active) > 0 {
		nodes = append(nodes, inkPrimitive(n.Bounds(), InkStroke{Points: n.active, Color: n.pen, Width: n.width}))
	}
	return nodes
}
