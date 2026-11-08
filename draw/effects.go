package draw

import (
	"image/color"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// LinearGradient fills a rectangle along the given clockwise angle in degrees.
// The stops are at the minimum and maximum projections of its corners.
func LinearGradient(bounds rl.Rectangle, degrees float32, from, to color.RGBA) {
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	angle := float64(degrees) * math.Pi / 180
	dx, dy := math.Cos(angle), math.Sin(angle)
	points := roundedOutline(bounds, 0)
	x, y := float64(bounds.Width)*dx, float64(bounds.Height)*dy
	low, high := min(0, x)+min(0, y), max(0, x)+max(0, y)
	projections := [4]float64{0, x, x + y, y}

	rl.SetTexture(0)
	rl.Begin(rl.Triangles)
	for _, i := range [...]int{0, 3, 2, 0, 2, 1} {
		t := (projections[i] - low) / (high - low)
		col := mixColor(from, to, t)
		rl.Color4ub(col.R, col.G, col.B, col.A)
		rl.Vertex2f(points[i].X, points[i].Y)
	}
	rl.End()
}

func mixColor(a, b color.RGBA, t float64) color.RGBA {
	mix := func(from, to uint8) uint8 {
		return uint8(math.Round(float64(from) + (float64(to)-float64(from))*t))
	}
	return color.RGBA{R: mix(a.R, b.R), G: mix(a.G, b.G), B: mix(a.B, b.B), A: mix(a.A, b.A)}
}

// RadialGradient fills a circle from its center color to its edge color.
func RadialGradient(center rl.Vector2, radius float32, inner, outer color.RGBA) {
	if radius > 0 {
		rl.DrawCircleGradient(center, radius, inner, outer)
	}
}

// Shadow draws a rounded rectangular shadow. Blur is a feather radius in pixels.
// Paint the foreground shape after its shadow.
func Shadow(bounds rl.Rectangle, radius float32, offset rl.Vector2, blur float32, tint color.RGBA) {
	if bounds.Width <= 0 || bounds.Height <= 0 || tint.A == 0 {
		return
	}
	bounds.X += offset.X
	bounds.Y += offset.Y
	if blur <= 0 {
		Rect(bounds, radius, Style{Fill: tint})
		return
	}
	const layers = 16
	opacity := float64(tint.A) / 255
	var previous float64
	for i := layers; i >= 0; i-- {
		spread := blur * float32(i) / layers
		t := 1 - float64(i)/layers
		target := opacity * t * t * (3 - 2*t)
		tint.A = uint8(math.Round(255 * (target - previous) / (1 - previous)))
		previous = target
		Rect(rl.Rectangle{
			X: bounds.X - spread, Y: bounds.Y - spread,
			Width: bounds.Width + 2*spread, Height: bounds.Height + 2*spread,
		}, max(0, radius)+spread, Style{Fill: tint})
	}
}
