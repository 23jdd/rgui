package rgui

import (
	"testing"
	"unicode/utf8"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func testText(value string) rl.Vector2 {
	return rl.Vector2{X: float32(utf8.RuneCountInString(value) * 10), Y: 20}
}

func testLayout(node Node, width, height float32) {
	node.component.measure(rl.Vector2{X: width, Y: height}, testText)
	node.component.arrange(rl.Rectangle{Width: width, Height: height})
}

func checkBounds(t *testing.T, node Node, x, y, width, height float32) {
	t.Helper()
	want := rl.Rectangle{X: x, Y: y, Width: width, Height: height}
	if got := node.component.bounds; got != want {
		t.Fatalf("bounds = %+v, want %+v", got, want)
	}
}

func TestHorizontalStackLayout(t *testing.T) {
	label := Text(nil)
	label.component.text = "Title"
	label = label.WithLayout(Box{VerticalAlignment: Center})
	button := Button("OK", func() {}).WithLayout(Box{Width: 60, Margin: Insets{Left: 4}})
	root := StackPanel(Horizontal, 8, label, button).WithLayout(Box{Padding: UniformInsets(10)})
	testLayout(root, 200, 80)
	checkBounds(t, label, 10, 30, 50, 20)
	checkBounds(t, button, 72, 10, 60, 60)
	if root.component.buttonAt(rl.Vector2{X: 80, Y: 30}) != button.component {
		t.Fatal("hit testing must use the arranged button bounds")
	}
}

func TestLegacyStackLayout(t *testing.T) {
	label, button := Text(nil), Button("OK", func() {})
	root := Stack(label, button)
	layoutRoot(root.component, rl.Vector2{X: 480, Y: 300}, testText)
	checkBounds(t, label, 100, 36, 280, 48)
	checkBounds(t, button, 100, 96, 280, 48)
}
