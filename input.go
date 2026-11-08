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
	PointerEnter
	PointerLeave
	Tap
	SecondaryTap
	KeyRepeat
	TextInput
	CompositionChanged
)

type EventPhase uint8

const (
	Preview EventPhase = iota
	Bubble
	Direct
)

// PointerButton identifies the button that generated a pointer event.
type PointerButton uint8

const (
	NoPointerButton PointerButton = iota
	PrimaryButton
	SecondaryButton
)

// InputOptions describes how a node participates in input.
// Negative TabIndex excludes a node from keyboard navigation.
type InputOptions struct {
	Disabled, HitTestInvisible, Focusable bool
	TabIndex                              int
}

// PreviewInputHandler handles events during the root-to-target pass.
type PreviewInputHandler interface{ PreviewInput(*InputEvent) }

// Implement the handlers for the input events a node accepts.
type (
	PointerDownHandler        interface{ PointerDown(*InputEvent) }
	PointerMoveHandler        interface{ PointerMove(*InputEvent) }
	PointerUpHandler          interface{ PointerUp(*InputEvent) }
	PointerWheelHandler       interface{ PointerWheel(*InputEvent) }
	PointerCaptureLostHandler interface{ PointerCaptureLost(*InputEvent) }
	KeyDownHandler            interface{ KeyDown(*InputEvent) }
	KeyUpHandler              interface{ KeyUp(*InputEvent) }
	PointerEnterHandler       interface{ PointerEnter(*InputEvent) }
	PointerLeaveHandler       interface{ PointerLeave(*InputEvent) }
	Tappable                  interface{ Tapped(*InputEvent) }
	SecondaryTappable         interface{ SecondaryTapped(*InputEvent) }
	KeyRepeatHandler          interface{ KeyRepeat(*InputEvent) }
	TextInputHandler          interface{ TextInput(*InputEvent) }
	CompositionHandler        interface{ CompositionChanged(*InputEvent) }
)

// Focusable opts a node into focus navigation and receives focus notifications.
type Focusable interface {
	FocusGained(*InputEvent)
	FocusLost(*InputEvent)
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
	Button                  PointerButton
	Key                     int32
	Text                    string
	Composition             Composition
	Shift, Ctrl, Alt        bool
	Handled                 bool

	router  *inputRouter
	current *component
}

// CapturePointer routes subsequent mouse movement and release to CurrentTarget.
func (e *InputEvent) CapturePointer() bool {
	if e.router == nil || !e.router.down || e.current == nil || e.current.pending == nil || !e.current.inputEnabled() ||
		(e.Kind != PointerMove && (e.Kind != PointerDown || e.Button != PrimaryButton)) {
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

// Focus moves focus to a mounted, focusable node.
func (e *InputEvent) Focus(node Node) bool {
	return e.router != nil && e.router.focus != nil && e.router.focus.focusNode(node, e.router.viewport)
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
	if c.popup {
		return nil
	}
	options := c.layoutBase().Input
	if options.HitTestInvisible {
		return nil
	}
	if !options.Disabled && (!c.measuredBox.ClipToBounds || containsPoint(c.visualRect(c.contentBounds()), point)) {
		for i := len(c.children) - 1; i >= 0; i-- {
			if hit := hitNode(c.children[i], point); hit != nil {
				return hit
			}
		}
	}
	if containsPoint(c.visualRect(c.bounds), point) {
		return c
	}
	return nil
}

func containsPoint(bounds rl.Rectangle, point rl.Vector2) bool {
	return bounds.Width > 0 && bounds.Height > 0 && point.X >= bounds.X && point.Y >= bounds.Y &&
		point.X < bounds.X+bounds.Width && point.Y < bounds.Y+bounds.Height
}

func visibleBounds(c *component) rl.Rectangle {
	bounds := c.visualRect(c.bounds)
	for node := c.parent; node != nil; node = node.parent {
		if node.popup {
			break
		}
		if node.measuredBox.ClipToBounds {
			bounds = intersectRect(bounds, node.visualRect(node.contentBounds()))
		}
	}
	return bounds
}

func intersectRect(a, b rl.Rectangle) rl.Rectangle {
	x, y := max(a.X, b.X), max(a.Y, b.Y)
	return rl.Rectangle{X: x, Y: y, Width: max(0, min(a.X+a.Width, b.X+b.Width)-x), Height: max(0, min(a.Y+a.Height, b.Y+b.Height)-y)}
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
			event.LocalPosition = node.localPoint(event.Position)
			dispatchInput(event.CurrentTarget, event)
		}
	}
	event.current, event.CurrentTarget = nil, nil
}

func directPointerEvent(node *component, kind EventKind, point rl.Vector2) {
	target := node.eventNode()
	dispatchInput(target, &InputEvent{
		Kind: kind, Phase: Direct, Source: target, CurrentTarget: target, Position: point,
		LocalPosition: node.localPoint(point),
	})
}

