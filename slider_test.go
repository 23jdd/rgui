package rgui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestSliderCaptureAndKeyboard(t *testing.T) {
	value := NewSignal(0.0)
	slider := Slider(value, 0, 100, 10).(*slider)
	root := componentOf(slider)
	root.mount(make(subscriptions))
	defer root.unmount()
	testLayout(root, 100, 34)
	var router inputRouter
	router.pointer(PointerDown, root, rl.Vector2{X: 53, Y: 17}, rl.Vector2{})
	if value.Get() != 50 || router.capture != root {
		t.Fatal("pointer should snap and capture")
	}
	router.pointer(PointerMove, nil, rl.Vector2{X: 200, Y: 17}, rl.Vector2{})
	router.pointer(PointerUp, nil, rl.Vector2{X: 200, Y: 17}, rl.Vector2{})
	if value.Get() != 100 || router.capture != nil {
		t.Fatal("drag should clamp and release")
	}
	slider.KeyDown(&InputEvent{Source: slider, Key: rl.KeyLeft})
	if value.Get() != 90 {
		t.Fatal("arrow key should move by one step")
	}
}
