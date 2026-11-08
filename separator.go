package rgui

import (
	"image/color"

	"rgui/draw"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Separator draws a noninteractive line along the given direction.
func Separator(direction Direction, thickness float32, color color.RGBA) Node {
	return &separator{direction: direction, thickness: max(0, thickness), color: color}
}

type separator struct {
	BaseNode
	direction Direction
	thickness float32
	color     color.RGBA
}

func (s *separator) Measure(MeasureContext) rl.Vector2 {
	return s.direction.size(0, s.thickness)
}

func (s *separator) Primitives(ctx PaintContext) []Node {
	line := s.Bounds()
	if s.direction == Horizontal {
		size := min(s.thickness, line.Height)
		line.Y += (line.Height - size) / 2
		line.Height = size
	} else {
		size := min(s.thickness, line.Width)
		line.X += (line.Width - size) / 2
		line.Width = size
	}
	return []Node{&RectNode{Rect: line, Style: draw.Style{Fill: ctx.color("--separator-color", s.color)}}}
}
