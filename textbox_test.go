package rgui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestTextBoxCompositionReplacesSelectionOnlyOnCommit(t *testing.T) {
	value := NewSignal("a😀b")
	box := TextBox(value).(*textBox)
	box.anchor, box.caret = 1, 2
	box.last = value.Get()
	box.CompositionChanged(&InputEvent{Composition: Composition{Text: "ni", Cursor: 2}})
	line, _, _, _ := box.display(box.read())
	if string(line) != "anib" || value.Get() != "a😀b" {
		t.Fatal("preedit should only change the displayed text")
	}
	box.CompositionChanged(&InputEvent{})
	if value.Get() != "a😀b" {
		t.Fatal("cancel changed committed text")
	}
	box.TextInput(&InputEvent{Text: "你\n"})
	if value.Get() != "a你b" || box.caret != 2 || box.anchor != 2 {
		t.Fatalf("value=%q caret=%d anchor=%d", value.Get(), box.caret, box.anchor)
	}
}

func TestTextBoxEditsAndTracksExternalChanges(t *testing.T) {
	value := NewSignal("abcd")
	box := TextBox(value).(*textBox)
	node := componentOf(box)
	pending := make(subscriptions)
	node.mount(pending)
	defer node.unmount()
	testLayout(node, 100, 42)
	box.KeyDown(&InputEvent{Key: rl.KeyEnd})
	box.KeyDown(&InputEvent{Key: rl.KeyLeft, Shift: true})
	box.KeyDown(&InputEvent{Key: rl.KeyDelete})
	if value.Get() != "abc" {
		t.Fatalf("selection delete = %q", value.Get())
	}
	value.Set("x")
	flushBindings(pending)
	if box.caret != 1 || box.anchor != 1 || !nonempty(node.visualDamage) {
		t.Fatal("external text changes should clamp selection and invalidate the visual")
	}
}
