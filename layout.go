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
// A nonpositive maximum is unbounded. Minimum sizes take precedence over
// maximum and explicit sizes. Margin is outside the node; Padding is inside it.
type Box struct {
	Width, Height                          float32
	MinWidth, MinHeight                    float32
	MaxWidth, MaxHeight                    float32
	Margin, Padding                        Insets
	HorizontalAlignment, VerticalAlignment Alignment
	ClipToBounds                           bool
}

// WithLayout attaches layout options to a built-in or custom node.
func WithLayout(node Node, box Box) Node {
	c := componentOf(node)
	c.layoutBase().Box = box
	c.syncLayoutProperties()
	return c
}

type measureText func(string) rl.Vector2

func (c *component) Measure(ctx MeasureContext) rl.Vector2 {
	c.layoutChanged()
	return c.measure(ctx.Available, ctx.MeasureText)
}

func (c *component) Place(slot rl.Rectangle) {
	c.arrange(slot)
}

type panelLayout interface {
	measure(children []*component, available rl.Vector2, text measureText) rl.Vector2
	arrange(children []*component, area rl.Rectangle)
}

func panel(layout panelLayout, children []Node) *component {
	node := &component{panel: layout}
	if err := node.replaceChildren(children); err != nil {
		panic(err)
	}
	return node
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
	base := c.layoutBase()
	c.syncLayoutProperties()
	if c.measureValid && c.available == available {
		return c.desired
	}
	box := c.measuredBox
	x := layoutLimits(box.Width, box.MinWidth, box.MaxWidth)
	y := layoutLimits(box.Height, box.MinHeight, box.MaxHeight)
	width := x.clamp(available.X - box.Margin.Left - box.Margin.Right)
	height := y.clamp(available.Y - box.Margin.Top - box.Margin.Bottom)
	padding := c.contentInsets(box)
	content := inset(rl.Rectangle{Width: width, Height: height}, padding)
	var desired rl.Vector2
	if c.custom != nil {
		desired = c.custom.Measure(MeasureContext{
			Available: rl.Vector2{X: content.Width, Y: content.Height}, MeasureText: text,
		})
	}
	if c.panel != nil {
		measure := text
		if c.legacy {
			measure = func(value string) rl.Vector2 {
				size := text(value)
				size.Y = rowHeight
				return size
			}
		}
		children := c.panel.measure(c.children, rl.Vector2{X: content.Width, Y: content.Height}, measure)
		desired.X, desired.Y = max(desired.X, children.X), max(desired.Y, children.Y)
	} else if c.custom == nil {
		desired = text(c.text)
		if c.button {
			desired.X += 24
			desired.Y = max(desired.Y, rowHeight)
		}
	}
	desired.X += padding.Left + padding.Right
	desired.Y += padding.Top + padding.Bottom
	desired.X = max(0, x.clamp(desired.X)+box.Margin.Left+box.Margin.Right)
	desired.Y = max(0, y.clamp(desired.Y)+box.Margin.Top+box.Margin.Bottom)
	c.desired = desired
	base.desired = desired
	c.available, c.measureValid, c.placeValid = available, true, false
	c.measuredText = c.text
	c.treeDirty = false
	return desired
}

func (c *component) arrange(slot rl.Rectangle) {
	scale := c.layoutScale
	if c.parent != nil {
		scale = c.parent.placedScale
	}
	if c.placeValid && c.slot == slot && c.placedScale == scale {
		return
	}
	box := c.measuredBox
	c.slot = slot
	c.placeValid = true
	c.placedScale = scale
	area := inset(slot, box.Margin)
	width := c.desired.X - box.Margin.Left - box.Margin.Right
	height := c.desired.Y - box.Margin.Top - box.Margin.Bottom
	x := layoutLimits(box.Width, box.MinWidth, box.MaxWidth)
	y := layoutLimits(box.Height, box.MinHeight, box.MaxHeight)
	area.X, area.Width = align(area.X, area.Width, width, x, box.HorizontalAlignment)
	area.Y, area.Height = align(area.Y, area.Height, height, y, box.VerticalAlignment)
	c.BaseNode.Place(snapBounds(area, scale))
	if c.panel != nil {
		c.panel.arrange(c.children, c.contentBounds())
	}
	if c.custom != nil {
		c.layoutBase().slot = slot
		c.custom.Place(c.contentBounds())
	}
}

type axisLimits struct {
	min, max float32
}

func layoutLimits(explicit, minimum, maximum float32) axisLimits {
	minimum = max(0, minimum)
	if maximum <= 0 {
		maximum = float32(math.Inf(1))
	}
	limits := axisLimits{min: minimum, max: max(minimum, maximum)}
	if explicit > 0 {
		size := limits.clamp(explicit)
		return axisLimits{min: size, max: size}
	}
	return limits
}

func (limits axisLimits) clamp(size float32) float32 {
	return max(limits.min, min(limits.max, size))
}

func align(start, available, desired float32, limits axisLimits, alignment Alignment) (float32, float32) {
	size := limits.clamp(min(available, desired))
	if alignment == Stretch {
		size = limits.clamp(available)
	}
	switch alignment {
	case Stretch:
		start += max(0, (available-size)/2)
	case Center:
		start += (available - size) / 2
	case End:
		start += available - size
	}
	return start, size
}

// InvalidateMeasure schedules a new layout for a custom node whose intrinsic
// size has changed. Call it on the UI thread.
func (n *BaseNode) InvalidateMeasure() {
	if n.host != nil {
		n.host.invalidateMeasure()
	}
}

func (c *component) invalidateMeasure() {
	for node := c; node != nil; node = node.parent {
		node.measureValid, node.placeValid = false, false
	}
}

func (c *component) invalidatePlace() {
	for node := c; node != nil; node = node.parent {
		node.placeValid = false
	}
}

func (c *component) syncLayoutProperties() {
	box := c.layoutBase().Box
	old := c.measuredBox
	if old != box {
		c.measuredBox = box
		old.HorizontalAlignment, old.VerticalAlignment = box.HorizontalAlignment, box.VerticalAlignment
		if old == box {
			c.invalidatePlace()
		} else {
			c.invalidateMeasure()
		}
	}
	if c.measuredText != c.text {
		c.invalidateMeasure()
	}
}

func (c *component) layoutChanged() bool {
	c.syncLayoutProperties()
	for _, child := range c.children {
		child.layoutChanged()
	}
	return !c.measureValid || !c.placeValid || c.treeDirty
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
	if root.legacy || (root.panel == nil && root.custom == nil) {
		area.Width = min(area.Width, 280)
		area.X = (viewport.X - area.Width) / 2
		area.Y = 36
		area.Height = max(0, viewport.Y-area.Y)
	}
	root.Measure(MeasureContext{Available: rl.Vector2{X: area.Width, Y: area.Height}, MeasureText: text})
	if root.panel == nil && root.custom == nil {
		area.Height = root.desired.Y
		if root.measuredBox.Height <= 0 {
			area.Height = max(rowHeight, area.Height)
		}
	}
	root.Place(area)
}
