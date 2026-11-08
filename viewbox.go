package rgui

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Viewbox scales a child from its natural size into the space available.
// Stretch uses the same modes as Image.
func Viewbox(child Node, stretch ImageStretch) Node {
	layout := &viewboxLayout{stretch: stretch}
	node := panel(layout, []Node{child})
	node.ClipToBounds = true
	layout.owner = node
	return node
}

type viewboxLayout struct {
	owner   *component
	stretch ImageStretch
	natural rl.Vector2
}

type visualTransform struct{ offset, scale rl.Vector2 }

func (v *viewboxLayout) scale(available rl.Vector2) rl.Vector2 {
	if v.natural.X <= 0 || v.natural.Y <= 0 {
		return rl.Vector2{X: 1, Y: 1}
	}
	x, y := available.X/v.natural.X, available.Y/v.natural.Y
	if math.IsInf(float64(x), 1) {
		x = y
	}
	if math.IsInf(float64(y), 1) {
		y = x
	}
	if math.IsInf(float64(x), 1) || v.stretch == ImageNone {
		x, y = 1, 1
	}
	switch v.stretch {
	case ImageUniform:
		x, y = min(x, y), min(x, y)
	case ImageUniformToFill:
		x, y = max(x, y), max(x, y)
	}
	return rl.Vector2{X: x, Y: y}
}

func (v *viewboxLayout) measure(children []*component, available rl.Vector2, text measureText) rl.Vector2 {
	unbounded := float32(math.Inf(1))
	v.natural = children[0].measure(rl.Vector2{X: unbounded, Y: unbounded}, text)
	scale := v.scale(available)
	return rl.Vector2{X: min(available.X, v.natural.X*scale.X), Y: min(available.Y, v.natural.Y*scale.Y)}
}

func (v *viewboxLayout) arrange(children []*component, area rl.Rectangle) {
	scale := v.scale(rl.Vector2{X: area.Width, Y: area.Height})
	v.owner.transform = &visualTransform{scale: scale, offset: rl.Vector2{
		X: area.X + (area.Width-v.natural.X*scale.X)/2,
		Y: area.Y + (area.Height-v.natural.Y*scale.Y)/2,
	}}
	children[0].arrange(rl.Rectangle{Width: v.natural.X, Height: v.natural.Y})
}

func (c *component) visualRect(bounds rl.Rectangle) rl.Rectangle {
	for node := c.parent; node != nil; node = node.parent {
		if node.popup {
			break
		}
		if t := node.transform; t != nil {
			bounds.X = t.offset.X + bounds.X*t.scale.X
			bounds.Y = t.offset.Y + bounds.Y*t.scale.Y
			bounds.Width *= t.scale.X
			bounds.Height *= t.scale.Y
		}
	}
	return bounds
}

func (c *component) localPoint(point rl.Vector2) rl.Vector2 {
	bounds := c.visualRect(rl.Rectangle{X: c.bounds.X, Y: c.bounds.Y, Width: 1, Height: 1})
	return rl.Vector2{X: (point.X - bounds.X) / max(0.000001, bounds.Width), Y: (point.Y - bounds.Y) / max(0.000001, bounds.Height)}
}

func (n *BaseNode) localPointer(point rl.Vector2) rl.Vector2 {
	bounds := rl.Rectangle{X: n.bounds.X, Y: n.bounds.Y, Width: 1, Height: 1}
	if n.host != nil {
		bounds = n.host.visualRect(bounds)
	}
	return rl.Vector2{X: (point.X - bounds.X) / max(0.000001, bounds.Width), Y: (point.Y - bounds.Y) / max(0.000001, bounds.Height)}
}
