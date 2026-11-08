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
	node.Measure(MeasureContext{Available: rl.Vector2{X: width, Y: height}, MeasureText: testText})
	node.Place(rl.Rectangle{Width: width, Height: height})
}

func checkBounds(t *testing.T, node Node, x, y, width, height float32) {
	t.Helper()
	want := rl.Rectangle{X: x, Y: y, Width: width, Height: height}
	if got := componentOf(node).bounds; got != want {
		t.Fatalf("bounds = %+v, want %+v", got, want)
	}
}

func TestHorizontalStackLayout(t *testing.T) {
	label := Text(nil)
	componentOf(label).text = "Title"
	label = WithLayout(label, Box{VerticalAlignment: Center})
	button := WithLayout(Button("OK", func() {}), Box{Width: 60, Margin: Insets{Left: 4}})
	root := WithLayout(StackPanel(Horizontal, 8, label, button), Box{Padding: UniformInsets(10)})
	testLayout(root, 200, 80)
	checkBounds(t, label, 10, 30, 50, 20)
	checkBounds(t, button, 72, 10, 60, 60)
	if componentOf(root).buttonAt(rl.Vector2{X: 80, Y: 30}) != componentOf(button) {
		t.Fatal("hit testing must use the arranged button bounds")
	}
}

func TestLegacyStackLayout(t *testing.T) {
	label, button := Text(nil), Button("OK", func() {})
	root := Stack(label, button)
	layoutRoot(componentOf(root), rl.Vector2{X: 480, Y: 300}, testText)
	checkBounds(t, label, 100, 36, 280, 48)
	checkBounds(t, button, 100, 96, 280, 48)
}

func TestBoxLimitsAndAlignment(t *testing.T) {
	node := WithLayout(Text(nil), Box{
		Width: 30, MinWidth: 80, MaxWidth: 60, MaxHeight: 30,
		Margin:              Insets{Left: 10, Right: 10, Top: 5, Bottom: 5},
		HorizontalAlignment: End,
	})
	testLayout(node, 200, 100)
	checkBounds(t, node, 110, 35, 80, 30)
	if got := componentOf(node).DesiredSize(); got != (rl.Vector2{X: 100, Y: 30}) {
		t.Fatalf("desired size = %+v", got)
	}

	testLayout(node, 60, 20)
	checkBounds(t, node, -30, 5, 80, 10)
}

func TestEmbeddedBaseNodeBoxAndLayoutChanges(t *testing.T) {
	base := &BaseNode{Box: Box{
		MinWidth: 40, MinHeight: 30, MaxWidth: 70,
		Margin: Insets{Left: 5, Right: 5}, Padding: UniformInsets(2),
		HorizontalAlignment: End,
	}}
	root := StackPanel(Vertical, 0, base)
	testLayout(root, 100, 100)
	if got := base.DesiredSize(); got != (rl.Vector2{X: 50, Y: 30}) {
		t.Fatalf("desired size = %+v", got)
	}
	if got := base.LayoutSlot(); got != (rl.Rectangle{Width: 100, Height: 30}) {
		t.Fatalf("layout slot = %+v", got)
	}
	if got := base.Bounds(); got != (rl.Rectangle{X: 57, Y: 2, Width: 36, Height: 26}) {
		t.Fatalf("content bounds = %+v", got)
	}
	base.HorizontalAlignment = Stretch
	base.MinHeight = 45
	if !componentOf(root).layoutChanged() {
		t.Fatal("editing an embedded child's box should invalidate layout")
	}
	testLayout(root, 100, 100)
	if componentOf(root).layoutChanged() {
		t.Fatal("measurement should acknowledge the new box")
	}
	if got := base.Bounds(); got != (rl.Rectangle{X: 17, Y: 2, Width: 66, Height: 41}) {
		t.Fatalf("updated content bounds = %+v", got)
	}
}
