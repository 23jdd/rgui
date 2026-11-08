package rgui

import (
	"slices"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestEventRoutingAndCapture(t *testing.T) {
	var order []string
	child := &BaseNode{Box: Box{Width: 40, Height: 40}}
	child.Input.OnEvent = func(e *InputEvent) {
		switch {
		case e.Kind == PointerDown && e.Phase == Preview:
			order = append(order, "child preview")
		case e.Kind == PointerDown:
			order = append(order, "child bubble")
			if !e.CapturePointer() {
				t.Fatal("capture failed")
			}
		case e.Kind == PointerUp && e.Phase == Bubble:
			order = append(order, "captured release")
		case e.Kind == PointerCaptureLost && e.Phase == Bubble:
			order = append(order, "capture lost")
		}
	}
	base := &BaseNode{Input: InputOptions{OnEvent: func(e *InputEvent) {
		if e.Kind != PointerDown {
			return
		}
		if e.Phase == Preview {
			order = append(order, "root preview")
		} else {
			order = append(order, "root bubble")
		}
	}}}
	root := componentOf(WithChildren(base, child))
	root.mount(make(subscriptions))
	defer root.unmount()
	testLayout(root, 100, 100)
	var router inputRouter
	router.pointer(PointerDown, componentOf(child), rl.Vector2{X: 35, Y: 35}, rl.Vector2{})
	router.pointer(PointerUp, nil, rl.Vector2{X: 150}, rl.Vector2{})
	want := []string{"root preview", "child preview", "child bubble", "root bubble", "captured release", "capture lost"}
	if !slices.Equal(order, want) || router.capture != nil {
		t.Fatalf("route = %v, capture = %v", order, router.capture)
	}
}

func TestEventHandledAndTabPriority(t *testing.T) {
	targetCalls := 0
	first := WithInput(Button("first", func() {}), InputOptions{TabIndex: 2, OnEvent: func(*InputEvent) { targetCalls++ }})
	second := WithInput(Button("second", func() {}), InputOptions{TabIndex: 1})
	root := componentOf(WithInput(Stack(first, second), InputOptions{OnEvent: func(e *InputEvent) {
		if e.Kind == PointerDown {
			e.Handled = true
		}
	}}))
	event := &InputEvent{Kind: PointerDown}
	routeEvent(componentOf(first), event)
	if !event.Handled || targetCalls != 0 {
		t.Fatal("preview did not cancel routing")
	}
	testLayout(root, 200, 200)
	var focus focusState
	focus.move(root, rl.Vector2{X: 200, Y: 200}, false)
	if focus.button != componentOf(second) {
		t.Fatal("TabIndex did not set traversal priority")
	}
}
