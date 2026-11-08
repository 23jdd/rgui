package rgui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"testing"
)

func TestListBoxSelectionAndScrollIntoView(t *testing.T) {
	selected := NewSignal(-1)
	root := componentOf(ListBox([]string{"A", "B", "C", "D", "E"}, selected))
	pending := make(subscriptions)
	root.mount(pending)
	defer root.unmount()
	testLayout(root, 120, 68)
	list := root.children[0].custom.(*listBox)
	list.Tapped(&InputEvent{Source: list, Position: rl.Vector2{Y: 45}})
	if selected.Get() != 1 {
		t.Fatal("expected second row selected")
	}
	list.KeyDown(&InputEvent{Source: list, Key: rl.KeyEnd})
	flushBindings(pending)
	testLayout(root, 120, 68)
	if selected.Get() != 4 || list.Bounds().Y != -102 {
		t.Fatalf("index=%d bounds=%+v", selected.Get(), list.Bounds())
	}
}
