package rgui

import "testing"

func TestListViewRoutesCellClickToRow(t *testing.T) {
	selected := NewSignal(-1)
	root := componentOf(ListView([]GridViewColumn{{Header: "One"}, {Header: "Two"}},
		[][]string{{"a", "b"}, {"c", "d"}}, selected))
	row := root.children[0].children[2]
	routeEvent(row.children[0].children[1], &InputEvent{Kind: Tap})
	if selected.Get() != 1 {
		t.Fatal("tap on cell should select its row")
	}
}
