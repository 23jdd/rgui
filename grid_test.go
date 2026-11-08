package rgui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestGridToolbarResize(t *testing.T) {
	label := Text(nil)
	componentOf(label).text = "Count: 0"
	spacer := StackPanel(Vertical, 0)
	button := Button("+", func() {})
	root := Grid(
		[]Track{Fixed(48)}, []Track{Auto(), Star(1), Fixed(60)},
		Cell{Child: label}, Cell{Column: 1, Child: spacer}, Cell{Column: 2, Child: button},
	)
	testLayout(root, 300, 48)
	checkBounds(t, label, 0, 0, 80, 48)
	checkBounds(t, spacer, 80, 0, 160, 48)
	checkBounds(t, button, 240, 0, 60, 48)
	componentOf(label).text = "Count: 100"
	testLayout(root, 200, 48)
	checkBounds(t, label, 0, 0, 100, 48)
	checkBounds(t, spacer, 100, 0, 40, 48)
	checkBounds(t, button, 140, 0, 60, 48)
	if componentOf(root).buttonAt(rl.Vector2{X: 175, Y: 24}) != componentOf(button) ||
		componentOf(root).buttonAt(rl.Vector2{X: 250, Y: 24}) != nil {
		t.Fatal("hit testing should follow the new bounds")
	}
	testLayout(root, 120, 48)
	checkBounds(t, spacer, 100, 0, 0, 48)
	checkBounds(t, button, 100, 0, 60, 48)
	if windowButtonAt(componentOf(root), rl.Vector2{X: 125, Y: 24}, rl.Vector2{X: 120, Y: 48}) != nil {
		t.Fatal("a button outside the viewport should not be hit")
	}
}

func TestGridStarWeightsOnBothAxes(t *testing.T) {
	first, second := Text(nil), Text(nil)
	root := Grid(
		[]Track{Star(1), Star(2)}, []Track{Star(1), Star(2)},
		Cell{Child: first}, Cell{Row: 1, Column: 1, Child: second},
	)
	testLayout(root, 300, 150)
	checkBounds(t, first, 0, 0, 100, 50)
	checkBounds(t, second, 100, 50, 200, 100)
}
