package rgui

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type windowViewport struct {
	size, scale   rl.Vector2
	width, height int32
}

func viewportForPixels(width, height int32, scale rl.Vector2) windowViewport {
	if scale.X <= 0 {
		scale.X = 1
	}
	if scale.Y <= 0 {
		scale.Y = 1
	}
	return windowViewport{
		width: width, height: height, scale: scale,
		size: rl.Vector2{X: float32(width) / scale.X, Y: float32(height) / scale.Y},
	}
}

func snapBounds(bounds rl.Rectangle, scale rl.Vector2) rl.Rectangle {
	if scale.X <= 0 || scale.Y <= 0 {
		return bounds
	}
	edge := func(value, factor float32) float32 {
		return float32(math.Round(float64(value*factor))) / factor
	}
	x, y := edge(bounds.X, scale.X), edge(bounds.Y, scale.Y)
	right := edge(bounds.X+bounds.Width, scale.X)
	bottom := edge(bounds.Y+bounds.Height, scale.Y)
	return rl.Rectangle{X: x, Y: y, Width: max(0, right-x), Height: max(0, bottom-y)}
}
