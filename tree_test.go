package rgui

import "testing"

func TestReplaceChildrenKeepsMountedSiblings(t *testing.T) {
	signal := NewSignal("initial")
	evaluations := 0
	label := Text(func() string { evaluations++; return signal.Get() })
	button := Button("click", func() {})
	root := componentOf(WithChildren(&BaseNode{}, label, button))
	pending := make(subscriptions)
	root.mount(pending)
	defer root.unmount()
	var click clickState
	var focus focusState
	click.pressed, focus.button = componentOf(button), componentOf(button)
	if err := SetChildren(root, label); err != nil {
		t.Fatal(err)
	}
	pruneInput(&click, &focus)
	if evaluations != 1 || click.pressed != nil || focus.button != nil {
		t.Fatal("removal should preserve the sibling and clear detached input")
	}
	signal.Set("changed")
	if err := SetChildren(root); err != nil {
		t.Fatal(err)
	}
	flushBindings(pending)
	if evaluations != 1 {
		t.Fatal("removed binding should not evaluate")
	}
}

func TestSetChildrenRejectsInvalidTreeAtomically(t *testing.T) {
	leaf := &BaseNode{}
	root := componentOf(WithChildren(&BaseNode{}, leaf))
	other := WithChildren(&BaseNode{})
	if SetChildren(root, leaf, leaf) == nil || SetChildren(root, root) == nil || SetChildren(other, leaf) == nil {
		t.Fatal("accepted duplicate, cycle or second parent")
	}
	if len(root.children) != 1 || root.children[0] != componentOf(leaf) {
		t.Fatal("invalid mutations must preserve the previous tree")
	}
}
