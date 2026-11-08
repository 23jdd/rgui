package main

import (
	"runtime"

	"rgui/draw"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	rl.SetConfigFlags(rl.FlagWindowHighdpi | rl.FlagMsaa4xHint)
	rl.InitWindow(960, 650, "rgui primitives")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	image := rl.GenImageChecked(48, 48, 8, 8, rl.DarkBlue, rl.SkyBlue)
	texture := rl.LoadTextureFromImage(image)
	rl.UnloadImage(image)
	defer rl.UnloadTexture(texture)

	label := draw.TextStyle{Font: rl.GetFontDefault(), Size: 18, Spacing: 1, Color: rl.RayWhite}
	body := draw.TextStyle{Font: label.Font, Size: 16, Spacing: 1, Color: rl.LightGray}
	wrapped := draw.TextBlock("This paragraph wraps at the measured width. The same lines are used for painting.", 155, 5, body)
	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.NewColor(23, 28, 38, 255))
		var canvas draw.Canvas
		draw.Text(point(24, 16), "rgui drawing primitives", draw.TextStyle{Font: label.Font, Size: 26, Spacing: 1, Color: rl.RayWhite})
		panel(24, 60, "1  Basic shapes and text", label, func() { basics(label) })
		panel(492, 60, "2  Geometry and paths", label, geometry)
		panel(24, 354, "3  Text, textures and scopes", label, func() { content(&canvas, texture, wrapped, body) })
		panel(492, 354, "4  Gradients, shadow and dashes", label, effects)
		rl.EndDrawing()
	}
}

func point(x, y float32) rl.Vector2 { return rl.Vector2{X: x, Y: y} }

func rect(x, y, w, h float32) rl.Rectangle {
	return rl.Rectangle{X: x, Y: y, Width: w, Height: h}
}

func panel(x, y float32, title string, label draw.TextStyle, paint func()) {
	draw.Rect(rect(x, y, 444, 270), 12, draw.Style{Fill: rl.NewColor(34, 41, 54, 255)})
	draw.Text(point(x+16, y+16), title, label)
	paint()
}

func basics(label draw.TextStyle) {
	stroke := draw.Stroke{Color: rl.SkyBlue, Width: 3}
	draw.Rect(rect(54, 126, 155, 72), 16, draw.Style{Fill: rl.DarkBlue, Stroke: stroke})
	draw.Circle(point(308, 163), 36, draw.Style{Fill: rl.DarkPurple, Stroke: draw.Stroke{Color: rl.Pink, Width: 3}})
	draw.Line(point(54, 244), point(197, 216), stroke)
	draw.Text(point(234, 235), "Measured text", label)
	size := draw.MeasureText("Measured text", label)
	draw.Rect(rect(230, 231, size.X+8, size.Y+8), 0, draw.Style{Stroke: draw.Stroke{Color: rl.Gray, Width: 1}})
}

func geometry() {
	outline := draw.Stroke{Color: rl.SkyBlue, Width: 3}
	draw.Ellipse(point(563, 153), 44, 25, draw.Style{Fill: rl.DarkBlue, Stroke: outline})
	draw.Arc(point(687, 154), 39, 30, 25, 305, draw.Stroke{Color: rl.Gold, Width: 5})
	draw.Ring(point(824, 151), 21, 36, -75, 225, rl.Pink)
	draw.Polyline([]rl.Vector2{point(520, 241), point(551, 209), point(582, 239), point(611, 210)}, false, outline)
	draw.Polygon([]rl.Vector2{point(632, 243), point(663, 207), point(696, 243), point(665, 230)}, draw.Style{Fill: rl.Pink})

	var path draw.Path
	path.MoveTo(point(718, 236))
	path.LineTo(point(740, 236))
	path.QuadTo(point(759, 186), point(780, 236))
	path.CubicTo(point(804, 204), point(828, 186), point(865, 236))
	path.Close()
	path.Draw(draw.Style{Fill: rl.NewColor(87, 133, 210, 120), Stroke: outline})

	path.MoveTo(point(723, 283))
	path.QuadTo(point(749, 248), point(775, 283))
	path.CubicTo(point(799, 313), point(830, 253), point(860, 283))
	path.Draw(draw.Style{Stroke: draw.Stroke{Color: rl.Gold, Width: 3}})
}

func content(canvas *draw.Canvas, texture rl.Texture2D, wrapped draw.TextLayout, body draw.TextStyle) {
	draw.Text(point(49, 402), "Wrapped text", body)
	wrapped.Draw(point(49, 427))
	draw.Text(point(222, 402), "Image crop", body)
	draw.Image(texture, rect(8, 8, 24, 24), rect(223, 426, 78, 50), rl.White)
	draw.Text(point(325, 402), "NinePatch", body)
	patch := rl.NPatchInfo{Source: rect(0, 0, 48, 48), Left: 8, Top: 8, Right: 8, Bottom: 8}
	draw.NinePatch(texture, patch, rect(325, 426, 112, 50), rl.White)

	draw.Text(point(49, 551), "Nested clips + transforms", body)
	canvas.Clip(rect(257, 505, 165, 82), func() {
		draw.Rect(rect(257, 505, 165, 82), 0, draw.Style{Fill: rl.DarkBlue})
		canvas.Clip(rect(287, 520, 158, 48), func() {
			canvas.Transform(point(335, 536), 27, point(1.2, 1), func() {
				canvas.Transform(point(-49, -20), 0, point(1, 1), func() {
					draw.Rect(rect(0, 0, 127, 43), 5, draw.Style{Fill: rl.Pink})
				})
			})
		})
		draw.Line(point(257, 576), point(443, 576), draw.Stroke{Color: rl.SkyBlue, Width: 3})
	})
	draw.Circle(point(435, 576), 6, draw.Style{Fill: rl.Gold})
}

func effects() {
	draw.LinearGradient(rect(520, 412, 152, 62), 35, rl.DarkPurple, rl.SkyBlue)
	draw.RadialGradient(point(777, 446), 43, rl.Gold, rl.NewColor(216, 70, 140, 0))
	box := rect(536, 517, 111, 57)
	draw.Shadow(box, 11, point(8, 9), 13, rl.Black)
	draw.Rect(box, 11, draw.Style{Fill: rl.DarkPurple})
	draw.Polyline([]rl.Vector2{point(689, 568), point(735, 518), point(781, 564), point(846, 519)}, false,
		draw.Stroke{Color: rl.Gold, Width: 4, Dash: 12, Gap: 7})
}
