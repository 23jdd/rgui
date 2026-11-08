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

type interactionNode struct {
	BaseNode
	hover    []EventKind
	taps     []PointerButton
	text     string
	repeated []int32
}

func (n *interactionNode) PointerEnter(e *InputEvent)    { n.hover = append(n.hover, e.Kind) }
func (n *interactionNode) PointerLeave(e *InputEvent)    { n.hover = append(n.hover, e.Kind) }
func (n *interactionNode) Tapped(e *InputEvent)          { n.taps = append(n.taps, e.Button) }
func (n *interactionNode) SecondaryTapped(e *InputEvent) { n.taps = append(n.taps, e.Button) }
func (n *interactionNode) TextInput(e *InputEvent)       { n.text += e.Text }
func (n *interactionNode) KeyRepeat(e *InputEvent)       { n.repeated = append(n.repeated, e.Key) }

func TestHoverAndClickInterfaces(t *testing.T) {
	parent, left, right := &interactionNode{}, &interactionNode{}, &interactionNode{}
	root := componentOf(WithChildren(parent, StackPanel(Horizontal, 0, left, right)))
	root.mount(make(subscriptions))
	defer root.unmount()
	var router inputRouter
	a, b := componentOf(left), componentOf(right)
	point := rl.Vector2{X: 10, Y: 10}
	router.updateHover(a, point)
	router.updateHover(b, point)
	router.updateHover(b, point)
	router.updateHover(nil, point)
	for _, node := range []*interactionNode{parent, left, right} {
		if !slices.Equal(node.hover, []EventKind{PointerEnter, PointerLeave}) {
			t.Fatalf("hover events = %v", node.hover)
		}
	}
	router.pointer(PointerDown, a, point, rl.Vector2{})
	router.secondary(PointerDown, a, point)
	router.secondary(PointerUp, a, point)
	router.pointer(PointerUp, a, point, rl.Vector2{})
	router.pointer(PointerDown, a, point, rl.Vector2{})
	router.pointer(PointerUp, b, point, rl.Vector2{})
	router.secondary(PointerDown, a, point)
	router.secondary(PointerUp, b, point)
	if !slices.Equal(left.taps, []PointerButton{SecondaryButton, PrimaryButton}) || len(right.taps) != 0 {
		t.Fatalf("left taps = %v, right taps = %v", left.taps, right.taps)
	}
}

func TestTextAndRepeatDispatchSeparately(t *testing.T) {
	node := &interactionNode{}
	target := componentOf(node)
	routeEvent(target, &InputEvent{Kind: KeyDown, Key: rl.KeyA})
	routeEvent(target, &InputEvent{Kind: TextInput, Text: "中"})
	routeEvent(target, &InputEvent{Kind: KeyRepeat, Key: rl.KeyBackspace})
	if node.text != "中" || !slices.Equal(node.repeated, []int32{rl.KeyBackspace}) {
		t.Fatalf("text = %q, repeated keys = %v", node.text, node.repeated)
	}
}
