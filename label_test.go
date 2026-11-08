package rgui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"testing"
)

func TestLabelMnemonicMovesFocusToItsTarget(t *testing.T) {
	box := TextBox(nil)
	root := componentOf(StackPanel(Vertical, 0, Label("_Name", box), box))
	root.mount(make(subscriptions))
	defer root.unmount()
	testLayout(root, 200, 100)
	var focus focusState
	if !focusAccessKey(root, &focus, rl.Vector2{X: 200, Y: 100}, 'n') || focus.button != componentOf(box) {
		t.Fatal("mnemonic should focus the associated input")
	}
}
