package rgui

import (
	"image/color"
	"math"
	"time"

	"rgui/draw"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// FrameStats reports work done for one frame. Present includes frame pacing
// and platform event polling in raylib's EndDrawing.
type FrameStats struct {
	Layout, Paint, Present time.Duration
	Painted                bool
	Damage                 rl.Rectangle // Logical coordinates.
	PaintedPixels          int64
}

// InvalidateVisual marks the node's current bounds for repaint.
func (n *BaseNode) InvalidateVisual() {
	n.InvalidateRect(rl.Rectangle{Width: n.bounds.Width, Height: n.bounds.Height})
}

// InvalidateRect marks a rectangle relative to the node's Bounds for repaint.
func (n *BaseNode) InvalidateRect(local rl.Rectangle) {
	if n.host == nil {
		return
	}
	local.X += n.bounds.X
	local.Y += n.bounds.Y
	n.host.visualDamage = unionRect(n.host.visualDamage, intersectRect(local, n.bounds))
}

func nonempty(r rl.Rectangle) bool { return r.Width > 0 && r.Height > 0 }

func unionRect(a, b rl.Rectangle) rl.Rectangle {
	if !nonempty(a) {
		return b
	}
	if !nonempty(b) {
		return a
	}
	x, y := min(a.X, b.X), min(a.Y, b.Y)
	return rl.Rectangle{X: x, Y: y, Width: max(a.X+a.Width, b.X+b.Width) - x, Height: max(a.Y+a.Height, b.Y+b.Height) - y}
}

func (c *component) takeVisualDamage(ancestorDisabled bool) (rl.Rectangle, bool) {
	damage := c.visualRect(c.visualDamage)
	c.visualDamage = rl.Rectangle{}
	disabled := ancestorDisabled || c.layoutBase().Input.Disabled || c.button && c.onClick == nil
	changed := disabled != c.lastDisabled
	if changed {
		damage = unionRect(damage, c.visualRect(c.bounds))
		c.lastDisabled = disabled
	}
	for _, child := range c.children {
		childDamage, childChanged := child.takeVisualDamage(disabled)
		damage = unionRect(damage, childDamage)
		changed = changed || childChanged
	}
	return damage, changed
}

type paintState struct{ hovered, pressed, focused *component }

func (s paintState) damage(previous paintState) rl.Rectangle {
	var damage rl.Rectangle
	for _, pair := range [][2]*component{{s.hovered, previous.hovered}, {s.pressed, previous.pressed}, {s.focused, previous.focused}} {
		if pair[0] == pair[1] {
			continue
		}
		for _, node := range pair {
			if node != nil {
				damage = unionRect(damage, node.visualRect(node.bounds))
			}
		}
	}
	return damage
}

type retainedFrame struct {
	texture rl.RenderTexture2D
	window  windowViewport
}

func (f *retainedFrame) close() {
	if f.texture.ID != 0 {
		rl.UnloadRenderTexture(f.texture)
		f.texture = rl.RenderTexture2D{}
	}
}

func (f *retainedFrame) prepare(window windowViewport) (ready, fresh bool) {
	if f.texture.ID != 0 && f.window.width == window.width && f.window.height == window.height {
		fresh = f.window != window
		f.window = window
		return true, fresh
	}
	f.close()
	f.window = window
	if window.width <= 0 || window.height <= 0 {
		return false, false
	}
	f.texture = rl.LoadRenderTexture(window.width, window.height)
	if !rl.IsRenderTextureValid(f.texture) {
		f.close()
		return false, false
	}
	return true, true
}

func (f *retainedFrame) update(damage rl.Rectangle, background color.RGBA, paint func(*draw.Canvas)) {
	rl.BeginTextureMode(f.texture)
	defer rl.EndTextureMode()
	canvas := draw.NewCanvas(f.window.scale, f.window.height)
	canvas.Clip(damage, func() {
		rl.ClearBackground(background)
		rl.PushMatrix()
		defer rl.PopMatrix()
		rl.Scalef(f.window.scale.X, f.window.scale.Y, 1)
		paint(&canvas)
		rl.DrawRenderBatchActive()
	})
}

func (f *retainedFrame) present() {
	rl.DrawRenderBatchActive()
	rl.DisableColorBlend()
	defer rl.EnableColorBlend()
	rl.DrawTexturePro(f.texture.Texture,
		rl.Rectangle{Width: float32(f.window.width), Height: -float32(f.window.height)},
		rl.Rectangle{Width: f.window.size.X, Height: f.window.size.Y}, rl.Vector2{}, 0, rl.White)
	rl.DrawRenderBatchActive()
}

func (window windowViewport) fullDamage() rl.Rectangle {
	return rl.Rectangle{Width: window.size.X, Height: window.size.Y}
}

func (window windowViewport) pixelDamage(damage rl.Rectangle) rl.Rectangle {
	// Expand outwards so fractional dirty edges never miss a touched pixel.
	x := float32(math.Floor(float64(damage.X * window.scale.X)))
	y := float32(math.Floor(float64(damage.Y * window.scale.Y)))
	right := float32(math.Ceil(float64((damage.X + damage.Width) * window.scale.X)))
	bottom := float32(math.Ceil(float64((damage.Y + damage.Height) * window.scale.Y)))
	return intersectRect(rl.Rectangle{X: x, Y: y, Width: right - x, Height: bottom - y},
		rl.Rectangle{Width: float32(window.width), Height: float32(window.height)})
}

func (window windowViewport) logicalDamage(pixels rl.Rectangle) rl.Rectangle {
	return rl.Rectangle{X: pixels.X / window.scale.X, Y: pixels.Y / window.scale.Y,
		Width: pixels.Width / window.scale.X, Height: pixels.Height / window.scale.Y}
}
