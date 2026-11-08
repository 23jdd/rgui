package rgui

import "testing"

func TestExpanderMountsContentOnceAndCleansBindings(t *testing.T) {
	expanded := NewSignal(true)
	value := NewSignal("content")
	evaluations := 0
	content := Text(func() string { evaluations++; return value.Get() })
	root := componentOf(Expander("Details", content, expanded))
	pending := make(subscriptions)
	root.mount(pending)
	defer root.unmount()
	if evaluations != 1 {
		t.Fatalf("evaluations = %d", evaluations)
	}
	expanded.Set(false)
	flushBindings(pending)
	value.Set("updated")
	flushBindings(pending)
	if evaluations != 1 || componentOf(content).pending != nil {
		t.Fatal("collapsed content was not detached")
	}
	expanded.Set(true)
	flushBindings(pending)
	if evaluations != 2 || componentOf(content).text != "updated" {
		t.Fatal("expanded content did not remount")
	}
}
