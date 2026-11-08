package rgui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestScrollViewerClipsAndClamps(t *testing.T) {
	button := WithLayout(Button("Tall", func() {}), Box{Height: 300})
	offset := NewSignal(rl.Vector2{})
	root := componentOf(ScrollViewer(button, offset, false, true))
	pending := make(subscriptions)
	root.mount(pending)
	defer root.unmount()
	testLayout(root, 100, 100)
	if root.buttonAt(rl.Vector2{X: 10, Y: 101}) != nil {
		t.Fatal("overflow should not receive input")
	}
	routeEvent(componentOf(button), &InputEvent{Kind: PointerWheel, Wheel: rl.Vector2{Y: -10}})
	flushBindings(pending)
	testLayout(root, 100, 100)
	checkBounds(t, button, 0, -200, 100, 300)
	if offset.Get().Y != 200 {
		t.Fatalf("offset = %+v", offset.Get())
	}
}
