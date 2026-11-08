package draw

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Canvas tracks nested drawing scopes. Use one Canvas per drawing pass.
type Canvas struct {
	clip              *rl.Rectangle
	scale             rl.Vector2
	framebufferHeight int32
}

// NewCanvas creates a canvas whose clip edges are mapped directly to pixels.
func NewCanvas(scale rl.Vector2, framebufferHeight int32) Canvas {
	return Canvas{scale: scale, framebufferHeight: framebufferHeight}
}

// Clip limits drawing to bounds in screen coordinates, intersected with any
// enclosing clip. Transforms affect the content inside the clip.
func (c *Canvas) Clip(bounds rl.Rectangle, paint func()) {
	previous := c.clip
	if previous != nil {
		bounds = intersect(bounds, *previous)
	}
	c.clip = &bounds
	c.applyClip(bounds)
	defer func() {
		c.clip = previous
		if previous != nil {
			c.applyClip(*previous)
		} else {
			rl.EndScissorMode()
		}
	}()
	paint()
}

func intersect(a, b rl.Rectangle) rl.Rectangle {
	x, y := max(a.X, b.X), max(a.Y, b.Y)
	right, bottom := min(a.X+a.Width, b.X+b.Width), min(a.Y+a.Height, b.Y+b.Height)
	return rl.Rectangle{X: x, Y: y, Width: max(0, right-x), Height: max(0, bottom-y)}
}

func (c *Canvas) applyClip(bounds rl.Rectangle) {
	if c.framebufferHeight > 0 {
		x := int32(math.Round(float64(bounds.X * c.scale.X)))
		y := int32(math.Round(float64(bounds.Y * c.scale.Y)))
		right := int32(math.Round(float64((bounds.X + bounds.Width) * c.scale.X)))
		bottom := int32(math.Round(float64((bounds.Y + bounds.Height) * c.scale.Y)))
		rl.DrawRenderBatchActive()
		rl.EnableScissorTest()
		rl.Scissor(x, c.framebufferHeight-bottom, max(0, right-x), max(0, bottom-y))
		return
	}
	x, y := math.Ceil(float64(bounds.X)), math.Ceil(float64(bounds.Y))
	right := math.Floor(float64(bounds.X + bounds.Width))
	bottom := math.Floor(float64(bounds.Y + bounds.Height))
	rl.BeginScissorMode(int32(x), int32(y), int32(max(0, right-x)), int32(max(0, bottom-y)))
}

// Transform scales and rotates around the local origin, then translates it.
// Nested transforms compose with their parents. Use positive scale factors.
func (c *Canvas) Transform(offset rl.Vector2, degrees float32, scale rl.Vector2, paint func()) {
	rl.PushMatrix()
	defer rl.PopMatrix()
	rl.Translatef(offset.X, offset.Y, 0)
	rl.Rotatef(degrees, 0, 0, 1)
	rl.Scalef(scale.X, scale.Y, 1)
	paint()
}
