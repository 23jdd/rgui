package rgui

import (
	"slices"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type eventNode struct {
	BaseNode
	name    string
	order   *[]string
	capture bool
	block   bool
}

func (n *eventNode) PreviewInput(e *InputEvent) {
	if e.Kind == PointerDown {
		*n.order = append(*n.order, n.name+" preview")
		e.Handled = n.block
	}
}

func (n *eventNode) PointerDown(e *InputEvent) {
	*n.order = append(*n.order, n.name+" bubble")
	if n.capture {
		e.CapturePointer()
	}
}

type captureNode struct{ eventNode }

func (n *captureNode) PointerUp(*InputEvent) {
	*n.order = append(*n.order, "captured release")
}

func (n *captureNode) PointerCaptureLost(*InputEvent) {
	*n.order = append(*n.order, "capture lost")
}

func TestEventRoutingAndCapture(t *testing.T) {
	var order []string
	child := &captureNode{eventNode: eventNode{name: "child", order: &order, capture: true}}
	child.Box = Box{Width: 40, Height: 40}
	base := &eventNode{name: "root", order: &order}
	root := componentOf(WithChildren(base, child))
	root.mount(make(subscriptions))
	defer root.unmount()
	testLayout(root, 100, 100)
	var router inputRouter
	router.pointer(PointerDown, componentOf(child), rl.Vector2{X: 35, Y: 35}, rl.Vector2{})
	if router.capture != componentOf(child) {
		t.Fatal("pointer handler did not capture")
	}
	router.pointer(PointerUp, nil, rl.Vector2{X: 150}, rl.Vector2{})
	want := []string{"root preview", "child preview", "child bubble", "root bubble", "captured release", "capture lost"}
	if !slices.Equal(order, want) || router.capture != nil {
		t.Fatalf("route = %v, capture = %v", order, router.capture)
	}
}

type keyboardNode struct {
	eventNode
	keys   []int32
	gained int
	lost   int
}

func (n *keyboardNode) KeyDown(e *InputEvent)   { n.keys = append(n.keys, e.Key) }
func (n *keyboardNode) FocusGained(*InputEvent) { n.gained++ }
func (n *keyboardNode) FocusLost(*InputEvent)   { n.lost++ }

func TestPreviewCancellationAndInterfaceFocus(t *testing.T) {
	var order []string
	target := &keyboardNode{eventNode: eventNode{name: "target", order: &order}}
	target.Box = Box{Height: 40}
	target.Input.TabIndex = 1
	button := WithInput(Button("button", func() {}), InputOptions{TabIndex: 2})
	base := &eventNode{name: "root", order: &order, block: true}
	root := componentOf(WithChildren(base, Stack(target, button)))
	event := &InputEvent{Kind: PointerDown}
	routeEvent(componentOf(target), event)
	if !event.Handled || !slices.Equal(order, []string{"root preview"}) {
		t.Fatalf("preview did not cancel routing: %v", order)
	}
	testLayout(root, 200, 200)
	var focus focusState
	viewport := rl.Vector2{X: 200, Y: 200}
	focus.move(root, viewport, false)
	if focus.button != componentOf(target) || target.gained != 1 {
		t.Fatal("Focusable should opt in without a configuration flag")
	}
	routeEvent(focus.button, &InputEvent{Kind: KeyDown, Key: rl.KeyRight})
	if !slices.Equal(target.keys, []int32{rl.KeyRight}) {
		t.Fatalf("keys = %v", target.keys)
	}
	focus.move(root, viewport, false)
	if focus.button != componentOf(button) || target.lost != 1 {
		t.Fatal("TabIndex or FocusLost did not work")
	}
}
