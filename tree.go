package rgui

import (
	"fmt"
	"reflect"
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// WithChildren gives a custom node children. Children occupy its content area
// and paint in declaration order over the parent. Retain the returned Node for
// subsequent SetChildren calls.
func WithChildren(node Node, children ...Node) Node {
	c := componentOf(node)
	if c == nil || c.custom == nil {
		panic("rgui: WithChildren requires a custom node")
	}
	if err := c.replaceChildren(children); err != nil {
		panic(err)
	}
	c.panel, c.mutable = overlayLayout{}, true
	return c
}

// SetChildren replaces a custom container's children on the UI thread.
// Existing children that appear in the new list stay mounted.
func SetChildren(parent Node, children ...Node) error {
	c, ok := parent.(*component)
	if !ok || c == nil || !c.mutable {
		return fmt.Errorf("rgui: SetChildren requires the result of WithChildren")
	}
	return c.replaceChildren(children)
}

func (c *component) replaceChildren(nodes []Node) error {
	next := make([]*component, 0, len(nodes))
	for _, node := range nodes {
		child := componentOf(node)
		if child == nil {
			return fmt.Errorf("rgui: nil child")
		}
		if child.custom != nil {
			for _, previous := range c.children {
				if sameCustom(previous.custom, child.custom) {
					child = previous
					break
				}
			}
		}
		if child.parent != nil && child.parent != c || child.pending != nil && child.parent != c {
			return fmt.Errorf("rgui: child already belongs to a tree")
		}
		for ancestor := c; ancestor != nil; ancestor = ancestor.parent {
			if ancestor == child || sameCustom(ancestor.custom, child.custom) {
				return fmt.Errorf("rgui: node tree cannot contain a cycle")
			}
		}
		for _, previous := range next {
			if previous == child || sameCustom(previous.custom, child.custom) {
				return fmt.Errorf("rgui: duplicate child")
			}
		}
		next = append(next, child)
	}
	root := c
	for root.parent != nil {
		root = root.parent
	}
	seen := make(map[Node]bool)
	if !uniqueCustomPointers(root, c, next, seen) {
		return fmt.Errorf("rgui: custom node already appears in the tree")
	}
	if slices.Equal(c.children, next) {
		return nil
	}
	for _, child := range c.children {
		if !slices.Contains(next, child) {
			if child.pending != nil {
				child.unmount()
			}
			child.parent = nil
		}
	}
	old := c.children
	c.children = next
	c.treeDirty = true
	for _, child := range next {
		child.parent = c
		if c.pending != nil && !slices.Contains(old, child) {
			child.mount(c.pending)
		}
	}
	return nil
}

func uniqueCustomPointers(node, replaced *component, next []*component, seen map[Node]bool) bool {
	if node.custom != nil && reflect.ValueOf(node.custom).Kind() == reflect.Pointer {
		if seen[node.custom] {
			return false
		}
		seen[node.custom] = true
	}
	children := node.children
	if node == replaced {
		children = next
	}
	for _, child := range children {
		if !uniqueCustomPointers(child, replaced, next, seen) {
			return false
		}
	}
	return true
}

func sameCustom(a, b Node) bool {
	if a == nil || b == nil {
		return false
	}
	v := reflect.ValueOf(a)
	return v.Kind() == reflect.Pointer && v.Type() == reflect.TypeOf(b) && a == b
}

type overlayLayout struct{}

func (overlayLayout) measure(children []*component, available rl.Vector2, text measureText) rl.Vector2 {
	var desired rl.Vector2
	for _, child := range children {
		size := child.measure(available, text)
		desired.X, desired.Y = max(desired.X, size.X), max(desired.Y, size.Y)
	}
	return desired
}

func (overlayLayout) arrange(children []*component, area rl.Rectangle) {
	for _, child := range children {
		child.arrange(area)
	}
}

func pruneInput(click *clickState, focus *focusState) {
	if click.pressed != nil && click.pressed.pending == nil {
		click.pressed = nil
	}
	if focus.button != nil && focus.button.pending == nil {
		focus.set(nil)
	}
}