func dispatchInput(node Node, event *InputEvent) {
	if event.Phase == Preview {
		if handler, ok := node.(PreviewInputHandler); ok {
			handler.PreviewInput(event)
		}
		return
	}
	switch event.Kind {
	case PointerDown:
		if handler, ok := node.(PointerDownHandler); ok {
			handler.PointerDown(event)
		}
	case PointerMove:
		if handler, ok := node.(PointerMoveHandler); ok {
			handler.PointerMove(event)
		}
	case PointerUp:
		if handler, ok := node.(PointerUpHandler); ok {
			handler.PointerUp(event)
		}
	case PointerWheel:
		if handler, ok := node.(PointerWheelHandler); ok {
			handler.PointerWheel(event)
		}
	case PointerCaptureLost:
		if handler, ok := node.(PointerCaptureLostHandler); ok {
			handler.PointerCaptureLost(event)
		}
	case KeyDown:
		if handler, ok := node.(KeyDownHandler); ok {
			handler.KeyDown(event)
		}
	case KeyUp:
		if handler, ok := node.(KeyUpHandler); ok {
			handler.KeyUp(event)
		}
	case FocusGained:
		if handler, ok := node.(Focusable); ok {
			handler.FocusGained(event)
		}
	case FocusLost:
		if handler, ok := node.(Focusable); ok {
			handler.FocusLost(event)
		}
	case PointerEnter:
		if handler, ok := node.(PointerEnterHandler); ok {
			handler.PointerEnter(event)
		}
	case PointerLeave:
		if handler, ok := node.(PointerLeaveHandler); ok {
			handler.PointerLeave(event)
		}
	case Tap:
		if handler, ok := node.(Tappable); ok {
			handler.Tapped(event)
		}
	case SecondaryTap:
		if handler, ok := node.(SecondaryTappable); ok {
			handler.SecondaryTapped(event)
		}
	case KeyRepeat:
		if handler, ok := node.(KeyRepeatHandler); ok {
			handler.KeyRepeat(event)
		}
	case TextInput:
		if handler, ok := node.(TextInputHandler); ok {
			handler.TextInput(event)
		}
	case CompositionChanged:
		if handler, ok := node.(CompositionHandler); ok {
			handler.CompositionChanged(event)
		}
	}
}

type inputRouter struct {
	focus                 *focusState
	viewport              rl.Vector2
	capture               *component
	hoverPath             []*component
	press, secondaryPress *component
	point                 rl.Vector2
	down                  bool
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
	r.press = nil
	r.secondaryPress = nil
	r.setCapture(nil)
}

func (r *inputRouter) prune() {
	if r.capture != nil && (r.capture.pending == nil || !r.capture.inputEnabled()) {
		r.setCapture(nil)
	}
	if r.press != nil && (r.press.pending == nil || !r.press.inputEnabled()) {
		r.press = nil
	}
	if r.secondaryPress != nil && (r.secondaryPress.pending == nil || !r.secondaryPress.inputEnabled()) {
		r.secondaryPress = nil
	}
}

func (r *inputRouter) updateHover(hit *component, point rl.Vector2) {
	if !hit.inputEnabled() || hit.pending == nil {
		hit = nil
	}
	if len(r.hoverPath) == 0 && hit == nil || len(r.hoverPath) > 0 && r.hoverPath[len(r.hoverPath)-1] == hit {
		return
	}
	var path []*component
	for node := hit; node != nil; node = node.parent {
		path = append(path, node)
	}
	slices.Reverse(path)
	common := 0
	for common < len(path) && common < len(r.hoverPath) && path[common] == r.hoverPath[common] {
		common++
	}
	old := r.hoverPath
	r.hoverPath = path
	for i := len(old) - 1; i >= common; i-- {
		directPointerEvent(old[i], PointerLeave, point)
	}
	for _, node := range path[common:] {
		if node.pending != nil && node.inputEnabled() {
			directPointerEvent(node, PointerEnter, point)
		}
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
	button := NoPointerButton
	if kind == PointerDown || kind == PointerUp || (kind == PointerMove && r.down) {
		button = PrimaryButton
	}
	event := &InputEvent{Kind: kind, Position: point, Wheel: wheel, Button: button, router: r}
	if target.inputEnabled() && target.pending != nil {
		routeEvent(target, event)
	}
	if kind == PointerDown {
		r.press = nil
		if !event.Handled && hit.inputEnabled() && hit.pending != nil {
			r.press = hit
		}
	}
	if kind == PointerUp {
		pressed := r.press
		r.press = nil
		r.down = false
		r.setCapture(nil)
		if !event.Handled && pressed == hit && hit.inputEnabled() && hit.pending != nil {
			tap := &InputEvent{Kind: Tap, Position: point, Button: PrimaryButton, router: r}
			routeEvent(hit, tap)
			return tap.Handled
		}
	}
	return event.Handled
}

func (r *inputRouter) secondary(kind EventKind, hit *component, point rl.Vector2) {
	event := &InputEvent{Kind: kind, Position: point, Button: SecondaryButton}
	if hit.inputEnabled() && hit.pending != nil {
		routeEvent(hit, event)
	}
	if kind == PointerDown {
		r.secondaryPress = nil
		if !event.Handled && hit.inputEnabled() && hit.pending != nil {
			r.secondaryPress = hit
		}
		return
	}
	pressed := r.secondaryPress
	r.secondaryPress = nil
	if !event.Handled && pressed == hit && hit.inputEnabled() && hit.pending != nil {
		routeEvent(hit, &InputEvent{Kind: SecondaryTap, Position: point, Button: SecondaryButton})
	}
}

func hitWindow(root *component, mouse, viewport rl.Vector2) *component {
	if mouse.X < 0 || mouse.Y < 0 || mouse.X >= viewport.X || mouse.Y >= viewport.Y {
		return nil
	}
	layers := popupLayers(root, nil)
	for i := len(layers) - 1; i >= 0; i-- {
		if layers[i].inputEnabled() {
			if hit := hitNode(layers[i].children[0], mouse); hit != nil {
				return hit
			}
		}
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
