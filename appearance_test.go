package rgui

import (
	"os"
	"runtime"
	"testing"

	"rgui/draw"
	"rgui/style"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestStylesheetLayoutAndSelectorChanges(t *testing.T) {
	box := WithStyle(TextBox(nil), StyleOptions{ID: "entry"})
	root := componentOf(WithStyle(StackPanel(Vertical, 0, box), StyleOptions{Classes: []string{"form"}}))
	root.sheet = mustStyle(`
  .form > textbox { width: 120px; height: 44px; margin: 5px 10px; font-size: 21px; }
  .form:hover #entry { height: 90px; border-color: #f60; }
`)
	testLayout(root, 200, 200)
	checkBounds(t, box, 40, 5, 120, 44)
	entry := componentOf(box)
	values := style.Resolve(entry.cssTarget(true, paintState{hovered: entry}), standardStyle, root.sheet)
	if values.Get("border-color") != "#f60" || entry.staticStyle.metrics.size != 21 {
		t.Fatalf("unexpected focused theme or metrics: %v, %+v", values, entry.staticStyle.metrics)
	}
	if got := entry.custom.(*textBox).IMECaret().Height; got != 21 {
		t.Fatalf("caret height = %v, want 21", got)
	}
	WithStyle(root, StyleOptions{})
	testLayout(root, 200, 200)
	checkBounds(t, box, 0, 0, 200, 42)
}

func TestStylesheetPaintsHoveredDescendant(t *testing.T) {
	if os.Getenv("RGUI_WINDOW_TEST") != "1" {
		t.Skip("set RGUI_WINDOW_TEST=1 to check styled raylib output")
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	rl.SetConfigFlags(rl.FlagWindowHidden)
	rl.InitWindow(96, 72, "rgui CSS test")
	if !rl.IsWindowReady() {
		t.Fatal("could not initialize raylib")
	}
	defer rl.CloseWindow()
	button := componentOf(Button("", func() {}))
	root := componentOf(StackPanel(Vertical, 0, button))
	root.sheet = mustStyle(`
  :root { background: #101820; }
  button { background: #123456; border: 2px solid #fedcba; border-radius: 0; }
  stackpanel:hover > button { background: #987654; }
`)
	testLayout(root, 96, 72)
	font := rl.GetFontDefault()
	window := viewportForPixels(96, 72, rl.Vector2{X: 1, Y: 1})
	var frame retainedFrame
	defer frame.close()
	if ready, _ := frame.prepare(window); !ready {
		t.Fatal("could not create a render texture")
	}
	frame.update(window.fullDamage(), windowBackground(root.sheet), func(canvas *draw.Canvas) {
		root.paintScene(canvas, font, paintState{hovered: button}, nil)
	})
	image := rl.LoadImageFromTexture(frame.texture.Texture)
	if image == nil {
		t.Fatal("could not read render texture")
	}
	defer rl.UnloadImage(image)
	for _, sample := range []struct {
		x, y int32
		want rl.Color
	}{
		{48, 71 - 20, rl.NewColor(152, 118, 84, 255)},
		{48, 71 - 1, rl.NewColor(254, 220, 186, 255)},
		{48, 71 - 60, rl.NewColor(16, 24, 32, 255)},
	} {
		if got := rl.GetImageColor(*image, sample.x, sample.y); got != sample.want {
			t.Errorf("pixel (%d, %d) = %+v, want %+v", sample.x, sample.y, got, sample.want)
		}
	}
}
