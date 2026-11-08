package draw

import rl "github.com/gen2brain/raylib-go/raylib"

// Path holds one contour. Its zero value starts at (0, 0).
type Path struct {
	points []rl.Vector2
	closed bool
}

// MoveTo starts a new contour, replacing the previous one.
func (p *Path) MoveTo(start rl.Vector2) {
	p.points = append(p.points[:0], start)
	p.closed = false
}

// LineTo adds a straight segment.
func (p *Path) LineTo(end rl.Vector2) {
	p.start()
	p.points = append(p.points, end)
}

// QuadTo adds a quadratic Bézier segment.
func (p *Path) QuadTo(control, end rl.Vector2) {
	p.start()
	from := p.points[len(p.points)-1]
	const steps = 24
	for i := 1; i <= steps; i++ {
		t := float32(i) / steps
		p.points = append(p.points, lerp(lerp(from, control, t), lerp(control, end, t), t))
	}
}

// CubicTo adds a cubic Bézier segment.
func (p *Path) CubicTo(first, second, end rl.Vector2) {
	p.start()
	from := p.points[len(p.points)-1]
	const steps = 32
	for i := 1; i <= steps; i++ {
		t := float32(i) / steps
		a, b, c := lerp(from, first, t), lerp(first, second, t), lerp(second, end, t)
		p.points = append(p.points, lerp(lerp(a, b, t), lerp(b, c, t), t))
	}
}

func (p *Path) start() {
	if len(p.points) == 0 {
		p.MoveTo(rl.Vector2{})
	}
}

// Close joins the end to the start so the contour can be filled.
func (p *Path) Close() { p.closed = true }

// Draw paints the path. Open contours only use the stroke.
func (p *Path) Draw(style Style) {
	if p.closed {
		Polygon(p.points, style)
		return
	}
	Polyline(p.points, false, style.Stroke)
}

// Polygon fills a simple closed polygon, including concave polygons in either
// winding order. Self-intersecting polygons and holes are outside its scope.
func Polygon(points []rl.Vector2, style Style) {
	if style.Fill.A != 0 {
		for _, triangle := range triangulate(points) {
			rl.DrawTriangle(triangle[0], triangle[1], triangle[2], style.Fill)
		}
	}
	Polyline(points, true, style.Stroke)
}

func cross(a, b, c rl.Vector2) float64 {
	return (float64(b.X)-float64(a.X))*(float64(c.Y)-float64(a.Y)) -
		(float64(b.Y)-float64(a.Y))*(float64(c.X)-float64(a.X))
}

func triangulate(points []rl.Vector2) [][3]rl.Vector2 {
	vertices := make([]rl.Vector2, 0, len(points))
	for _, point := range points {
		if len(vertices) == 0 || point != vertices[len(vertices)-1] {
			vertices = append(vertices, point)
		}
	}
	if len(vertices) > 1 && vertices[0] == vertices[len(vertices)-1] {
		vertices = vertices[:len(vertices)-1]
	}
	for i := 0; len(vertices) >= 3 && i < len(vertices); {
		n := len(vertices)
		if cross(vertices[(i+n-1)%n], vertices[i], vertices[(i+1)%n]) == 0 {
			vertices = append(vertices[:i], vertices[i+1:]...)
			i = max(0, i-1)
		} else {
			i++
		}
	}
	if len(vertices) < 3 {
		return nil
	}

	var area float64
	for i := 1; i+1 < len(vertices); i++ {
		area += cross(vertices[0], vertices[i], vertices[i+1])
	}
	if area == 0 {
		return nil
	}
	// Raylib expects counter-clockwise winding in screen coordinates.
	if area > 0 {
		for i, j := 0, len(vertices)-1; i < j; i, j = i+1, j-1 {
			vertices[i], vertices[j] = vertices[j], vertices[i]
		}
	}

	triangles := make([][3]rl.Vector2, 0, len(vertices)-2)
	for len(vertices) > 3 {
		i := findEar(vertices)
		if i < 0 {
			return nil
		}
		n := len(vertices)
		triangles = append(triangles, [3]rl.Vector2{vertices[(i+n-1)%n], vertices[i], vertices[(i+1)%n]})
		vertices = append(vertices[:i], vertices[i+1:]...)
	}
	return append(triangles, [3]rl.Vector2{vertices[0], vertices[1], vertices[2]})
}

func findEar(vertices []rl.Vector2) int {
	n := len(vertices)
	for i, b := range vertices {
		prev, next := (i+n-1)%n, (i+1)%n
		a, c := vertices[prev], vertices[next]
		if cross(a, b, c) >= 0 {
			continue
		}
		clear := true
		for j, p := range vertices {
			if j != i && j != prev && j != next && cross(a, b, p) <= 0 && cross(b, c, p) <= 0 && cross(c, a, p) <= 0 {
				clear = false
				break
			}
		}
		if clear {
			return i
		}
	}
	return -1
}
