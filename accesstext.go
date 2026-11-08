package rgui

import (
	"rgui/draw"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// AccessText underlines the mnemonic after the first unescaped underscore.
// Use two underscores for a literal underscore.
func AccessText(label string) Node {
	text, index := parseAccessText(label)
	return &accessText{text: text, index: index}
}

func parseAccessText(label string) ([]rune, int) {
	input := []rune(label)
	text := make([]rune, 0, len(input))
	index := -1
	for i := 0; i < len(input); i++ {
		if input[i] == '_' && i+1 < len(input) {
			if input[i+1] == '_' {
				i++
			} else if index < 0 {
				index = len(text)
				continue
			}
		}
		text = append(text, input[i])
	}
	return text, index
}

type accessText struct {
	BaseNode
	text        []rune
	index       int
	measureText func(string) rl.Vector2
}

func (a *accessText) Measure(ctx MeasureContext) rl.Vector2 {
	a.measureText = ctx.MeasureText
	return ctx.MeasureText(string(a.text))
}

func (a *accessText) Primitives(ctx PaintContext) []Node {
	bounds := a.Bounds()
	y := bounds.Y + (bounds.Height-fontSize)/2
	nodes := []Node{&TextNode{Position: rl.Vector2{X: bounds.X, Y: y}, Text: string(a.text), Style: textStyle(ctx.Font)}}
	if a.index >= 0 && a.measureText != nil {
		left := a.measureText(string(a.text[:a.index])).X
		right := a.measureText(string(a.text[:a.index+1])).X
		nodes = append(nodes, &RectNode{Rect: rl.Rectangle{X: bounds.X + left, Y: y + fontSize - 1, Width: max(0, right-left), Height: 1},
			Style: draw.Style{Fill: rl.RayWhite}})
	}
	return nodes
}
