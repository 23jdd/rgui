package rgui

import "testing"

func TestTreeViewExpandsAndSelectsNestedNode(t *testing.T) {
	selected := NewSignal("")
	root := componentOf(TreeView([]TreeItem{{Key: "p", Label: "Parent", Children: []TreeItem{{Key: "c", Label: "Child"}}}}, selected))
	pending := make(subscriptions)
	root.mount(pending)
	defer root.unmount()
	branch := root.children[0].children[0]
	branch.children[0].children[0].onClick()
	flushBindings(pending)
	child := branch.children[1].children[0].children[0]
	child.onClick()
	if selected.Get() != "c" {
		t.Fatal("expected nested selection")
	}
	branch.children[0].children[0].onClick()
	flushBindings(pending)
	if child.pending != nil || selected.Get() != "c" {
		t.Fatal("collapse should detach child and keep selection")
	}
}
