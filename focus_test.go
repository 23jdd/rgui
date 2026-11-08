package rgui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestFocusTraversal(t *testing.T) {
	first := Button("First", func() {})
	disabled := WithDisabled(Button("Disabled", func() {}), true)
	last := Button("Last", func() {})
	outside := Button("Outside", func() {})
	root := StackPanel(Vertical, 0, first, Stack(disabled, last), outside)
	viewport := rl.Vector2{X: 200, Y: 144}
	testLayout(root, viewport.X, viewport.Y)
	var focus focusState
	for _, step := range []struct {
		backward bool
		want     Node
	}{
		{false, first},
		{false, last},
		{false, first},
		{true, last},
		{true, first},
	} {
		focus.move(componentOf(root), viewport, step.backward)
		if focus.button != componentOf(step.want) {
			t.Fatalf("unexpected focus, want %s", componentOf(step.want).text)
		}
	}
	if componentOf(root).buttonAt(rl.Vector2{X: 20, Y: 60}) != nil {
		t.Fatal("disabled button should not accept mouse input")
	}
}

func TestKeyboardActivationAndCancellation(t *testing.T) {
	count := 0
	button := Button("Count", func() { count++ })
	var focus focusState
	focus.set(componentOf(button))
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
	focus.set(componentOf(Button("Other", func() { count++ })))
	focus.activate(rl.KeySpace, false, true)
	focus.set(componentOf(button))
	focus.activate(rl.KeyEnter, true, false)
	WithDisabled(button, true)
	focus.activate(rl.KeyEnter, false, true)
	if count != 1 {
		t.Fatal("cancelled, moved or disabled focus must not activate on release")
	}
}
