package rgui

import (
	"rgui/draw"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Node is a component in a view. Create a new node for each position in the tree.
type Node struct {
	component *component
}

type component struct {
	children  []*component
	readText  func() string
	onClick   func()
	button    bool
	disabled  bool
	text      string
	boundText *binding
	bounds    rl.Rectangle
	desired   rl.Vector2
	box       Box
	panel     panelLayout
	legacy    bool
}

// Stack lays out its children vertically in declaration order.
func Stack(children ...Node) Node {
	stack := StackPanel(Vertical, stackGap, children...)
	stack.component.legacy = true
	return stack
}

// Text displays the result of read. The callback should only compute text.
func Text(read func() string) Node {
	return Node{&component{readText: read}}
}

// Button calls onClick on a completed mouse click or keyboard activation.
func Button(label string, onClick func()) Node {
	return Node{&component{text: label, button: true, onClick: onClick}}
}

// WithDisabled sets whether this button accepts input. Use it on the UI thread.
func (n Node) WithDisabled(disabled bool) Node {
	n.component.disabled = disabled
	return n
}

func (c *component) enabledButton() bool {
	return c != nil && c.button && !c.disabled && c.onClick != nil
}

func (c *component) mount(pending subscriptions) {
	if c.readText != nil {
		c.boundText = &binding{
			evaluate:     func() { c.text = c.readText() },
			dependencies: make(map[*subscriptions]struct{}),
			pending:      pending,
		}
		c.boundText.evaluateTracked()
	}
	for _, child := range c.children {
		child.mount(pending)
	}
}

func (c *component) unmount() {
	for _, child := range c.children {
		child.unmount()
	}
	if c.boundText != nil {
		c.boundText.dispose()
		c.boundText = nil
	}
}

const (
	rowHeight     = 48
	stackGap      = 12
	fontSize      = 24
	letterSpacing = 1
)

func (c *component) buttonAt(point rl.Vector2) *component {
	if c.enabledButton() && c.bounds.Width > 0 && c.bounds.Height > 0 && rl.CheckCollisionPointRec(point, c.bounds) {
		return c
	}
	for i := len(c.children) - 1; i >= 0; i-- {
		child := c.children[i]
		if button := child.buttonAt(point); button != nil {
			return button
		}
	}
	return nil
}

func (c *component) draw(font rl.Font, hovered, pressed, focused *component) {
	if c.children != nil {
		for _, child := range c.children {
			child.draw(font, hovered, pressed, focused)
		}
		return
	}
	if c.bounds.Width <= 0 || c.bounds.Height <= 0 {
		return
	}
	style := textStyle(font)
	if c.button {
		fill := rl.DarkBlue
		switch {
		case !c.enabledButton():
			fill = rl.NewColor(57, 64, 76, 255)
			style.Color = rl.LightGray
		case c == pressed:
			fill = rl.DarkGray
		case c == hovered:
			fill = rl.Blue
		}
		draw.Rect(c.bounds, 0, draw.Style{Fill: fill})
		if c == focused {
			draw.Rect(inset(c.bounds, UniformInsets(2)), 0, draw.Style{
				Stroke: draw.Stroke{Color: rl.Gold, Width: 2},
			})
		}
	}
	content := inset(c.bounds, c.box.Padding)
	size := draw.MeasureText(c.text, style)
	position := rl.Vector2{
		X: content.X + (content.Width-size.X)/2,
		Y: content.Y + (content.Height-size.Y)/2,
	}
	if size.X <= content.Width && size.Y <= content.Height {
		draw.Text(position, c.text, style)
		return
	}
	var canvas draw.Canvas
	canvas.Clip(content, func() { draw.Text(position, c.text, style) })
}

func textStyle(font rl.Font) draw.TextStyle {
	return draw.TextStyle{Font: font, Size: fontSize, Spacing: letterSpacing, Color: rl.RayWhite}
}
