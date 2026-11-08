package draw

import rl "github.com/gen2brain/raylib-go/raylib"

// Primitive is one drawing operation in a node's paint list.
type Primitive interface {
	Draw()
}

// PrimitiveFunc adapts a drawing function to a Primitive.
type PrimitiveFunc func()

func (paint PrimitiveFunc) Draw() { paint() }

// RectPrimitive is a filled or stroked rectangle.
type RectPrimitive struct {
	Bounds rl.Rectangle
	Radius float32
	Style  Style
}

func (r RectPrimitive) Draw() { Rect(r.Bounds, r.Radius, r.Style) }

// TextPrimitive is a string with a position and text style.
type TextPrimitive struct {
	Position rl.Vector2
	Text     string
	Style    TextStyle
}

func (t TextPrimitive) Draw() { Text(t.Position, t.Text, t.Style) }

// Render draws primitives in slice order inside an active drawing pass.
func Render(primitives []Primitive) {
	for _, primitive := range primitives {
		if primitive != nil {
			primitive.Draw()
		}
	}
}
