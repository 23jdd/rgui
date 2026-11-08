package rgui

import (
	"os"
	"runtime"
	"testing"

	"rgui/draw"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestDamageRoundsOutAndStaysInsideFramebuffer(t *testing.T) {
	window := viewportForPixels(101, 79, rl.Vector2{X: 1.5, Y: 1.5})
	got := window.pixelDamage(rl.Rectangle{X: -2, Y: 1.5, Width: 100, Height: 4})
	want := rl.Rectangle{Y: 2, Width: 101, Height: 7}
	if got != want {
		t.Fatalf("damage = %+v, want %+v", got, want)
	}
}

func TestRetainedFrameMatchesFullPaint(t *testing.T) {
	if os.Getenv("RGUI_WINDOW_TEST") != "1" {
		t.Skip("set RGUI_WINDOW_TEST=1 to test raylib render textures")
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	rl.SetConfigFlags(rl.FlagWindowHidden)
	rl.InitWindow(96, 72, "rgui retained-frame test")
	if !rl.IsWindowReady() {
		t.Fatal("could not initialize raylib")
	}
	defer rl.CloseWindow()
	background := rl.NewColor(20, 30, 40, 170)
	var incremental, full retainedFrame
	defer incremental.close()
	defer full.close()
	for _, window := range []windowViewport{
		viewportForPixels(97, 73, rl.Vector2{X: 1.25, Y: 1.25}),
		viewportForPixels(113, 87, rl.Vector2{X: 1.5, Y: 1.5}),
	} {
		for _, frame := range []*retainedFrame{&incremental, &full} {
			if ready, fresh := frame.prepare(window); !ready || !fresh {
				t.Fatal("expected a new render target")
			}
		}
		paint := func(frame *retainedFrame, damage, moving rl.Rectangle) {
			frame.update(damage, background, func(canvas *draw.Canvas) {
				draw.Rect(rl.Rectangle{X: 5, Y: 6, Width: 35, Height: 28}, 0,
					draw.Style{Fill: rl.NewColor(220, 50, 50, 120)})
				canvas.Clip(rl.Rectangle{X: 10, Y: 8, Width: 35, Height: 28}, func() {
					draw.Rect(moving, 0, draw.Style{Fill: rl.NewColor(40, 90, 230, 160)})
				})
			})
		}
		old := rl.Rectangle{X: 10, Y: 10, Width: 18, Height: 14}
		moved := rl.Rectangle{X: 20, Y: 12, Width: 18, Height: 14}
		paint(&incremental, window.fullDamage(), old)
		if ready, fresh := incremental.prepare(window); !ready || fresh {
			t.Fatal("unchanged viewport should retain the previous texture")
		}
		damage := window.logicalDamage(window.pixelDamage(unionRect(old, moved)))
		paint(&incremental, damage, moved)
		paint(&full, window.fullDamage(), moved)
		checkFramePixels(t, &incremental, &full, background)
	}
}

func checkFramePixels(t *testing.T, incremental, full *retainedFrame, background rl.Color) {
	t.Helper()
	got := rl.LoadImageFromTexture(incremental.texture.Texture)
	want := rl.LoadImageFromTexture(full.texture.Texture)
	if got == nil || want == nil {
		t.Fatal("could not read render textures")
	}
	defer rl.UnloadImage(got)
	defer rl.UnloadImage(want)
	var sawBackground, sawContent bool
	for y := int32(0); y < full.window.height; y++ {
		for x := int32(0); x < full.window.width; x++ {
			a, b := rl.GetImageColor(*got, x, y), rl.GetImageColor(*want, x, y)
			if a != b {
				t.Fatalf("partial and full differ at (%d, %d): %+v vs %+v", x, y, a, b)
			}
			sawBackground = sawBackground || a == background
			sawContent = sawContent || a != background
		}
	}
	if !sawBackground || !sawContent {
		t.Fatal("expected both background and painted content")
	}
}
