package rgui

import rl "github.com/gen2brain/raylib-go/raylib"

// GridSplitter edits a FixedFrom track. Horizontal resizes a column along
// the X axis and Vertical resizes a row along the Y axis.
func GridSplitter(direction Direction, size *Signal[float32], minimum, maximum float32) Node {
	validateRange(float64(minimum), float64(maximum))
	if size == nil {
		size = NewSignal(minimum)
	}
	set := func(value float32) { size.Set(float32(rangeValue(float64(value), float64(minimum), float64(maximum)))) }
	var initial float32
	handle := Thumb(DragCallbacks{
		Started: func(rl.Vector2) { initial = size.Get() },
		Moved: func(delta rl.Vector2) {
			distance, _ := direction.axes(delta)
			set(initial + distance)
		},
		Completed: func(_ rl.Vector2, canceled bool) {
			if canceled {
				set(initial)
			}
		},
	}).(*thumb)
	handle.desired = direction.size(8, 14)
	handle.Input.Focusable = true
	handle.keyboard = func(e *InputEvent) {
		switch e.Key {
		case rl.KeyLeft:
			if direction != Horizontal {
				return
			}
			set(size.Get() - 8)
		case rl.KeyRight:
			if direction != Horizontal {
				return
			}
			set(size.Get() + 8)
		case rl.KeyUp:
			if direction != Vertical {
				return
			}
			set(size.Get() - 8)
		case rl.KeyDown:
			if direction != Vertical {
				return
			}
			set(size.Get() + 8)
		default:
			return
		}
		e.Handled = true
	}
	return handle
}
