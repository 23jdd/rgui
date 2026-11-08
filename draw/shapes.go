// Package draw provides 2D primitives for use inside a raylib drawing pass.
package draw

import (
	"image/color"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Stroke describes a solid or dashed outline. Dash and Gap are lengths in pixels.
type Stroke struct {
	Color color.RGBA
	Width float32
	Dash  float32
	Gap   float32
}

// Style describes a shape's fill and outline. Transparent colors omit the paint.
type Style struct {
	Fill   color.RGBA
	Stroke Stroke
}

// Rect draws a rectangle with a corner radius in pixels.
func Rect(bounds rl.Rectangle, radius float32, style Style) {
	side := min(bounds.Width, bounds.Height)
	if side <= 0 {
		return
	}
	roundness := min(1, max(0, 2*radius/side))
	if style.Fill.A != 0 {
		if roundness == 0 {
			rl.DrawRectangleRec(bounds, style.Fill)
		} else {
			rl.DrawRectangleRounded(bounds, roundness, 12, style.Fill)
		}
	}
	stroke := style.Stroke
	if stroke.Width > 0 && stroke.Color.A != 0 {
		Polyline(roundedOutline(bounds, roundness*side/2), true, stroke)
	}
}

func roundedOutline(r rl.Rectangle, radius float32) []rl.Vector2 {
	right, bottom := r.X+r.Width, r.Y+r.Height
	if radius == 0 {
		return []rl.Vector2{{X: r.X, Y: r.Y}, {X: right, Y: r.Y}, {X: right, Y: bottom}, {X: r.X, Y: bottom}}
	}
	centers := []rl.Vector2{
		{X: right - radius, Y: r.Y + radius}, {X: right - radius, Y: bottom - radius},
		{X: r.X + radius, Y: bottom - radius}, {X: r.X + radius, Y: r.Y + radius},
	}
	var outline []rl.Vector2
	for i, center := range centers {
		start := float32(i*90 - 90)
		outline = append(outline, arcPoints(center, radius, radius, start, start+90)...)
	}
	return outline
}

// Circle draws a circle centered at center.
func Circle(center rl.Vector2, radius float32, style Style) {
	Ellipse(center, radius, radius, style)
}

// Ellipse draws an ellipse with separate horizontal and vertical radii.
func Ellipse(center rl.Vector2, rx, ry float32, style Style) {
	if rx <= 0 || ry <= 0 {
		return
	}
	if style.Fill.A != 0 {
		rl.DrawEllipseV(center, rx, ry, style.Fill)
	}
	if style.Stroke.Width > 0 && style.Stroke.Color.A != 0 {
		points := arcPoints(center, rx, ry, 0, 360)
		Polyline(points[:len(points)-1], true, style.Stroke)
	}
}

// Arc draws an elliptical arc. Angles are clockwise degrees from the right.
func Arc(center rl.Vector2, rx, ry, start, end float32, stroke Stroke) {
	if rx > 0 && ry > 0 && start != end {
		Polyline(arcPoints(center, rx, ry, start, end), false, stroke)
	}
}

// Ring fills a circular band between two radii. Angles are clockwise degrees.
func Ring(center rl.Vector2, inner, outer, start, end float32, fill color.RGBA) {
	if inner >= 0 && outer > inner && start != end && fill.A != 0 {
		rl.DrawRing(center, inner, outer, start, start+sweep(start, end), 0, fill)
	}
}

func sweep(start, end float32) float32 {
	return max(-360, min(360, end-start))
}

func arcPoints(center rl.Vector2, rx, ry, start, end float32) []rl.Vector2 {
	span := sweep(start, end)
	steps := max(1, int(math.Ceil(math.Abs(float64(span))/5)))
	points := make([]rl.Vector2, steps+1)
	for i := range points {
		angle := float64(start+span*float32(i)/float32(steps)) * math.Pi / 180
		points[i] = rl.Vector2{
			X: center.X + rx*float32(math.Cos(angle)),
			Y: center.Y + ry*float32(math.Sin(angle)),
		}
	}
	return points
}

// Line draws a single segment.
func Line(from, to rl.Vector2, stroke Stroke) {
	Polyline([]rl.Vector2{from, to}, false, stroke)
}

// Polyline connects points, optionally joining the last point to the first.
// A dash pattern continues across segment boundaries.
func Polyline(points []rl.Vector2, closed bool, stroke Stroke) {
	if len(points) < 2 || stroke.Width <= 0 || stroke.Color.A == 0 {
		return
	}
	walkStroke(points, closed, stroke.Dash, stroke.Gap, func(from, to rl.Vector2) {
		rl.DrawLineEx(from, to, stroke.Width, stroke.Color)
	})
}

func walkStroke(points []rl.Vector2, closed bool, dash, gap float32, emit func(rl.Vector2, rl.Vector2)) {
	segments := len(points) - 1
	if closed {
		segments++
	}
	dashed := dash > 0 && gap > 0
	phase, remaining := true, float64(dash)
	for i := range segments {
		from, to := points[i], points[(i+1)%len(points)]
		length := math.Hypot(float64(to.X-from.X), float64(to.Y-from.Y))
		if length == 0 {
			continue
		}
		if !dashed {
			emit(from, to)
			continue
		}
		for offset := 0.0; offset < length; {
			step := min(remaining, length-offset)
			if phase {
				emit(lerp(from, to, float32(offset/length)), lerp(from, to, float32((offset+step)/length)))
			}
			if offset+step == offset {
				break
			}
			offset += step
			remaining -= step
			if remaining <= 0 {
				phase = !phase
				remaining = float64(gap)
				if phase {
					remaining = float64(dash)
				}
			}
		}
	}
}

func lerp(a, b rl.Vector2, t float32) rl.Vector2 {
	return rl.Vector2{X: a.X + (b.X-a.X)*t, Y: a.Y + (b.Y-a.Y)*t}
}
