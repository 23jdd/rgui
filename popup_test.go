package rgui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"testing"
)

func TestPopupOverlaysLayoutAndHitOrder(t *testing.T) {
	anchor := WithLayout(Button("Open", func() {}), Box{Height: 30})
	behind := WithLayout(Button("Behind", func() {}), Box{Height: 50})
	above := WithLayout(Button("Above", func() {}), Box{Height: 40})
	open := NewSignal(true)
	root := componentOf(StackPanel(Vertical, 0, anchor, Popup(above, open), behind))
	pending := make(subscriptions)
	root.mount(pending)
	defer root.unmount()
	testLayout(root, 120, 140)
	checkBounds(t, behind, 0, 30, 120, 50)
	checkBounds(t, above, 0, 30, 120, 40)
	point, viewport := rl.Vector2{X: 20, Y: 42}, rl.Vector2{X: 120, Y: 140}
	if hitWindow(root, point, viewport) != componentOf(above) {
		t.Fatal("overlay must receive the hit")
	}
	open.Set(false)
	flushBindings(pending)
	testLayout(root, 120, 140)
	if hitWindow(root, point, viewport) != componentOf(behind) || componentOf(above).pending != nil {
		t.Fatal("closing popup should detach it and restore hit testing")
	}
}
