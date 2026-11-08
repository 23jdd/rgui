package rgui

import (
	"math"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestFractionalDPISharedEdgesAndHitTesting(t *testing.T) {
	for _, factor := range []float32{1.25, 1.5} {
		left, right := Button("left", func() {}), Button("right", func() {})
		root := componentOf(Grid(nil, []Track{Star(1), Star(2)}, Cell{Child: left}, Cell{Column: 1, Child: right}))
		scale := rl.Vector2{X: factor, Y: factor}
		for _, dimensions := range [][2]int32{{601, 433}, {2560, 1493}, {601, 433}} {
			view := viewportForPixels(dimensions[0], dimensions[1], scale)
			root.layoutScale = view.scale
			testLayout(root, view.size.X, view.size.Y)
			a, b := componentOf(left).bounds, componentOf(right).bounds
			pixel := func(value float32) int32 { return int32(math.Round(float64(value * factor))) }
			if pixel(a.X+a.Width) != pixel(b.X) || pixel(b.X+b.Width) != view.width ||
				pixel(b.Y+b.Height) != view.height || a.X != 0 || a.Y != 0 {
				t.Fatalf("DPI %.2f: layout edges do not cover the framebuffer: %+v, %+v", factor, a, b)
			}
			atEdge := rl.Vector2{X: b.X, Y: 10}
			if windowButtonAt(root, atEdge, view.size) != componentOf(right) ||
				windowButtonAt(root, rl.Vector2{X: b.X - 0.5/factor, Y: 10}, view.size) != componentOf(left) {
				t.Fatalf("DPI %.2f: hit testing disagrees with the shared edge", factor)
			}
		}
	}
}
