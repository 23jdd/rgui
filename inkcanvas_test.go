package rgui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"testing"
)

func TestInkCanvasCommitsOnlyCompletedStrokes(t *testing.T) {
	strokes := NewSignal([]InkStroke(nil))
	root := componentOf(InkCanvas(strokes, rl.SkyBlue, 2))
	root.mount(make(subscriptions))
	defer root.unmount()
	testLayout(root, 100, 100)
	var input inputRouter
	point := rl.Vector2{X: 5, Y: 5}
	input.pointer(PointerDown, root, point, rl.Vector2{})
	input.cancel()
	if len(strokes.Get()) != 0 {
		t.Fatal("canceled stroke should be discarded")
	}
	input.pointer(PointerDown, root, point, rl.Vector2{})
	input.pointer(PointerUp, nil, rl.Vector2{X: 110, Y: 20}, rl.Vector2{})
	got := strokes.Get()
	if len(got) != 1 || len(got[0].Points) != 2 || got[0].Points[1].X != 100 {
		t.Fatalf("stroke = %+v", got)
	}
}
