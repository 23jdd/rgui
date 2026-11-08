package rgui

import (
	"math"

	"rgui/draw"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Slider edits a signal within a range. A positive step snaps to increments
// from minimum. A zero step allows continuous movement.
func Slider(value *Signal[float64], minimum, maximum, step float64) Node {
	validateRange(minimum, maximum)
	if step < 0 || math.IsNaN(step) || math.IsInf(step, 0) {
		panic("rgui: slider step must be finite and nonnegative")
	}
	if value == nil {
		value = NewSignal(minimum)
	}
	return &slider{value: value, minimum: minimum, maximum: maximum, step: step}
}

type slider struct {
	BaseNode
	value                  *Signal[float64]
	minimum, maximum, step float64
	focused, dragging      bool
}

const sliderThumb = 14

func (*slider) Measure(ctx MeasureContext) rl.Vector2 {
	return rl.Vector2{X: min(200, ctx.Available.X), Y: 34}
}

func (s *slider) bind(pending subscriptions) *binding {
	return bindVisual(&s.BaseNode, s.value, pending)
}

func (s *slider) FocusGained(e *InputEvent) {
	if e.Source == s {
		s.focused = true
		s.InvalidateVisual()
	}
}

func (s *slider) FocusLost(e *InputEvent) {
	if e.Source == s {
		s.focused = false
		s.InvalidateVisual()
	}
}

func (s *slider) set(value float64) {
	value = rangeValue(value, s.minimum, s.maximum)
	if s.step > 0 && value > s.minimum && value < s.maximum {
		value = rangeValue(s.minimum+math.Round((value-s.minimum)/s.step)*s.step, s.minimum, s.maximum)
	}
	if s.value.Get() != value {
		s.value.Set(value)
	}
}

func (s *slider) track() rl.Rectangle {
	bounds := s.Bounds()
	return rl.Rectangle{X: bounds.X + sliderThumb/2, Y: bounds.Y + (bounds.Height-4)/2, Width: max(0, bounds.Width-sliderThumb), Height: 4}
}

func (s *slider) movePointer(x float32) {
	track := s.track()
	if track.Width > 0 {
		fraction := float64(max(0, min(1, (x-track.X)/track.Width)))
		s.set(s.minimum + fraction*(s.maximum-s.minimum))
	}
}

func (s *slider) PointerDown(e *InputEvent) {
	if e.Source == s && e.Button == PrimaryButton {
		s.dragging = e.CapturePointer()
		s.movePointer(s.Bounds().X + s.localPointer(e.Position).X)
	}
}

func (s *slider) PointerMove(e *InputEvent) {
	if s.dragging && e.Source == s {
		s.movePointer(s.Bounds().X + s.localPointer(e.Position).X)
	}
}

func (s *slider) PointerUp(e *InputEvent) {
	if s.dragging && e.Source == s && e.Button == PrimaryButton {
		s.movePointer(s.Bounds().X + s.localPointer(e.Position).X)
		s.dragging = false
	}
}

func (s *slider) PointerCaptureLost(*InputEvent) { s.dragging = false }

func (s *slider) KeyRepeat(e *InputEvent) { s.KeyDown(e) }
func (s *slider) KeyDown(e *InputEvent) {
	if e.Source != s {
		return
	}
	step := s.step
	if step == 0 {
		step = (s.maximum - s.minimum) / 100
	}
	value := rangeValue(s.value.Get(), s.minimum, s.maximum)
	switch e.Key {
	case rl.KeyLeft, rl.KeyDown:
		s.set(value - step)
	case rl.KeyRight, rl.KeyUp:
		s.set(value + step)
	case rl.KeyHome:
		s.set(s.minimum)
	case rl.KeyEnd:
		s.set(s.maximum)
	default:
		return
	}
	e.Handled = true
}

func (s *slider) Primitives(ctx PaintContext) []Node {
	track := s.track()
	fill := track
	fill.Width *= float32((rangeValue(s.value.Get(), s.minimum, s.maximum) - s.minimum) / (s.maximum - s.minimum))
	thumb := rl.Rectangle{X: track.X + fill.Width - sliderThumb/2, Y: track.Y + (track.Height-sliderThumb)/2, Width: sliderThumb, Height: sliderThumb}
	color := ctx.color("--thumb-color", ctx.TextStyle().Color)
	if s.focused {
		color = ctx.color("--focus", rl.Gold)
	}
	radius := ctx.Style.Length("--track-radius", 2)
	return []Node{
		&RectNode{Rect: track, Radius: radius, Style: draw.Style{Fill: ctx.color("--track", rl.DarkGray)}},
		&RectNode{Rect: fill, Radius: radius, Style: draw.Style{Fill: ctx.color("--accent", rl.SkyBlue)}},
		&RectNode{Rect: thumb, Radius: ctx.Style.Length("--thumb-radius", sliderThumb/2), Style: draw.Style{Fill: color}},
	}
}
