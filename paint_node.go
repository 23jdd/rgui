package rgui

import (
	"rgui/draw"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// RectNode draws a rectangle at Rect in a paint list, or at its assigned
// content bounds when it participates in layout.
type RectNode struct {
	BaseNode
	Rect   rl.Rectangle
	Radius float32
	Style  draw.Style
}

func (n *RectNode) Measure(MeasureContext) rl.Vector2 {
	return rl.Vector2{X: n.Rect.Width, Y: n.Rect.Height}
}

func (n *RectNode) Draw(PaintContext) {
	bounds := n.Rect
	if n.placed {
		bounds = n.Bounds()
	}
	draw.Rect(bounds, n.Radius, n.Style)
}

// TextNode draws text at Position in a paint list or at the assigned origin
// when it participates in layout. Without a Style.Font it uses the window's
// default font, size and spacing.
type TextNode struct {
	BaseNode
	Position rl.Vector2
	Text     string
	Style    draw.TextStyle
}

func (n *TextNode) Measure(ctx MeasureContext) rl.Vector2 {
	if n.Style.Font.BaseSize == 0 {
		return ctx.MeasureText(n.Text)
	}
	return draw.MeasureText(n.Text, n.textStyle(n.Style.Font))
}

func (n *TextNode) Draw(ctx PaintContext) {
	position := n.Position
	if n.placed {
		bounds := n.Bounds()
		position = rl.Vector2{X: bounds.X, Y: bounds.Y}
	}
	draw.Text(position, n.Text, n.textStyle(ctx.Font))
}

func (n *TextNode) textStyle(font rl.Font) draw.TextStyle {
	style := n.Style
	if style.Font.BaseSize == 0 {
		style.Font = font
		style.Size = fontSize
		style.Spacing = letterSpacing
	}
	if style.Size <= 0 {
		style.Size = fontSize
	}
	return style
}

// DrawNode adapts an immediate drawing callback to a Node. It has no
// preferred size. Use WithLayout when adding it as a layout child.
func DrawNode(paint func(PaintContext)) Node {
	return &callbackNode{paint: paint}
}

type callbackNode struct {
	BaseNode
	paint func(PaintContext)
}

func (n *callbackNode) Draw(ctx PaintContext) {
	if n.paint != nil {
		n.paint(ctx)
	}
}

func renderNode(node Node, ctx PaintContext) {
	if node == nil {
		return
	}
	leaf := node
	if c, ok := node.(*component); ok && c.custom != nil {
		leaf = c.custom
	}
	if drawable, ok := leaf.(interface{ Draw(PaintContext) }); ok {
		drawable.Draw(ctx)
	}
	for _, child := range node.Primitives(ctx) {
		renderNode(child, ctx)
	}
}
