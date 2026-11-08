package rgui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"testing"
)

func TestGridSplitterResizesTrack(t *testing.T) {
	size := NewSignal(float32(45))
	left := Text(nil)
	splitter := GridSplitter(Horizontal, size, 20, 90)
	root := componentOf(Grid(nil, []Track{FixedFrom(size), Fixed(8), Star(1)}, Cell{Child: left}, Cell{Column: 1, Child: splitter}))
	pending := make(subscriptions)
	root.mount(pending)
	defer root.unmount()
	testLayout(root, 150, 50)
	checkBounds(t, left, 0, 0, 45, 50)
	routeEvent(componentOf(splitter), &InputEvent{Kind: KeyDown, Key: rl.KeyRight})
	flushBindings(pending)
	testLayout(root, 150, 50)
	checkBounds(t, left, 0, 0, 53, 50)
}
