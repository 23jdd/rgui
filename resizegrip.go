package rgui

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ResizeGrip is a bottom-right drag handle for a resizable native window.
func ResizeGrip() Node {
	var start, scale rl.Vector2
	var resizable bool
	handle := Thumb(DragCallbacks{
		Started: func(rl.Vector2) {
			resizable = rl.IsWindowState(rl.FlagWindowResizable)
			scale = resizeScale()
			start = rl.Vector2{X: float32(rl.GetScreenWidth()), Y: float32(rl.GetScreenHeight())}
		},
		Moved: func(delta rl.Vector2) {
			if resizable {
				size := resizeDimensions(start, delta, scale)
				rl.SetWindowSize(int(size.X), int(size.Y))
			}
		},
	}).(*thumb)
	handle.desired = rl.Vector2{X: 18, Y: 18}
	handle.HorizontalAlignment = End
	handle.VerticalAlignment = End
	return styleName(handle, "resizegrip")
}

func resizeDimensions(start, delta, scale rl.Vector2) rl.Vector2 {
	return rl.Vector2{
		X: max(1, float32(math.Round(float64((start.X+delta.X)*scale.X)))),
		Y: max(1, float32(math.Round(float64((start.Y+delta.Y)*scale.Y)))),
	}
}
