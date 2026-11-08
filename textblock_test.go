package rgui

import "testing"

func TestTextBlockReflowsAndReacts(t *testing.T) {
	value := NewSignal("ab cd")
	node := TextBlock(value.Get, Start).(*textBlock)
	root := componentOf(node)
	pending := make(subscriptions)
	root.mount(pending)
	defer root.unmount()
	testLayout(root, 30, 100)
	if len(node.lines) != 2 || node.DesiredSize().Y != 2*fontSize {
		t.Fatalf("narrow text lines: %v", node.lines)
	}
	testLayout(root, 60, 100)
	if len(node.lines) != 1 {
		t.Fatalf("wide text lines: %v", node.lines)
	}
	value.Set("ab\ncd")
	flushBindings(pending)
	testLayout(root, 60, 100)
	if len(node.lines) != 2 {
		t.Fatalf("updated text lines: %v", node.lines)
	}
}
