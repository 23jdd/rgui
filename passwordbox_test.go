package rgui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"testing"
)

func TestPasswordBoxMasksAndDoesNotCut(t *testing.T) {
	value := NewSignal("a😀中")
	box := PasswordBox(value).(*textBox)
	box.KeyDown(&InputEvent{Key: rl.KeyA, Ctrl: true})
	box.KeyDown(&InputEvent{Key: rl.KeyX, Ctrl: true})
	box.CompositionChanged(&InputEvent{Composition: Composition{Text: "secret"}})
	line, _, _, _ := box.display(box.read())
	if value.Get() != "a😀中" || string(line) != "***" || !box.suppressIME() {
		t.Fatalf("password masking or cut protection failed: %q", line)
	}
}
