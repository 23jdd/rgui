package rgui

import (
	"reflect"
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
	Primitives(PaintContext) []Node
}

// BaseNode provides default behavior for embedding in a custom node.
// Its zero value has no preferred size and produces no primitives.
type BaseNode struct {
	Box
	Input   InputOptions
	bounds  rl.Rectangle
	placed  bool
	slot    rl.Rectangle
	desired rl.Vector2
	host    *component
}

var _ Node = (*BaseNode)(nil)

// Measure returns a zero desired size. Override it for content-based sizing.
func (*BaseNode) Measure(MeasureContext) rl.Vector2 { return rl.Vector2{} }

// Place stores the rectangle assigned by layout. Call it when overriding Place
// if you want Bounds to reflect the assigned rectangle.
func (n *BaseNode) Place(bounds rl.Rectangle) { n.bounds, n.placed = bounds, true }

// Bounds returns the most recent rectangle passed to Place.
func (n *BaseNode) Bounds() rl.Rectangle { return n.bounds }

// LayoutSlot returns the rectangle allocated by the parent before applying Box.
func (n *BaseNode) LayoutSlot() rl.Rectangle { return n.slot }

// DesiredSize returns the last measured size including margin.
func (n *BaseNode) DesiredSize() rl.Vector2 { return n.desired }

func (n *BaseNode) layoutBase() *BaseNode { return n }

// Primitives returns an empty paint list.
func (*BaseNode) Primitives(PaintContext) []Node { return nil }

// MeasureContext supplies the available space and the window's text metrics.
// An unconstrained dimension is positive infinity. Desired sizes must be finite.
type MeasureContext struct {
	Available   rl.Vector2
	MeasureText func(string) rl.Vector2
}

// PaintContext supplies the window's default font.
type PaintContext struct {
	Font   rl.Font
	canvas *draw.Canvas

	hovered, pressed, focused bool
}

func componentOf(node Node) *component {
	switch node := node.(type) {
	case nil:
		return nil
	case *component:
		return node
	default:
		value := reflect.ValueOf(node)
		if value.Kind() == reflect.Pointer && value.IsNil() {
			return nil
		}
		if owner, ok := node.(interface{ layoutBase() *BaseNode }); ok {
			base := owner.layoutBase()
			if base.host == nil {
				base.host = &component{custom: node}
			}
			return base.host
		}
		return &component{custom: node}
	}
}

type component struct {
	BaseNode
	children                 []*component
	parent                   *component
	pending                  subscriptions
	mutable                  bool
	treeDirty                bool
	readText                 func() string
	onClick                  func()
	button                   bool
	text                     string
	boundText                *binding
	measuredBox              Box
	measuredText             string
	available                rl.Vector2
	layoutScale, placedScale rl.Vector2
	measureValid, placeValid bool
	panel                    panelLayout
	legacy                   bool
	custom                   Node
	decoration               *BorderStyle
	visualDamage             rl.Rectangle
	lastDisabled             bool
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
		c.Input.Disabled = disabled
	}
	return node
}

func (c *component) enabledButton() bool {
	return c != nil && c.button && c.onClick != nil && c.inputEnabled()
}

func (c *component) mount(pending subscriptions) {
	c.pending = pending
	c.invalidateMeasure()
	if c.readText != nil {
		c.boundText = &binding{
			evaluate: func() {
				value := c.readText()
				if value != c.text {
					c.text = value
					c.invalidateMeasure()
				}
			},
			dependencies: make(map[*subscriptions]struct{}),
			pending:      pending,
		}
		c.boundText.evaluateTracked()
	} else if custom, ok := c.custom.(interface{ bind(subscriptions) *binding }); ok {
		c.boundText = custom.bind(pending)
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
	c.pending = nil
}

const (
	rowHeight     = 48
	stackGap      = 12
	fontSize      = 24
	letterSpacing = 1
)

func (c *component) buttonAt(point rl.Vector2) *component {
	return nearestButton(hitNode(c, point))
}

func (c *component) paint(canvas *draw.Canvas, font rl.Font, hovered, pressed, focused *component, damage *rl.Rectangle) {
	visible := damage == nil || nonempty(intersectRect(c.bounds, *damage))
	if visible && (c.custom != nil || c.panel == nil || c.decoration != nil) {
		c.drawSelf(canvas, font, hovered, pressed, focused)
	}
	if len(c.children) == 0 {
		return
	}
	children := func() {
		for _, child := range c.children {
			child.paint(canvas, font, hovered, pressed, focused, damage)
		}
	}
	if c.measuredBox.ClipToBounds {
		if damage != nil && !nonempty(intersectRect(c.contentBounds(), *damage)) {
			return
		}
		canvas.Clip(c.contentBounds(), children)
	} else {
		children()
	}
}

func (c *component) contentBounds() rl.Rectangle {
	return snapBounds(inset(c.bounds, c.contentInsets(c.measuredBox)), c.placedScale)
}

func (c *component) drawSelf(canvas *draw.Canvas, font rl.Font, hovered, pressed, focused *component) {
	if c.bounds.Width <= 0 || c.bounds.Height <= 0 {
		return
	}
	ctx := PaintContext{
		Font:    font,
		canvas:  canvas,
		hovered: c == hovered, pressed: c == pressed, focused: c == focused,
	}
	if c.custom != nil {
		canvas.Clip(c.contentBounds(), func() { renderNode(c, ctx) })
		return
	}
	renderNode(c, ctx)
}

func (c *component) Primitives(ctx PaintContext) []Node {
	if c.custom != nil {
		return c.custom.Primitives(ctx)
	}
	if c.decoration != nil {
		style := c.decoration
		return []Node{
			&RectNode{Rect: c.bounds, Radius: style.Radius, Style: draw.Style{Fill: style.Background}},
			&RectNode{Rect: inset(c.bounds, UniformInsets(style.Thickness/2)), Radius: max(0, style.Radius-style.Thickness/2),
				Style: draw.Style{Stroke: draw.Stroke{Color: style.BorderColor, Width: style.Thickness}}},
		}
	}
	if c.panel != nil {
		return nil
	}
	style := textStyle(ctx.Font)
	var primitives []Node
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
		primitives = append(primitives, &RectNode{Rect: c.bounds, Style: draw.Style{Fill: fill}})
		if ctx.focused {
			primitives = append(primitives, &RectNode{
				Rect:  inset(c.bounds, UniformInsets(2)),
				Style: draw.Style{Stroke: draw.Stroke{Color: rl.Gold, Width: 2}},
			})
		}
	}
	content := c.contentBounds()
	size := draw.MeasureText(c.text, style)
	text := &TextNode{
		Position: rl.Vector2{
			X: content.X + (content.Width-size.X)/2,
			Y: content.Y + (content.Height-size.Y)/2,
		},
		Text: c.text, Style: style,
	}
	if size.X <= content.Width && size.Y <= content.Height {
		return append(primitives, text)
	}
	return append(primitives, DrawNode(func(ctx PaintContext) {
		canvas := ctx.canvas
		if canvas == nil {
			canvas = &draw.Canvas{}
		}
		canvas.Clip(content, func() { text.Draw(ctx) })
	}))
}

func textStyle(font rl.Font) draw.TextStyle {
	return draw.TextStyle{Font: font, Size: fontSize, Spacing: letterSpacing, Color: rl.RayWhite}
}
