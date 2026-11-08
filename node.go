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
	text      string
	boundText *binding
	bounds    rl.Rectangle
}

// Stack lays out its children vertically in declaration order.
func Stack(children ...Node) Node {
	stack := &component{children: make([]*component, len(children))}
	for i, child := range children {
		stack.children[i] = child.component
	}
	return Node{stack}
}

// Text displays the result of read. The callback should only compute text.
func Text(read func() string) Node {
	return Node{&component{readText: read}}
}

// Button calls onClick when the left mouse button is pressed and released inside it.
func Button(label string, onClick func()) Node {
	return Node{&component{text: label, onClick: onClick}}
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

func (c *component) layout(x, y, width float32) float32 {
	if c.children == nil {
		c.bounds = rl.Rectangle{X: x, Y: y, Width: width, Height: rowHeight}
		return rowHeight
	}
	var height float32
	for i, child := range c.children {
		if i > 0 {
			height += stackGap
		}
		height += child.layout(x, y+height, width)
	}
	return height
}

func (c *component) buttonAt(point rl.Vector2) *component {
	if c.onClick != nil && rl.CheckCollisionPointRec(point, c.bounds) {
		return c
	}
	for _, child := range c.children {
		if button := child.buttonAt(point); button != nil {
			return button
		}
	}
	return nil
}

func (c *component) draw(font rl.Font, hovered, pressed *component) {
	if c.children != nil {
		for _, child := range c.children {
			child.draw(font, hovered, pressed)
		}
		return
	}
	if c.onClick != nil {
		fill := rl.DarkBlue
		if c == hovered {
			fill = rl.Blue
			if c == pressed {
				fill = rl.DarkGray
			}
		}
		draw.Rect(c.bounds, 0, draw.Style{Fill: fill})
	}
	style := draw.TextStyle{Font: font, Size: fontSize, Spacing: letterSpacing, Color: rl.RayWhite}
	size := draw.MeasureText(c.text, style)
	position := rl.Vector2{
		X: c.bounds.X + (c.bounds.Width-size.X)/2,
		Y: c.bounds.Y + (c.bounds.Height-size.Y)/2,
	}
	draw.Text(position, c.text, style)
}
