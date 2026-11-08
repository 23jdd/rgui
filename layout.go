package rgui

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Direction is the axis along which a panel places successive children.
type Direction uint8

const (
	Vertical Direction = iota
	Horizontal
)

// Alignment positions a node inside the space allocated by its parent.
type Alignment uint8

const (
	Stretch Alignment = iota
	Start
	Center
	End
)

// Insets describes the four sides of a margin or padding in logical pixels.
type Insets struct {
	Left, Top, Right, Bottom float32
}

// UniformInsets sets the same value on every side.
func UniformInsets(size float32) Insets {
	return Insets{Left: size, Top: size, Right: size, Bottom: size}
}

// Box configures a node's layout. Nonpositive Width and Height mean Auto.
// Margin is outside the node; Padding is inside it.
type Box struct {
	Width, Height                          float32
	Margin, Padding                        Insets
	HorizontalAlignment, VerticalAlignment Alignment
}

// WithLayout configures this node and returns it for use in a view expression.
func (n Node) WithLayout(box Box) Node {
	n.component.box = box
	return n
}

type measureText func(string) rl.Vector2

type panelLayout interface {
	measure(children []*component, available rl.Vector2, text measureText) rl.Vector2
	arrange(children []*component, area rl.Rectangle)
}

func panel(layout panelLayout, children []Node) Node {
	node := &component{panel: layout, children: make([]*component, len(children))}
	for i, child := range children {
		node.children[i] = child.component
	}
	return Node{node}
}

// StackPanel arranges children in declaration order along direction.
func StackPanel(direction Direction, gap float32, children ...Node) Node {
	return panel(stackLayout{direction: direction, gap: max(0, gap)}, children)
}

type stackLayout struct {
	direction Direction
	gap       float32
}

func (s stackLayout) measure(children []*component, available rl.Vector2, text measureText) rl.Vector2 {
	_, crossLimit := s.direction.axes(available)
	limit := s.direction.size(float32(math.Inf(1)), crossLimit)
	main, cross := s.gap*float32(max(0, len(children)-1)), float32(0)
	for _, child := range children {
		childMain, childCross := s.direction.axes(child.measure(limit, text))
		main += childMain
		cross = max(cross, childCross)
	}
	return s.direction.size(main, cross)
}

func (s stackLayout) arrange(children []*component, area rl.Rectangle) {
	var offset float32
	for _, child := range children {
		main, _ := s.direction.axes(child.desired)
		child.arrange(s.direction.slot(area, offset, main))
		offset += main + s.gap
	}
}

func (d Direction) axes(size rl.Vector2) (main, cross float32) {
	if d == Horizontal {
		return size.X, size.Y
	}
	return size.Y, size.X
}

func (d Direction) size(main, cross float32) rl.Vector2 {
	if d == Horizontal {
		return rl.Vector2{X: main, Y: cross}
	}
	return rl.Vector2{X: cross, Y: main}
}

func (d Direction) slot(area rl.Rectangle, offset, length float32) rl.Rectangle {
	if d == Horizontal {
		area.X += offset
		area.Width = length
	} else {
		area.Y += offset
		area.Height = length
	}
	return area
}

func (c *component) measure(available rl.Vector2, text measureText) rl.Vector2 {
	box := c.box
	width := available.X - box.Margin.Left - box.Margin.Right
	height := available.Y - box.Margin.Top - box.Margin.Bottom
	if box.Width > 0 {
		width = box.Width
	}
	if box.Height > 0 {
		height = box.Height
	}
	content := inset(rl.Rectangle{Width: width, Height: height}, box.Padding)
	var desired rl.Vector2
	if c.panel != nil {
		measure := text
		if c.legacy {
			measure = func(value string) rl.Vector2 {
				size := text(value)
				size.Y = rowHeight
				return size
			}
		}
		desired = c.panel.measure(c.children, rl.Vector2{X: content.Width, Y: content.Height}, measure)
	} else {
		desired = text(c.text)
		if c.onClick != nil {
			desired.X += 24
			desired.Y = max(desired.Y, rowHeight)
		}
	}
	desired.X += box.Padding.Left + box.Padding.Right
	desired.Y += box.Padding.Top + box.Padding.Bottom
	if box.Width > 0 {
		desired.X = box.Width
	}
	if box.Height > 0 {
		desired.Y = box.Height
	}
	desired.X = max(0, desired.X+box.Margin.Left+box.Margin.Right)
	desired.Y = max(0, desired.Y+box.Margin.Top+box.Margin.Bottom)
	c.desired = desired
	return desired
}

func (c *component) arrange(slot rl.Rectangle) {
	box := c.box
	area := inset(slot, box.Margin)
	width := c.desired.X - box.Margin.Left - box.Margin.Right
	height := c.desired.Y - box.Margin.Top - box.Margin.Bottom
	area.X, area.Width = align(area.X, area.Width, width, box.Width, box.HorizontalAlignment)
	area.Y, area.Height = align(area.Y, area.Height, height, box.Height, box.VerticalAlignment)
	c.bounds = area
	if c.panel != nil {
		c.panel.arrange(c.children, inset(area, box.Padding))
	}
}

func align(start, available, desired, fixed float32, alignment Alignment) (float32, float32) {
	size := max(0, min(available, desired))
	if fixed > 0 {
		size = fixed
	} else if alignment == Stretch {
		size = available
	}
	switch alignment {
	case Center:
		start += (available - size) / 2
	case End:
		start += available - size
	}
	return start, size
}

func inset(area rl.Rectangle, sides Insets) rl.Rectangle {
	area.X += sides.Left
	area.Y += sides.Top
	area.Width = max(0, area.Width-sides.Left-sides.Right)
	area.Height = max(0, area.Height-sides.Top-sides.Bottom)
	return area
}

func layoutRoot(root *component, viewport rl.Vector2, text measureText) {
	area := inset(rl.Rectangle{Width: viewport.X, Height: viewport.Y}, UniformInsets(16))
	if root.legacy || root.panel == nil {
		area.Width = min(area.Width, 280)
		area.X = (viewport.X - area.Width) / 2
		area.Y = 36
		area.Height = max(0, viewport.Y-area.Y)
	}
	root.measure(rl.Vector2{X: area.Width, Y: area.Height}, text)
	if root.panel == nil {
		area.Height = root.desired.Y
		if root.box.Height <= 0 {
			area.Height = max(rowHeight, area.Height)
		}
	}
	root.arrange(area)
}
