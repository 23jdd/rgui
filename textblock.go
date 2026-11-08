package rgui

import (
	"rgui/draw"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// TextBlock displays reactive text, wrapping at the available width. The
// alignment applies to individual lines inside the node.
func TextBlock(read func() string, alignment Alignment) Node {
	if read == nil {
		read = func() string { return "" }
	}
	return &textBlock{read: read, alignment: alignment}
}

type textBlock struct {
	BaseNode
	read        func() string
	text        string
	alignment   Alignment
	measureText func(string) rl.Vector2
	lines       []string
}

func (b *textBlock) bind(pending subscriptions) *binding {
	update := &binding{pending: pending, dependencies: make(map[*subscriptions]struct{}),
		evaluate: func() {
			if text := b.read(); text != b.text {
				b.text = text
				b.InvalidateMeasure()
			}
		},
	}
	update.evaluateTracked()
	return update
}

func (b *textBlock) Measure(ctx MeasureContext) rl.Vector2 {
	b.measureText = ctx.MeasureText
	if b.host == nil || b.host.pending == nil {
		b.text = b.read()
	}
	lines := b.wrap(ctx.Available.X)
	var width float32
	for _, line := range lines {
		width = max(width, b.measureText(line).X)
	}
	return rl.Vector2{X: width, Y: float32(len(lines)) * b.textMetrics().size}
}

func (b *textBlock) wrap(width float32) []string {
	return draw.WrapLines(b.text, max(0.01, width), func(line string) float32 {
		return b.measureText(line).X
	})
}

func (b *textBlock) Place(bounds rl.Rectangle) {
	b.BaseNode.Place(bounds)
	b.lines = b.wrap(bounds.Width)
}

func (b *textBlock) Primitives(ctx PaintContext) []Node {
	bounds := b.Bounds()
	style := ctx.TextStyle()
	nodes := make([]Node, 0, len(b.lines))
	for i, line := range b.lines {
		width := b.measureText(line).X
		x := bounds.X
		switch b.alignment {
		case Center:
			x += (bounds.Width - width) / 2
		case End:
			x += bounds.Width - width
		}
		nodes = append(nodes, &TextNode{Position: rl.Vector2{X: x, Y: bounds.Y + float32(i)*style.Size}, Text: line, Style: style})
	}
	return nodes
}
