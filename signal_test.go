package rgui

import (
	"fmt"
	"testing"
)

func TestTextBindingLifecycle(t *testing.T) {
	left, right := NewSignal(0), NewSignal(7)
	useRight := NewSignal(false)
	evaluations, unrelatedEvaluations := 0, 0
	label := Text(func() string {
		evaluations++
		if useRight.Get() {
			return fmt.Sprint(right.Get())
		}
		return fmt.Sprint(left.Get())
	})
	root := Stack(label, Text(func() string {
		unrelatedEvaluations++
		return "static"
	})).component
	pending := make(subscriptions)
	root.mount(pending)
	if label.component.text != "0" || evaluations != 1 {
		t.Fatal("mount did not evaluate the initial label once")
	}

	left.Set(1)
	left.Set(2)
	if left.Get() != 2 || label.component.text != "0" {
		t.Fatal("Set should update the signal immediately and defer the label update")
	}
	flushBindings(pending)
	if label.component.text != "2" || evaluations != 2 || unrelatedEvaluations != 1 {
		t.Fatal("frame should evaluate only the affected binding, once")
	}

	useRight.Set(true)
	flushBindings(pending)
	if label.component.text != "7" {
		t.Fatal("label did not switch dependencies")
	}
	left.Set(3)
	if len(pending) != 0 {
		t.Fatal("previous dependency still schedules the binding")
	}
	right.Set(8)
	root.unmount()
	right.Set(9)
	flushBindings(pending)
	if label.component.text != "7" || evaluations != 3 || len(pending) != 0 {
		t.Fatal("unmount should discard queued updates and unsubscribe")
	}
}

func TestClickRequiresMatchingPressAndRelease(t *testing.T) {
	count := 0
	button := Button("+1", func() { count++ }).component
	var click clickState
	click.update(button, true, false)
	click.update(button, false, false)
	click.update(nil, false, true)
	click.update(nil, true, false)
	click.update(button, false, true)
	if count != 0 {
		t.Fatal("holding or releasing outside the original target should not click")
	}
	click.update(button, true, false)
	click.update(button, false, true)
	if count != 1 {
		t.Fatalf("completed click count = %d, want 1", count)
	}
}
