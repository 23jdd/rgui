package rgui

import (
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type EventKind uint8

const (
	PointerDown EventKind = iota
	PointerMove
	PointerUp
	PointerWheel
	PointerCaptureLost
	KeyDown
	KeyUp
	FocusGained
	FocusLost
)

type EventPhase uint8

const (
	Preview EventPhase = iota
	Bubble
)

// InputOptions describes how a node participates in input.
// Negative TabIndex excludes a node from keyboard navigation.
type InputOptions struct {
	Disabled, HitTestInvisible, Focusable bool
	TabIndex                              int
	OnEvent                               func(*InputEvent)
}

// WithInput configures a node's input behavior. Buttons are focusable by default.
func WithInput(node Node, options InputOptions) Node {
	c := componentOf(node)
	c.layoutBase().Input = options
	return c
}

// InputEvent is valid for the duration of its handler.
type InputEvent struct {
	Kind                    EventKind
	Phase                   EventPhase
	Source, CurrentTarget   Node
	Position, LocalPosition rl.Vector2
	Wheel                   rl.Vector2
	Key                     int32
	Shift, Ctrl, Alt        bool
	Handled                 bool

	router  *inputRouter
	current *component
}

// CapturePointer routes subsequent mouse movement and release to CurrentTarget.
func (e *InputEvent) CapturePointer() bool {
	if e.router == nil || !e.router.down || e.current == nil || e.current.pending == nil || !e.current.inputEnabled() ||
		(e.Kind != PointerDown && e.Kind != PointerMove) {
		return false
	}
	e.router.setCapture(e.current)
	return true
}

// ReleasePointer releases capture if CurrentTarget owns it.
func (e *InputEvent) ReleasePointer() {
	if e.router != nil && e.router.capture == e.current {
		e.router.setCapture(nil)
	}
}

func (c *component) eventNode() Node {
	if c.custom != nil {
		return c.custom
	}
	return c
}

func (c *component) inputEnabled() bool {
	if c == nil {
		return false
	}
	for node := c; node != nil; node = node.parent {
		options := node.layoutBase().Input
		if options.Disabled || options.HitTestInvisible {
			return false
		}
	}
	return true
}

func hitNode(c *component, point rl.Vector2) *component {
	options := c.layoutBase().Input
	if options.HitTestInvisible {
		return nil
	}
	if !options.Disabled {
		for i := len(c.children) - 1; i >= 0; i-- {
			if hit := hitNode(c.children[i], point); hit != nil {
				return hit
			}
		}
	}
	if c.bounds.Width > 0 && c.bounds.Height > 0 && rl.CheckCollisionPointRec(point, c.bounds) {
		return c
	}
	return nil
}

func routeEvent(target *component, event *InputEvent) {
	if target == nil {
		return
	}
	event.Source = target.eventNode()
	var path []*component
	for node := target; node != nil; node = node.parent {
		path = append(path, node)
	}
	for _, phase := range [...]EventPhase{Preview, Bubble} {
		slices.Reverse(path)
		event.Phase = phase
		for _, node := range path {
			if event.Handled {
				break
			}
			event.current, event.CurrentTarget = node, node.eventNode()
			event.LocalPosition = rl.Vector2{X: event.Position.X - node.bounds.X, Y: event.Position.Y - node.bounds.Y}
			if handler := node.layoutBase().Input.OnEvent; handler != nil {
				handler(event)
			}
		}
	}
	event.current, event.CurrentTarget = nil, nil
}

type inputRouter struct {
	capture *component
	point   rl.Vector2
	down    bool
}

func (r *inputRouter) setCapture(target *component) {
	old := r.capture
	r.capture = target
	if old != nil && old != target {
		routeEvent(old, &InputEvent{Kind: PointerCaptureLost, Position: r.point})
	}
}

func (r *inputRouter) cancel() {
	r.down = false
	r.setCapture(nil)
}

func (r *inputRouter) prune() {
	if r.capture != nil && (r.capture.pending == nil || !r.capture.inputEnabled()) {
		r.setCapture(nil)
	}
}

func (r *inputRouter) pointer(kind EventKind, hit *component, point, wheel rl.Vector2) bool {
	r.point = point
	if kind == PointerDown {
		r.down = true
	}
	target := hit
	if r.capture != nil && (kind == PointerMove || kind == PointerUp) {
		target = r.capture
	}
	event := &InputEvent{Kind: kind, Position: point, Wheel: wheel, router: r}
	if target.inputEnabled() {
		routeEvent(target, event)
	}
	if kind == PointerUp {
		r.cancel()
	}
	return event.Handled
}

func hitWindow(root *component, mouse, viewport rl.Vector2) *component {
	if mouse.X < 0 || mouse.Y < 0 || mouse.X >= viewport.X || mouse.Y >= viewport.Y {
		return nil
	}
	return hitNode(root, mouse)
}

func nearestFocusable(hit *component, viewport rl.Vector2) *component {
	for node := hit; node != nil; node = node.parent {
		if node.focusable(viewport) {
			return node
		}
	}
	return nil
}

func nearestButton(hit *component) *component {
	if hit.inputEnabled() {
		for node := hit; node != nil; node = node.parent {
			if node.enabledButton() {
				return node
			}
		}
	}
	return nil
}
