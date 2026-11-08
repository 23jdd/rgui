package rgui

import (
	"image/color"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// BorderStyle describes a single-child panel's decoration.
type BorderStyle struct {
	Background, BorderColor color.RGBA
	Thickness, Radius       float32
	Padding                 Insets
	ClipToBounds            bool
}

// Border decorates a child. Clipping, when enabled, is rectangular.
func Border(child Node, style BorderStyle) Node {
	style.Thickness = max(0, style.Thickness)
	children := []Node(nil)
	if child != nil {
		children = append(children, child)
	}
	border := panel(overlayLayout{}, children)
	border.decoration = &style
	border.Box.Padding = style.Padding
	border.Box.ClipToBounds = style.ClipToBounds
	return border
}

func (c *component) contentInsets(box Box) Insets {
	padding := box.Padding
	if c.decoration != nil {
		t := c.decoration.Thickness
		padding.Left += t
		padding.Top += t
		padding.Right += t
		padding.Bottom += t
	}
	return padding
}

// EdgeOffset describes an optional offset from one canvas edge.
type EdgeOffset struct {
	Value float32
	Set   bool
}

// Anchor creates a canvas edge offset in logical pixels.
func Anchor(offset float32) EdgeOffset { return EdgeOffset{Value: offset, Set: true} }

// CanvasItem positions a child with optional edge offsets.
type CanvasItem struct {
	Child                    Node
	Left, Top, Right, Bottom EdgeOffset
}

// Canvas arranges children against its edges. Setting both edges on one axis
// gives the child a slot between those edges.
func Canvas(items ...CanvasItem) Node {
	layout := canvasLayout{items: append([]CanvasItem(nil), items...)}
	children := make([]Node, len(items))
	for i, item := range items {
		children[i] = item.Child
	}
	return panel(layout, children)
}

type canvasLayout struct{ items []CanvasItem }

func (c canvasLayout) measure(children []*component, available rl.Vector2, text measureText) rl.Vector2 {
	var desired rl.Vector2
	for i, child := range children {
		item := c.items[i]
		size := child.measure(rl.Vector2{
			X: canvasLimit(available.X, item.Left, item.Right),
			Y: canvasLimit(available.Y, item.Top, item.Bottom),
		}, text)
		desired.X = max(desired.X, size.X+item.Left.offset()+item.Right.offset())
		desired.Y = max(desired.Y, size.Y+item.Top.offset()+item.Bottom.offset())
	}
	return desired
}

func (c canvasLayout) arrange(children []*component, area rl.Rectangle) {
	for i, child := range children {
		item := c.items[i]
		x, width := canvasAxis(area.Width, child.desired.X, item.Left, item.Right)
		y, height := canvasAxis(area.Height, child.desired.Y, item.Top, item.Bottom)
		child.arrange(rl.Rectangle{X: area.X + x, Y: area.Y + y, Width: width, Height: height})
	}
}

func (edge EdgeOffset) offset() float32 {
	if edge.Set {
		return edge.Value
	}
	return 0
}

func canvasLimit(available float32, first, last EdgeOffset) float32 {
	if first.Set && last.Set {
		return max(0, available-first.Value-last.Value)
	}
	return float32(math.Inf(1))
}

func canvasAxis(available, desired float32, first, last EdgeOffset) (offset, size float32) {
	if first.Set && last.Set {
		return first.Value, canvasLimit(available, first, last)
	}
	if last.Set {
		return available - last.Value - desired, desired
	}
	return first.offset(), desired
}
