package rgui

import (
	"slices"
	"testing"

	"rgui/draw"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type testPrimitiveNode struct {
	progress *Signal[float32]
}

func (n testPrimitiveNode) Measure(available rl.Vector2) rl.Vector2 {
	return rl.Vector2{X: min(80, available.X), Y: 20}
}

func (n testPrimitiveNode) Primitives(ctx PaintContext) []draw.Primitive {
	fill := ctx.Bounds
	fill.Width *= n.progress.Get()
	return []draw.Primitive{draw.RectPrimitive{Bounds: ctx.Bounds}, draw.RectPrimitive{Bounds: fill}}
}

func TestCustomNodeUsesArrangedContentAndCurrentState(t *testing.T) {
	progress := NewSignal(float32(0.25))
	var node Node = testPrimitiveNode{progress: progress}
	child := WithLayout(node, Box{Padding: UniformInsets(2), HorizontalAlignment: Center})
	root := StackPanel(Vertical, 0, child)
	testLayout(root, 100, 50)
	checkBounds(t, child, 8, 0, 84, 24)
	ctx := PaintContext{Bounds: child.component.bounds}
	primitives := child.Primitives(ctx)
	background := primitives[0].(draw.RectPrimitive)
	fill := primitives[1].(draw.RectPrimitive)
	if background.Bounds != (rl.Rectangle{X: 10, Y: 2, Width: 80, Height: 20}) || fill.Bounds.Width != 20 {
		t.Fatalf("unexpected custom geometry: background=%+v fill=%+v", background.Bounds, fill.Bounds)
	}
	progress.Set(0.75)
	if width := child.Primitives(ctx)[1].(draw.RectPrimitive).Bounds.Width; width != 60 {
		t.Fatalf("custom paint should read current state, got width %v", width)
	}
}

func TestRenderPreservesPrimitiveOrder(t *testing.T) {
	var order []int
	draw.Render([]draw.Primitive{
		draw.PrimitiveFunc(func() { order = append(order, 1) }),
		nil,
		draw.PrimitiveFunc(func() { order = append(order, 2) }),
	})
	if !slices.Equal(order, []int{1, 2}) {
		t.Fatalf("paint order = %v", order)
	}
}
