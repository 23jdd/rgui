package rgui

import (
	"image/color"
	"reflect"
	"strings"

	"rgui/draw"
	"rgui/style"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// StyleOptions sets a node's CSS type, ID and classes. Set before mounting.
// A blank Type retains the control's default type.
type StyleOptions struct {
	Type, ID string
	Classes  []string
}

// WithStyle sets a node's selector metadata. Call it on the UI thread.
func WithStyle(node Node, options StyleOptions) Node {
	c := componentOf(node)
	options.Classes = append([]string(nil), options.Classes...)
	c.layoutBase().Appearance = options
	c.styleRoot().styleRevision++
	c.invalidateMeasure()
	return c
}

func styleName(node Node, name string) Node {
	c := componentOf(node)
	c.styleName = name
	return c
}

var standardStyle = mustStyle(`
:root {
  --background: #141a24;
  --surface: #1d2634;
  --field: #222e3f;
  --track: #303f54;
  --border: #485b74;
  --text: #edf3fb;
  --muted: #a9b9cf;
  --accent: #78bfff;
  --selection: #314f73;
  --focus: #a6d5ff;
  --button: #2d4059;
  --button-hover: #3b5472;
  --button-active: #25354b;
  color: var(--text);
  background-color: var(--background);
}
button, repeatbutton {
  background-color: var(--button);
  border-color: var(--border);
  border-width: 1px;
  border-radius: 7px;
}
button:hover, repeatbutton:hover { background-color: var(--button-hover); }
button:active, repeatbutton:active { background-color: var(--button-active); }
button:selected { background-color: var(--selection); }
button:disabled, repeatbutton:disabled { background-color: var(--surface); color: var(--muted); }
button:focus, repeatbutton:focus, hyperlink:focus { border-color: var(--focus); border-width: 2px; }
hyperlink { color: var(--accent); border-radius: 5px; }
hyperlink:disabled { color: var(--muted); }
textbox, passwordbox, richtextbox, inkcanvas {
  background-color: var(--field);
  border-color: var(--border);
  border-width: 1px;
  border-radius: 7px;
}
textbox:focus, passwordbox:focus, richtextbox:focus { border-color: var(--focus); border-width: 2px; }
toolbar, statusbar, groupbox, menusurface, tooltipsurface, combosurface, calendarsurface, tableheader {
  background-color: var(--surface);
  border-color: var(--border);
  border-width: 1px;
  border-radius: 7px;
}
listviewrow:selected { background-color: var(--selection); }
listviewrow:focus { border-color: var(--focus); border-width: 1px; }
mediasurface { background-color: var(--background); border-radius: 7px; }
scrollbar { background-color: var(--track); border-radius: 3px; }
checkbox:disabled, radiobutton:disabled, slider:disabled, textbox:disabled, passwordbox:disabled, richtextbox:disabled {
  color: var(--muted);
  --accent: var(--muted);
}
`)

func mustStyle(css string) *style.Sheet {
	sheet, err := style.ParseCSS(css)
	if err != nil {
		panic(err)
	}
	return sheet
}

// DefaultStyleSheet returns the built-in CSS stylesheet.
func DefaultStyleSheet() *style.Sheet { return standardStyle }

func (c *component) cssName() string {
	if c.styleName != "" {
		return c.styleName
	}
	if c.link {
		return "hyperlink"
	}
	if c.button {
		return "button"
	}
	switch node := c.custom.(type) {
	case *choice:
		if node.radio {
			return "radiobutton"
		}
		return "checkbox"
	case *textBox:
		if node.masked {
			return "passwordbox"
		}
		return "textbox"
	case *imageNode:
		return "image"
	case *labelNode:
		return "label"
	case *pageNode:
		return "page"
	case *frameNode:
		return "frame"
	case *virtualizingStack:
		return "virtualizingstackpanel"
	case *callbackNode:
		return "drawnode"
	case nil:
	default:
		return strings.ToLower(reflect.Indirect(reflect.ValueOf(node)).Type().Name())
	}
	if c.decoration != nil {
		return "border"
	}
	switch c.panel.(type) {
	case stackLayout:
		return "stackpanel"
	case *gridLayout:
		return "grid"
	case dockLayout:
		return "dockpanel"
	case wrapLayout:
		return "wrappanel"
	case canvasLayout:
		return "canvas"
	case *viewboxLayout:
		return "viewbox"
	case popupLayout:
		return "popup"
	case nil:
		return "text"
	}
	return "panel"
}

func (c *component) cssTarget(dynamic bool, state paintState) *style.Target {
	if c == nil {
		return &style.Target{Type: "window"}
	}
	options := c.layoutBase().Appearance
	kind := options.Type
	if kind == "" {
		kind = c.cssName()
	}
	target := &style.Target{Type: kind, ID: options.ID, Classes: options.Classes, Parent: c.parent.cssTarget(dynamic, state)}
	if dynamic {
		if nodeWithin(state.hovered, c) {
			target.State |= style.Hover
		}
		if nodeWithin(state.pressed, c) {
			target.State |= style.Active
		}
		if c == state.focused {
			target.State |= style.Focus
		}
		if options := c.layoutBase().Input; options.Disabled || target.Parent.State&style.Disabled != 0 || c.button && c.onClick == nil {
			target.State |= style.Disabled
		}
		if c.selected != nil && c.selected() {
			target.State |= style.Selected
		}
		switch node := c.custom.(type) {
		case *choice:
			if node.checked() {
				target.State |= style.Checked
			}
		case *listViewRow:
			if node.selected.Get() == node.index {
				target.State |= style.Selected
			}
		}
	}
	return target
}

func nodeWithin(node, ancestor *component) bool {
	for ; node != nil; node = node.parent {
		if node == ancestor {
			return true
		}
	}
	return false
}

func (c *component) styleRoot() *component {
	for c.parent != nil {
		c = c.parent
	}
	return c
}

func (c *component) styledBox(box Box) Box {
	root := c.styleRoot()
	cache := &c.staticStyle
	if root.sheet == nil {
		if cache.sheet != nil {
			*cache = staticStyle{}
			c.invalidateMeasure()
		}
		return box
	}
	if cache.root != root || cache.sheet != root.sheet || cache.revision != root.styleRevision {
		values := style.Resolve(c.cssTarget(false, paintState{}), standardStyle, root.sheet)
		metrics := textMetrics{size: values.Length("font-size", fontSize), spacing: values.Length("letter-spacing", letterSpacing)}
		if metrics.size <= 0 {
			metrics.size = fontSize
		}
		if cache.metrics != metrics {
			c.invalidateMeasure()
		}
		*cache = staticStyle{root: root, sheet: root.sheet, revision: root.styleRevision, values: values, metrics: metrics}
	}
	values := cache.values
	box.Width = max(0, values.Length("width", box.Width))
	box.Height = max(0, values.Length("height", box.Height))
	box.MinWidth = max(0, values.Length("min-width", box.MinWidth))
	box.MinHeight = max(0, values.Length("min-height", box.MinHeight))
	box.MaxWidth = max(0, values.Length("max-width", box.MaxWidth))
	box.MaxHeight = max(0, values.Length("max-height", box.MaxHeight))
	box.Padding = styleInsets(values, "padding", box.Padding)
	box.Margin = styleInsets(values, "margin", box.Margin)
	return box
}

type staticStyle struct {
	root     *component
	sheet    *style.Sheet
	revision uint64
	values   style.Values
	metrics  textMetrics
}

type textMetrics struct{ size, spacing float32 }

func (m textMetrics) withDefaults() textMetrics {
	if m.size <= 0 {
		return textMetrics{size: fontSize, spacing: letterSpacing}
	}
	return m
}

func (n *BaseNode) textMetrics() textMetrics {
	if n.host != nil {
		return n.host.staticStyle.metrics.withDefaults()
	}
	return textMetrics{}.withDefaults()
}

func (c *component) styledMeasure(fallback measureText) measureText {
	font := c.styleRoot().styleFont
	metrics := c.staticStyle.metrics.withDefaults()
	if font.BaseSize == 0 || metrics == (textMetrics{size: fontSize, spacing: letterSpacing}) {
		return fallback
	}
	text := draw.TextStyle{Font: font, Size: metrics.size, Spacing: metrics.spacing}
	return func(value string) rl.Vector2 { return draw.MeasureText(value, text) }
}

func styleInsets(values style.Values, name string, fallback Insets) Insets {
	fallback.Top = values.Length(name+"-top", fallback.Top)
	fallback.Right = values.Length(name+"-right", fallback.Right)
	fallback.Bottom = values.Length(name+"-bottom", fallback.Bottom)
	fallback.Left = values.Length(name+"-left", fallback.Left)
	return fallback
}

func (ctx PaintContext) color(property string, fallback color.RGBA) color.RGBA {
	return ctx.Style.Color(property, fallback)
}

// TextStyle returns the shared text style for custom node painting.
func (ctx PaintContext) TextStyle() draw.TextStyle {
	metrics := ctx.metrics.withDefaults()
	return draw.TextStyle{Font: ctx.Font, Size: metrics.size, Spacing: metrics.spacing, Color: ctx.color("color", rl.RayWhite)}
}

func (c *component) drawFrame(ctx PaintContext) {
	var background, border color.RGBA
	var radius, width float32
	if decoration := c.decoration; decoration != nil {
		background, border, radius, width = decoration.Background, decoration.BorderColor, decoration.Radius, decoration.Thickness
	}
	background = ctx.color("background-color", background)
	border = ctx.color("border-color", border)
	width = max(0, ctx.Style.Length("border-width", width))
	radius = max(0, ctx.Style.Length("border-radius", radius))
	if background.A != 0 {
		draw.Rect(c.bounds, radius, draw.Style{Fill: background})
	}
	if border.A != 0 && width > 0 && ctx.Style.Get("border-style") != "none" {
		draw.Rect(inset(c.bounds, UniformInsets(width/2)), max(0, radius-width/2),
			draw.Style{Stroke: draw.Stroke{Color: border, Width: width}})
	}
}

func windowBackground(sheet *style.Sheet) color.RGBA {
	return style.Resolve(&style.Target{Type: "window"}, standardStyle, sheet).Color("background-color", rl.Black)
}
