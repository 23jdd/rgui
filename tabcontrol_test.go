package rgui

import "testing"

func TestTabControlSwitchesMountedContent(t *testing.T) {
	first, second := Text(func() string { return "one" }), Text(func() string { return "two" })
	selected := NewSignal(0)
	root := componentOf(TabControl([]TabItem{{"One", first}, {"Two", second}}, selected))
	pending := make(subscriptions)
	root.mount(pending)
	defer root.unmount()
	if componentOf(first).pending == nil || componentOf(second).pending != nil {
		t.Fatal("first tab should be mounted")
	}
	root.children[0].children[1].onClick()
	flushBindings(pending)
	if selected.Get() != 1 || componentOf(first).pending != nil || componentOf(second).pending == nil {
		t.Fatal("second tab should replace first")
	}
}
