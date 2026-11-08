package rgui

import "testing"

func TestFrameBackForwardAndBranch(t *testing.T) {
	a, b, c := Page("A", Text(nil)), Page("B", Text(nil)), Page("C", Text(nil))
	nav := NewNavigator(a)
	root := componentOf(Frame(nav))
	pending := make(subscriptions)
	root.mount(pending)
	defer root.unmount()
	nav.Navigate(b)
	flushBindings(pending)
	if nav.Title() != "B" || componentOf(a).pending != nil || componentOf(b).pending == nil {
		t.Fatal("navigation did not change pages")
	}
	if !nav.Back() || !nav.Forward() || !nav.Back() {
		t.Fatal("expected back and forward history")
	}
	nav.Navigate(c)
	flushBindings(pending)
	if nav.Title() != "C" || nav.Forward() || componentOf(c).pending == nil {
		t.Fatal("a new navigation must replace forward history")
	}
}
