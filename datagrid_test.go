package rgui

import "testing"

func TestDataGridEditsOnlyTargetCell(t *testing.T) {
	first, second := NewSignal("one"), NewSignal("two")
	root := componentOf(DataGrid([]GridViewColumn{{Header: "A"}, {Header: "B"}}, [][]*Signal[string]{{first, second}}))
	node := root.children[0].children[1].children[1]
	routeEvent(node, &InputEvent{Kind: TextInput, Text: "new "})
	if first.Get() != "one" || second.Get() != "new two" {
		t.Fatal("wrong cell was edited")
	}
}
