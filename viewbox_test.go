package rgui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"testing"
)

func TestViewboxScalesHitAndLocalCoordinates(t *testing.T) {
	child := WithLayout(Button("button", func() {}), Box{Width: 80, Height: 40})
	root := componentOf(Viewbox(child, ImageUniform))
	testLayout(root, 200, 200)
	button := componentOf(child)
	if got := button.visualRect(button.bounds); got != (rl.Rectangle{Y: 50, Width: 200, Height: 100}) {
		t.Fatalf("visual bounds=%+v", got)
	}
	point := rl.Vector2{X: 100, Y: 100}
	if hitNode(root, point) != button || button.localPoint(point) != (rl.Vector2{X: 40, Y: 20}) || hitNode(root, rl.Vector2{X: 100, Y: 25}) == button {
		t.Fatal("input should match the transformed child")
	}
}
