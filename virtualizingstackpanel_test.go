package rgui

import "testing"

func TestVirtualizingStackPanelReusesVisibleRows(t *testing.T) {
	offset := NewSignal(float32(0))
	created := make(map[int]Node)
	root := componentOf(VirtualizingStackPanel(1000, 20, offset, func(index int) Node {
		node := Text(nil)
		created[index] = node
		return node
	}))
	pending := make(subscriptions)
	root.mount(pending)
	defer root.unmount()
	testLayout(root, 200, 60)
	flushBindings(pending)
	testLayout(root, 200, 60)
	if len(root.children) != 4 || len(created) > 11 {
		t.Fatal("must mount only visible rows and overscan")
	}
	two := componentOf(created[2])
	offset.Set(40)
	flushBindings(pending)
	testLayout(root, 200, 60)
	if root.children[1] != two || two.pending == nil || componentOf(created[0]).pending != nil {
		t.Fatal("overlapping rows should retain identity and offscreen rows must detach")
	}
}
