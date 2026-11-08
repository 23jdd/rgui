package rgui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestRichTextBoxFormatsSelectionAndCommitsIME(t *testing.T) {
	value := NewSignal([]TextRun{{Text: "a\nb"}})
	box := RichTextBox(value).(*richTextBox)
	box.KeyDown(&InputEvent{Key: rl.KeyA, Ctrl: true})
	box.KeyDown(&InputEvent{Key: rl.KeyB, Ctrl: true})
	if got := value.Get(); len(got) != 1 || !got[0].Style.Bold || got[0].Text != "a\nb" {
		t.Fatalf("format = %+v", got)
	}
	box.CompositionChanged(&InputEvent{Composition: Composition{Text: "ni", Cursor: 2}})
	if glyphs, _, _, _ := box.display(); len(glyphs) != 2 || value.Get()[0].Text != "a\nb" {
		t.Fatal("preedit should not change the document")
	}
	box.TextInput(&InputEvent{Text: "你"})
	if got := value.Get(); len(got) != 1 || got[0].Text != "你" || !got[0].Style.Bold {
		t.Fatalf("commit = %+v", got)
	}
	box.KeyDown(&InputEvent{Key: rl.KeyEnter})
	if got := value.Get()[0].Text; got != "你\n" {
		t.Fatalf("newline = %q", got)
	}
}
