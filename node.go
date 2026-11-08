package rgui

import (
	"rgui/draw"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Node returns the primitives to paint, in back-to-front order.
// Create a new node for each position in the tree. Primitives may read current
// state and should describe drawing without changing application state.
type Node interface {
	Primitives(PaintContext) []draw.Primitive
}

// PaintContext supplies the arranged content bounds and the window's default font.
type PaintContext struct {
	Bounds rl.Rectangle
	Font   rl.Font

	hovered, pressed, focused bool
}

// Measurable optionally supplies a custom node's desired size before padding.
// The available dimensions may be infinite. Return a finite size.
type Measurable interface {
	Measure(available rl.Vector2) rl.Vector2
}

// Element is a built-in node, or a custom Node with layout options attached.
type Element struct {
	component *component
}

func componentOf(node Node) *component {
	switch node := node.(type) {
	case nil:
		return nil
	case Element:
		return node.component
	case *Element:
		if node == nil {
			return nil
		}
		return node.component
	default:
		return &component{custom: node}
	}
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
	custom    Node
}

// Stack lays out its children vertically in declaration order.
func Stack(children ...Node) Element {
	stack := StackPanel(Vertical, stackGap, children...)
	stack.component.legacy = true
	return stack
}

// Text displays the result of read. The callback should only compute text.
func Text(read func() string) Element {
	return Element{&component{readText: read}}
}

// Button calls onClick on a completed mouse click or keyboard activation.
func Button(label string, onClick func()) Element {
	return Element{&component{text: label, button: true, onClick: onClick}}
}

// WithDisabled sets whether this button accepts input. Use it on the UI thread.
func (n Element) WithDisabled(disabled bool) Element {
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
	ctx := PaintContext{
		Bounds:  c.bounds,
		Font:    font,
		hovered: c == hovered, pressed: c == pressed, focused: c == focused,
	}
	primitives := Element{c}.Primitives(ctx)
	if c.custom != nil {
		var canvas draw.Canvas
		canvas.Clip(inset(ctx.Bounds, c.box.Padding), func() { draw.Render(primitives) })
		return
	}
	draw.Render(primitives)
}

// Primitives returns the element's own paint list. Panels paint their children
// through the view tree.
func (n Element) Primitives(ctx PaintContext) []draw.Primitive {
	c := n.component
	if c == nil || c.panel != nil {
		return nil
	}
	if c.custom != nil {
		ctx.Bounds = inset(ctx.Bounds, c.box.Padding)
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
		primitives = append(primitives, draw.RectPrimitive{Bounds: ctx.Bounds, Style: draw.Style{Fill: fill}})
		if ctx.focused {
			primitives = append(primitives, draw.RectPrimitive{
				Bounds: inset(ctx.Bounds, UniformInsets(2)),
				Style:  draw.Style{Stroke: draw.Stroke{Color: rl.Gold, Width: 2}},
			})
		}
	}
	content := inset(ctx.Bounds, c.box.Padding)
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
