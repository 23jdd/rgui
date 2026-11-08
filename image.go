package rgui

import (
	"rgui/draw"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type ImageStretch uint8

const (
	ImageUniform ImageStretch = iota
	ImageFill
	ImageUniformToFill
	ImageNone
)

// Image displays a caller-owned raylib texture. Keep the texture alive while
// the node is mounted.
func Image(texture rl.Texture2D, stretch ImageStretch) Node {
	return &imageNode{texture: texture, stretch: stretch}
}

type imageNode struct {
	BaseNode
	texture rl.Texture2D
	stretch ImageStretch
}

func (n *imageNode) Measure(ctx MeasureContext) rl.Vector2 {
	w, h := float32(n.texture.Width), float32(n.texture.Height)
	if w <= 0 || h <= 0 {
		return rl.Vector2{}
	}
	scale := min(1, ctx.Available.X/w, ctx.Available.Y/h)
	return rl.Vector2{X: w * scale, Y: h * scale}
}

func (n *imageNode) geometry() (source, dest rl.Rectangle) {
	dest = n.Bounds()
	source.Width, source.Height = float32(n.texture.Width), float32(n.texture.Height)
	if !nonempty(source) || !nonempty(dest) || n.stretch == ImageFill {
		return source, dest
	}
	sx, sy := dest.Width/source.Width, dest.Height/source.Height
	if n.stretch == ImageUniformToFill {
		scale := max(sx, sy)
		w, h := dest.Width/scale, dest.Height/scale
		source.X, source.Y = (source.Width-w)/2, (source.Height-h)/2
		source.Width, source.Height = w, h
		return source, dest
	}
	scale := min(sx, sy)
	if n.stretch == ImageNone {
		scale = 1
	}
	w, h := source.Width*scale, source.Height*scale
	dest.X, dest.Y = dest.X+(dest.Width-w)/2, dest.Y+(dest.Height-h)/2
	dest.Width, dest.Height = w, h
	return source, dest
}

func (n *imageNode) Primitives(PaintContext) []Node {
	source, dest := n.geometry()
	if n.texture.ID == 0 || !nonempty(source) || !nonempty(dest) {
		return nil
	}
	return []Node{DrawNode(func(PaintContext) { draw.Image(n.texture, source, dest, rl.White) })}
}
