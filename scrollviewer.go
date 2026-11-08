package rgui

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ScrollViewer clips and scrolls a child along the enabled axes. Offsets are
// logical pixels. A nil signal creates private scroll state.
func ScrollViewer(child Node, offset *Signal[rl.Vector2], horizontal, vertical bool) Node {
	if offset == nil {
		offset = NewSignal(rl.Vector2{})
	}
	viewer := &scrollViewer{offset: offset, horizontal: horizontal, vertical: vertical}
	viewer.ClipToBounds = true
	root := panel(viewer, []Node{child})
	root.custom = viewer
	viewer.host = root
	return root
}

type scrollViewer struct {
	BaseNode
	offset               *Signal[rl.Vector2]
	horizontal, vertical bool
	extent, viewport     rl.Vector2
}

func (s *scrollViewer) bind(pending subscriptions) *binding {
	update := &binding{pending: pending, dependencies: make(map[*subscriptions]struct{}),
		evaluate: func() {
			s.offset.Get()
			s.host.invalidatePlace()
		},
	}
	update.evaluateTracked()
	return update
}

func (s *scrollViewer) measure(children []*component, available rl.Vector2, text measureText) rl.Vector2 {
	limit := available
	if s.horizontal {
		limit.X = float32(math.Inf(1))
	}
	if s.vertical {
		limit.Y = float32(math.Inf(1))
	}
	s.extent = children[0].measure(limit, text)
	result := s.extent
	if s.horizontal {
		result.X = min(result.X, available.X, 320)
	}
	if s.vertical {
		result.Y = min(result.Y, available.Y, 180)
	}
	return result
}

func (s *scrollViewer) clamp(value rl.Vector2) rl.Vector2 {
	var x, y float32
	if s.horizontal {
		x = float32(rangeValue(float64(value.X), 0, float64(max(0, s.extent.X-s.viewport.X))))
	}
	if s.vertical {
		y = float32(rangeValue(float64(value.Y), 0, float64(max(0, s.extent.Y-s.viewport.Y))))
	}
	return rl.Vector2{X: x, Y: y}
}

func (s *scrollViewer) arrange(children []*component, area rl.Rectangle) {
	s.viewport = rl.Vector2{X: area.Width, Y: area.Height}
	offset := s.clamp(s.offset.Get())
	if s.horizontal {
		area.X -= offset.X
		area.Width = max(area.Width, s.extent.X)
	}
	if s.vertical {
		area.Y -= offset.Y
		area.Height = max(area.Height, s.extent.Y)
	}
	children[0].arrange(area)
}

func (s *scrollViewer) PointerWheel(e *InputEvent) {
	old := s.clamp(s.offset.Get())
	x, y := e.Wheel.X, e.Wheel.Y
	if s.horizontal && (!s.vertical || e.Shift) {
		x, y = x+y, 0
	}
	next := s.clamp(rl.Vector2{X: old.X - 36*x, Y: old.Y - 36*y})
	if next != old {
		s.offset.Set(next)
		e.Handled = true
	}
}
