package rgui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestToolTipLeavesInputTransparent(t *testing.T) {
	button := Button("Target", func() {})
	root := componentOf(ToolTip(button, "Hint"))
	pending := make(subscriptions)
	root.mount(pending)
	defer root.unmount()
	var input inputRouter
	input.updateHover(componentOf(button), rl.Vector2{})
	flushBindings(pending)
	layers := popupLayers(root, nil)
	if len(layers) != 1 || !layers[0].children[0].Input.HitTestInvisible {
		t.Fatal("hint should be open and input-transparent")
	}
	input.updateHover(nil, rl.Vector2{})
	flushBindings(pending)
	if len(popupLayers(root, nil)) != 0 {
		t.Fatal("hint should close when the pointer leaves")
	}
}
