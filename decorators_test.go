package rgui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestCanvasAnchorsAndNestedClip(t *testing.T) {
	button := WithLayout(Button("anchor", func() {}), Box{Width: 30, Height: 20})
	stretch := Text(nil)
	inner := WithLayout(Canvas(
		CanvasItem{Child: button, Right: Anchor(-10), Bottom: Anchor(5)},
		CanvasItem{Child: stretch, Left: Anchor(5), Right: Anchor(15), Top: Anchor(2), Bottom: Anchor(40)},
	), Box{ClipToBounds: true})
	root := Border(inner, BorderStyle{Thickness: 2, Padding: UniformInsets(3), ClipToBounds: true})
	testLayout(root, 100, 80)
	checkBounds(t, button, 75, 50, 30, 20)
	checkBounds(t, stretch, 10, 7, 70, 28)
	if componentOf(root).buttonAt(rl.Vector2{X: 97, Y: 60}) != nil ||
		componentOf(root).buttonAt(rl.Vector2{X: 90, Y: 60}) != componentOf(button) {
		t.Fatal("overflow should be hittable only inside the shared clip")
	}
}

func TestGridSpansFixedAutoAndStar(t *testing.T) {
	label := Text(nil)
	componentOf(label).text = "123456789"
	auto := Text(nil)
	span := Text(nil)
	root := Grid(
		[]Track{Fixed(20), Star(1)}, []Track{Fixed(30), Auto(), Star(1)},
		Cell{Child: label, ColumnSpan: 2},
		Cell{Child: auto, Column: 1, Row: 1},
		Cell{Child: span, Column: 2, RowSpan: 2},
	)
	testLayout(root, 160, 80)
	checkBounds(t, label, 0, 0, 90, 20)
	checkBounds(t, auto, 30, 20, 60, 60)
	checkBounds(t, span, 90, 0, 70, 80)
}
