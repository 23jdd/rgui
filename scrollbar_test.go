package rgui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"testing"
)

func TestScrollBarPagesAndPlacesThumb(t *testing.T) {
	value := NewSignal(0.0)
	root := componentOf(ScrollBar(Vertical, value, 0, 100, 25))
	pending := make(subscriptions)
	root.mount(pending)
	defer root.unmount()
	testLayout(root, 16, 100)
	checkBounds(t, root.children[0], 0, 0, 16, 20)
	routeEvent(root, &InputEvent{Kind: Tap, Position: rl.Vector2{Y: 85}})
	flushBindings(pending)
	testLayout(root, 16, 100)
	if value.Get() != 25 {
		t.Fatal("expected a page scroll")
	}
	checkBounds(t, root.children[0], 0, 20, 16, 20)
}
