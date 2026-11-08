package rgui

import (
	"rgui/draw"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ScrollBar edits a scroll position. Page is the viewport length in the same
// units as the value. Direction selects the movement axis.
func ScrollBar(direction Direction, value *Signal[float64], minimum, maximum, page float64) Node {
	validateRange(minimum, maximum)
	if value == nil {
		value = NewSignal(minimum)
	}
	bar := &scrollBar{direction: direction, value: value, minimum: minimum, maximum: maximum, page: max(0, page)}
	bar.Input.Focusable = true
	var initial float64
	handle := Thumb(DragCallbacks{
		Started: func(rl.Vector2) { initial = rangeValue(value.Get(), minimum, maximum) },
		Moved: func(delta rl.Vector2) {
			length, _ := direction.axes(delta)
			if bar.travel > 0 {
				bar.set(initial + float64(length/bar.travel)*(maximum-minimum))
			}
		},
	})
	root := panel(bar, []Node{handle})
	root.custom = bar
	bar.host = root
	return root
}

type scrollBar struct {
	BaseNode
	direction                         Direction
	value                             *Signal[float64]
	minimum, maximum, page            float64
	travel, handleStart, handleLength float32
}

func (b *scrollBar) bind(pending subscriptions) *binding {
	update := &binding{pending: pending, dependencies: make(map[*subscriptions]struct{}), evaluate: func() {
		b.value.Get()
		b.host.invalidatePlace()
	}}
	update.evaluateTracked()
	return update
}

func (b *scrollBar) measure(children []*component, available rl.Vector2, text measureText) rl.Vector2 {
	children[0].measure(available, text)
	length, _ := b.direction.axes(available)
	return b.direction.size(min(100, length), 16)
}

func (b *scrollBar) arrange(children []*component, area rl.Rectangle) {
	length, _ := b.direction.axes(rl.Vector2{X: area.Width, Y: area.Height})
	b.handleLength = min(length, max(12, length*float32(b.page/(b.page+b.maximum-b.minimum))))
	b.travel = max(0, length-b.handleLength)
	fraction := (rangeValue(b.value.Get(), b.minimum, b.maximum) - b.minimum) / (b.maximum - b.minimum)
	b.handleStart = b.travel * float32(fraction)
	children[0].arrange(b.direction.slot(area, b.handleStart, b.handleLength))
}

func (b *scrollBar) set(value float64) {
	value = rangeValue(value, b.minimum, b.maximum)
	if value != b.value.Get() {
		b.value.Set(value)
	}
}

func (b *scrollBar) Tapped(e *InputEvent) {
	if e.Source != b {
		return
	}
	bounds := b.Bounds()
	position, _ := b.direction.axes(rl.Vector2{X: e.Position.X - bounds.X, Y: e.Position.Y - bounds.Y})
	if position < b.handleStart {
		b.set(b.value.Get() - b.page)
	} else if position > b.handleStart+b.handleLength {
		b.set(b.value.Get() + b.page)
	}
	e.Handled = true
}

func (b *scrollBar) KeyRepeat(e *InputEvent) { b.KeyDown(e) }
func (b *scrollBar) KeyDown(e *InputEvent) {
	if e.Source != b {
		return
	}
	value := b.value.Get()
	step := (b.maximum - b.minimum) / 20
	switch e.Key {
	case rl.KeyHome:
		value = b.minimum
	case rl.KeyEnd:
		value = b.maximum
	case rl.KeyPageUp:
		value -= b.page
	case rl.KeyPageDown:
		value += b.page
	case rl.KeyLeft, rl.KeyUp:
		value -= step
	case rl.KeyRight, rl.KeyDown:
		value += step
	default:
		return
	}
	b.set(value)
	e.Handled = true
}

func (b *scrollBar) Primitives(PaintContext) []Node {
	return []Node{&RectNode{Rect: b.Bounds(), Radius: 3, Style: draw.Style{Fill: rl.DarkGray}}}
}
