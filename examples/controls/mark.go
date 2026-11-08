package main

import (
	"rgui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type consoleMark struct{ rgui.BaseNode }

var consoleSprite = [...]string{
	".#############.",
	"##...........##",
	"#..#########..#",
	"#..#+++++++#..#",
	"#..#+++++++#..#",
	"#..#########..#",
	"#.............#",
	"#..#.....+....#",
	"#.###......+..#",
	"##.#.........##",
	".#############.",
}

const spritePixel = 4

func (*consoleMark) Measure(rgui.MeasureContext) rl.Vector2 {
	return rl.Vector2{X: float32(len(consoleSprite[0]) * spritePixel), Y: float32(len(consoleSprite) * spritePixel)}
}

func (m *consoleMark) Draw(ctx rgui.PaintContext) {
	box := m.Bounds()
	size := m.Measure(rgui.MeasureContext{})
	left := box.X + (box.Width-size.X)/2
	top := box.Y + (box.Height-size.Y)/2
	ink := ctx.TextStyle().Color
	screen := ctx.Style.Color("--accent", ink)
	for y, row := range consoleSprite {
		for x, pixel := range row {
			var tint rl.Color
			switch pixel {
			case '#':
				tint = ink
			case '+':
				tint = screen
			default:
				continue
			}
			rl.DrawRectangleRec(rl.Rectangle{X: left + float32(x*spritePixel), Y: top + float32(y*spritePixel),
				Width: spritePixel, Height: spritePixel}, tint)
		}
	}
}
