package rgui

import (
	"rgui/draw"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// CheckBox toggles a boolean signal with a click or Space.
func CheckBox(label string, value *Signal[bool]) Node {
	if value == nil {
		value = NewSignal(false)
	}
	node := &choice{label: label, checked: value.Get, activate: func() { value.Set(!value.Get()) }}
	node.subscribe = func(pending subscriptions) *binding { return bindVisual(&node.BaseNode, value, pending) }
	return node
}

type choice struct {
	BaseNode
	label                 string
	checked               func() bool
	activate              func()
	subscribe             func(subscriptions) *binding
	focused, armed, radio bool
}

func (n *choice) bind(pending subscriptions) *binding { return n.subscribe(pending) }

func (n *choice) Measure(ctx MeasureContext) rl.Vector2 {
	size := ctx.MeasureText(n.label)
	return rl.Vector2{X: size.X + 36, Y: max(34, size.Y)}
}

func (n *choice) Tapped(e *InputEvent) {
	if e.Source == n {
		n.activate()
		e.Handled = true
	}
}

func (n *choice) FocusGained(e *InputEvent) {
	if e.Source == n {
		n.focused = true
		n.InvalidateVisual()
	}
}

func (n *choice) FocusLost(e *InputEvent) {
	if e.Source == n {
		n.focused, n.armed = false, false
		n.InvalidateVisual()
	}
}

func (n *choice) KeyDown(e *InputEvent) {
	if e.Source == n && e.Key == rl.KeySpace {
		n.armed, e.Handled = true, true
	}
}

func (n *choice) KeyUp(e *InputEvent) {
	if e.Source == n && e.Key == rl.KeySpace {
		if n.armed {
			n.activate()
		}
		n.armed, e.Handled = false, true
	}
}

func (n *choice) Primitives(ctx PaintContext) []Node {
	bounds := n.Bounds()
	text := ctx.TextStyle()
	mark := rl.Rectangle{X: bounds.X + 5, Y: bounds.Y + (bounds.Height-18)/2, Width: 18, Height: 18}
	border := ctx.color("--border", rl.LightGray)
	if n.focused {
		border = ctx.color("--focus", rl.Gold)
	}
	radius := float32(3)
	if n.radio {
		radius = 9
	}
	nodes := []Node{
		&RectNode{Rect: mark, Radius: radius, Style: draw.Style{Fill: ctx.color("--field", rl.DarkGray), Stroke: draw.Stroke{Color: border, Width: 1}}},
		&TextNode{Position: rl.Vector2{X: bounds.X + 31, Y: bounds.Y + (bounds.Height-text.Size)/2}, Text: n.label, Style: text},
	}
	if n.checked() {
		if n.radio {
			nodes = append(nodes, &RectNode{Rect: inset(mark, UniformInsets(4)), Radius: 5, Style: draw.Style{Fill: ctx.color("--accent", rl.SkyBlue)}})
		} else {
			nodes = append(nodes, DrawNode(func(PaintContext) {
				draw.Polyline([]rl.Vector2{{X: mark.X + 4, Y: mark.Y + 9}, {X: mark.X + 8, Y: mark.Y + 13}, {X: mark.X + 14, Y: mark.Y + 5}},
					false, draw.Stroke{Color: ctx.color("--accent", rl.SkyBlue), Width: 2})
			}))
		}
	}
	return nodes
}
