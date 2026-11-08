package rgui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"testing"
)

func TestContextMenuOpensAtPointer(t *testing.T) {
	target := Text(func() string { return "target" })
	root := componentOf(ContextMenu(target, MenuItem{Label: "Action"}))
	pending := make(subscriptions)
	root.mount(pending)
	defer root.unmount()
	routeEvent(componentOf(target), &InputEvent{Kind: SecondaryTap, Position: rl.Vector2{X: 10, Y: 20}})
	flushBindings(pending)
	testLayout(root, 300, 200)
	layers := popupLayers(root, nil)
	if len(layers) != 1 || layers[0].children[0].bounds.X != 10 || layers[0].children[0].bounds.Y != 20 {
		t.Fatal("expected popup at pointer")
	}
	routeEvent(componentOf(target), &InputEvent{Kind: SecondaryTap, Position: rl.Vector2{X: 30, Y: 40}})
	flushBindings(pending)
	testLayout(root, 300, 200)
	if bounds := layers[0].children[0].bounds; bounds.X != 30 || bounds.Y != 40 {
		t.Fatalf("open menu did not move to the new pointer: %+v", bounds)
	}
}
