package rgui

import (
	"testing"
	"unicode/utf16"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestCompositionUsesCodePointOffsets(t *testing.T) {
	text := utf16.Encode([]rune("a😀中"))
	got := decodeComposition(text, 2, []byte{0, 1, 1, 0})
	want := Composition{Text: "a😀中", Cursor: 1, SelectionStart: 1, SelectionLength: 1}
	if got != want {
		t.Fatalf("composition = %+v, want %+v", got, want)
	}
}

func TestIMECaretUsesContentOriginAndDPI(t *testing.T) {
	caret := rl.Rectangle{X: 5.25, Y: 6, Width: 0, Height: 16}
	content := rl.Rectangle{X: 10, Y: 20}
	got := imeCaretPixels(caret, content, rl.Vector2{X: 1.25, Y: 1.5})
	want := imePixelRect{left: 19, top: 39, right: 20, bottom: 63}
	if got != want {
		t.Fatalf("caret = %+v, want %+v", got, want)
	}
}
