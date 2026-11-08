package rgui

import (
	"math"

	"rgui/draw"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ProgressBar displays a value between minimum and maximum. A nil signal
// displays the minimum.
func ProgressBar(value *Signal[float64], minimum, maximum float64) Node {
	validateRange(minimum, maximum)
	if value == nil {
		value = NewSignal(minimum)
	}
	return &progressBar{value: value, minimum: minimum, maximum: maximum}
}

type progressBar struct {
	BaseNode
	value            *Signal[float64]
	minimum, maximum float64
}

func (n *progressBar) Measure(ctx MeasureContext) rl.Vector2 {
	return rl.Vector2{X: min(180, ctx.Available.X), Y: 16}
}

func (n *progressBar) bind(pending subscriptions) *binding {
	return bindVisual(&n.BaseNode, n.value, pending)
}

func (n *progressBar) Primitives(ctx PaintContext) []Node {
	track := n.Bounds()
	fill := track
	fill.Width *= float32((rangeValue(n.value.Get(), n.minimum, n.maximum) - n.minimum) / (n.maximum - n.minimum))
	radius := ctx.Style.Length("--track-radius", 4)
	return []Node{
		&RectNode{Rect: track, Radius: radius, Style: draw.Style{Fill: ctx.color("--track", rl.DarkGray)}},
		&RectNode{Rect: fill, Radius: radius, Style: draw.Style{Fill: ctx.color("--accent", rl.SkyBlue)}},
	}
}

func validateRange(minimum, maximum float64) {
	if !(maximum > minimum) || math.IsInf(maximum-minimum, 0) {
		panic("rgui: range must have finite increasing endpoints")
	}
}

func rangeValue(value, minimum, maximum float64) float64 {
	if math.IsNaN(value) {
		return minimum
	}
	return max(minimum, min(maximum, value))
}

func bindVisual[T any](base *BaseNode, value *Signal[T], pending subscriptions) *binding {
	update := &binding{
		pending: pending, dependencies: make(map[*subscriptions]struct{}),
		evaluate: func() {
			value.Get()
			base.InvalidateVisual()
		},
	}
	update.evaluateTracked()
	return update
}
