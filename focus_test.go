package rgui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestFocusTraversal(t *testing.T) {
	first := Button("First", func() {})
	disabled := Button("Disabled", func() {}).WithDisabled(true)
	last := Button("Last", func() {})
	outside := Button("Outside", func() {})
	root := StackPanel(Vertical, 0, first, Stack(disabled, last), outside)
	viewport := rl.Vector2{X: 200, Y: 144}
	testLayout(root, viewport.X, viewport.Y)
	var focus focusState
	for _, step := range []struct {
		backward bool
		want     Element
	}{
		{false, first},
		{false, last},
		{false, first},
		{true, last},
		{true, first},
	} {
		focus.move(root.component, viewport, step.backward)
		if focus.button != step.want.component {
			t.Fatalf("unexpected focus, want %s", step.want.component.text)
		}
	}
	if root.component.buttonAt(rl.Vector2{X: 20, Y: 60}) != nil {
		t.Fatal("disabled button should not accept mouse input")
	}
}

func TestKeyboardActivationAndCancellation(t *testing.T) {
	count := 0
	button := Button("Count", func() { count++ })
	var focus focusState
	focus.set(button.component)
	focus.activate(rl.KeySpace, true, false)
	focus.activate(rl.KeySpace, false, false)
	focus.activate(rl.KeyEnter, false, true)
	if count != 0 {
		t.Fatal("holding Space or releasing another key must not activate the button")
	}
	focus.activate(rl.KeySpace, false, true)
	focus.activate(rl.KeySpace, false, true)
	if count != 1 {
		t.Fatalf("count = %d, want one activation", count)
	}
	focus.activate(rl.KeyEnter, true, false)
	focus.cancelPress()
	focus.activate(rl.KeyEnter, false, true)
	focus.activate(rl.KeySpace, true, false)
	focus.set(Button("Other", func() { count++ }).component)
	focus.activate(rl.KeySpace, false, true)
	focus.set(button.component)
	focus.activate(rl.KeyEnter, true, false)
	button.WithDisabled(true)
	focus.activate(rl.KeyEnter, false, true)
	if count != 1 {
		t.Fatal("cancelled, moved or disabled focus must not activate on release")
	}
}
