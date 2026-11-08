package rgui

import (
	"rgui/draw"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Node participates in measurement, placement, and painting.
// Create a new node for each position in the tree.
type Node interface {
	// Measure returns the desired size under the given constraints.
	Measure(MeasureContext) rl.Vector2
	// Place receives the final rectangle. A custom node receives its content area.
	Place(rl.Rectangle)
	// Primitives returns a paint list in back-to-front order, without changing application state.
	Primitives(PaintContext) []draw.Primitive
}

// BaseNode provides default behavior for embedding in a custom node.
// Its zero value has no preferred size and produces no primitives.
type BaseNode struct {
	Box
	bounds  rl.Rectangle
	slot    rl.Rectangle
	desired rl.Vector2
}

var _ Node = (*BaseNode)(nil)

// Measure returns a zero desired size. Override it for content-based sizing.
func (*BaseNode) Measure(MeasureContext) rl.Vector2 { return rl.Vector2{} }

// Place stores the rectangle assigned by layout. Call it when overriding Place
// if you want Bounds to reflect the assigned rectangle.
func (n *BaseNode) Place(bounds rl.Rectangle) { n.bounds = bounds }

// Bounds returns the most recent rectangle passed to Place.
func (n *BaseNode) Bounds() rl.Rectangle { return n.bounds }

// LayoutSlot returns the rectangle allocated by the parent before applying Box.
func (n *BaseNode) LayoutSlot() rl.Rectangle { return n.slot }

// DesiredSize returns the last measured size including margin.
func (n *BaseNode) DesiredSize() rl.Vector2 { return n.desired }

func (n *BaseNode) layoutBase() *BaseNode { return n }

// Primitives returns an empty paint list.
func (*BaseNode) Primitives(PaintContext) []draw.Primitive { return nil }

// MeasureContext supplies the available space and the window's text metrics.
// An unconstrained dimension is positive infinity. Desired sizes must be finite.
type MeasureContext struct {
	Available   rl.Vector2
	MeasureText func(string) rl.Vector2
}

// PaintContext supplies the window's default font.
type PaintContext struct {
	Font rl.Font

	hovered, pressed, focused bool
}

func componentOf(node Node) *component {
	switch node := node.(type) {
	case nil:
		return nil
	case *component:
		return node
	default:
		return &component{custom: node}
	}
}

type component struct {
	BaseNode
	children    []*component
	readText    func() string
	onClick     func()
	button      bool
	disabled    bool
	text        string
	boundText   *binding
	measuredBox Box
	panel       panelLayout
	legacy      bool
	custom      Node
}

func (c *component) layoutBase() *BaseNode {
	if custom, ok := c.custom.(interface{ layoutBase() *BaseNode }); ok {
		return custom.layoutBase()
	}
	return &c.BaseNode
}

// Stack lays out its children vertically in declaration order.
func Stack(children ...Node) Node {
	stack := panel(stackLayout{direction: Vertical, gap: stackGap}, children)
	stack.legacy = true
	return stack
}

// Text displays the result of read. The callback should only compute text.
func Text(read func() string) Node {
	return &component{readText: read}
}

// Button calls onClick on a completed mouse click or keyboard activation.
func Button(label string, onClick func()) Node {
	return &component{text: label, button: true, onClick: onClick}
}

// WithDisabled configures whether a button accepts input. Use it on the UI thread.
func WithDisabled(node Node, disabled bool) Node {
	if c, ok := node.(*component); ok && c.button {
		c.disabled = disabled
	}
	return node
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
	ctx := PaintContext{
		Font:    font,
		hovered: c == hovered, pressed: c == pressed, focused: c == focused,
	}
	primitives := c.Primitives(ctx)
	if c.custom != nil {
		var canvas draw.Canvas
		canvas.Clip(inset(c.bounds, c.measuredBox.Padding), func() { draw.Render(primitives) })
		return
	}
	draw.Render(primitives)
}

func (c *component) Primitives(ctx PaintContext) []draw.Primitive {
	if c.panel != nil {
		return nil
	}
	if c.custom != nil {
		return c.custom.Primitives(ctx)
	}
	style := textStyle(ctx.Font)
	var primitives []draw.Primitive
	if c.button {
		fill := rl.DarkBlue
		switch {
		case !c.enabledButton():
			fill = rl.NewColor(57, 64, 76, 255)
			style.Color = rl.LightGray
		case ctx.pressed:
			fill = rl.DarkGray
		case ctx.hovered:
			fill = rl.Blue
		}
		primitives = append(primitives, draw.RectPrimitive{Bounds: c.bounds, Style: draw.Style{Fill: fill}})
		if ctx.focused {
			primitives = append(primitives, draw.RectPrimitive{
				Bounds: inset(c.bounds, UniformInsets(2)),
				Style:  draw.Style{Stroke: draw.Stroke{Color: rl.Gold, Width: 2}},
			})
		}
	}
	content := inset(c.bounds, c.measuredBox.Padding)
	size := draw.MeasureText(c.text, style)
	text := draw.TextPrimitive{
		Position: rl.Vector2{
			X: content.X + (content.Width-size.X)/2,
			Y: content.Y + (content.Height-size.Y)/2,
		},
		Text: c.text, Style: style,
	}
	if size.X <= content.Width && size.Y <= content.Height {
		return append(primitives, text)
	}
	return append(primitives, draw.PrimitiveFunc(func() {
		var canvas draw.Canvas
		canvas.Clip(content, text.Draw)
	}))
}

func textStyle(font rl.Font) draw.TextStyle {
	return draw.TextStyle{Font: font, Size: fontSize, Spacing: letterSpacing, Color: rl.RayWhite}
}
