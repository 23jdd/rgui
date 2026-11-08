package rgui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"testing"
)

func TestCheckBoxPointerAndKeyboardToggle(t *testing.T) {
	value := NewSignal(false)
	node := CheckBox("Enable", value).(*choice)
	node.Tapped(&InputEvent{Source: node})
	node.KeyDown(&InputEvent{Source: node, Key: rl.KeySpace})
	if !value.Get() {
		t.Fatal("space should activate on release")
	}
	node.KeyUp(&InputEvent{Source: node, Key: rl.KeySpace})
	node.KeyUp(&InputEvent{Source: node, Key: rl.KeySpace})
	if value.Get() {
		t.Fatal("expected one toggle")
	}
}
