package rgui

import (
	"rgui/draw"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// DragCallbacks receives logical pointer movement relative to the initial press.
type DragCallbacks struct {
	Started   func(position rl.Vector2)
	Moved     func(delta rl.Vector2)
	Completed func(delta rl.Vector2, canceled bool)
}

// Thumb is a draggable handle using primary-pointer capture.
func Thumb(callbacks DragCallbacks) Node {
	return &thumb{callbacks: callbacks, desired: rl.Vector2{X: 14, Y: 14}}
}

type thumb struct {
	BaseNode
	callbacks     DragCallbacks
	desired       rl.Vector2
	origin, delta rl.Vector2
	inputScale    rl.Vector2
	dragging      bool
	keyboard      func(*InputEvent)
}

func (t *thumb) Measure(MeasureContext) rl.Vector2 { return t.desired }

func (t *thumb) PointerDown(e *InputEvent) {
	if e.Source != t || e.Button != PrimaryButton || !e.CapturePointer() {
		return
	}
	t.dragging = true
	t.origin, t.delta = e.Position, rl.Vector2{}
	unit := t.host.visualRect(rl.Rectangle{Width: 1, Height: 1})
	t.inputScale = rl.Vector2{X: max(0.000001, unit.Width), Y: max(0.000001, unit.Height)}
	t.InvalidateVisual()
	if t.callbacks.Started != nil {
		t.callbacks.Started(e.Position)
	}
}

func (t *thumb) PointerMove(e *InputEvent) {
	if !t.dragging || e.Source != t {
		return
	}
	delta := rl.Vector2{X: (e.Position.X - t.origin.X) / t.inputScale.X, Y: (e.Position.Y - t.origin.Y) / t.inputScale.Y}
	if delta != t.delta {
		t.delta = delta
		if t.callbacks.Moved != nil {
			t.callbacks.Moved(delta)
		}
	}
}

func (t *thumb) PointerUp(e *InputEvent) {
	if e.Source == t && e.Button == PrimaryButton && t.dragging {
		t.PointerMove(e)
		t.end(false)
	}
}

func (t *thumb) PointerCaptureLost(*InputEvent) { t.end(true) }

func (t *thumb) end(canceled bool) {
	if !t.dragging {
		return
	}
	t.dragging = false
	t.InvalidateVisual()
	if t.callbacks.Completed != nil {
		t.callbacks.Completed(t.delta, canceled)
	}
}

func (t *thumb) KeyDown(e *InputEvent) {
	if e.Source == t && t.keyboard != nil {
		t.keyboard(e)
	}
}

func (t *thumb) KeyRepeat(e *InputEvent) { t.KeyDown(e) }

func (t *thumb) Primitives(PaintContext) []Node {
	fill := rl.Gray
	if t.dragging {
		fill = rl.SkyBlue
	}
	return []Node{&RectNode{Rect: t.Bounds(), Radius: 2, Style: draw.Style{Fill: fill}}}
}
