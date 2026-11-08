package rgui

import "testing"

func TestDockPanelConsumesEdgesAndFills(t *testing.T) {
	top := WithLayout(Text(nil), Box{Height: 30})
	left := WithLayout(Text(nil), Box{Width: 50})
	bottom := WithLayout(Text(nil), Box{Height: 15})
	right := WithLayout(Text(nil), Box{Width: 20})
	center := Text(nil)
	root := DockPanel(true,
		Docked{Edge: DockTop, Child: top}, Docked{Edge: DockLeft, Child: left},
		Docked{Edge: DockBottom, Child: bottom}, Docked{Edge: DockRight, Child: right},
		Docked{Edge: DockTop, Child: center},
	)
	testLayout(root, 250, 150)
	checkBounds(t, top, 0, 0, 250, 30)
	checkBounds(t, left, 0, 30, 50, 120)
	checkBounds(t, bottom, 50, 135, 200, 15)
	checkBounds(t, right, 230, 30, 20, 105)
	checkBounds(t, center, 50, 30, 180, 105)
}

func TestWrapPanelReflowsInBothDirections(t *testing.T) {
	for _, direction := range []Direction{Horizontal, Vertical} {
		first := WithLayout(Text(nil), Box{Width: 40, Height: 40})
		second := WithLayout(Text(nil), Box{Width: 40, Height: 40})
		third := WithLayout(Text(nil), Box{Width: 40, Height: 40})
		root := WrapPanel(direction, 5, first, second, third)
		testLayout(root, 85, 85)
		secondAt, thirdAt := direction.size(45, 0), direction.size(0, 45)
		checkBounds(t, second, secondAt.X, secondAt.Y, 40, 40)
		checkBounds(t, third, thirdAt.X, thirdAt.Y, 40, 40)
		available := direction.size(130, 85)
		testLayout(root, available.X, available.Y)
		thirdAt = direction.size(90, 0)
		checkBounds(t, third, thirdAt.X, thirdAt.Y, 40, 40)
	}
}
