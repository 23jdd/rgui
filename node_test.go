package rgui

import (
	"slices"
	"testing"

	"rgui/draw"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type testPrimitiveNode struct {
	progress *Signal[float32]
	bounds   rl.Rectangle
	calls    []string
}

func (n *testPrimitiveNode) Measure(ctx MeasureContext) rl.Vector2 {
	n.calls = append(n.calls, "measure")
	return rl.Vector2{X: min(80, ctx.Available.X), Y: 20}
}

func (n *testPrimitiveNode) Place(bounds rl.Rectangle) {
	n.calls = append(n.calls, "place")
	n.bounds = bounds
}

func (n *testPrimitiveNode) Primitives(_ PaintContext) []draw.Primitive {
	n.calls = append(n.calls, "primitives")
	fill := n.bounds
	fill.Width *= n.progress.Get()
	return []draw.Primitive{draw.RectPrimitive{Bounds: n.bounds}, draw.RectPrimitive{Bounds: fill}}
}

func TestCustomNodeUsesArrangedContentAndCurrentState(t *testing.T) {
	progress := NewSignal(float32(0.25))
	node := &testPrimitiveNode{progress: progress}
	child := WithLayout(node, Box{Padding: UniformInsets(2), HorizontalAlignment: Center})
	root := StackPanel(Vertical, 0, child)
	testLayout(root, 100, 50)
	checkBounds(t, child, 8, 0, 84, 24)
	ctx := PaintContext{}
	primitives := child.Primitives(ctx)
	if !slices.Equal(node.calls, []string{"measure", "place", "primitives"}) {
		t.Fatalf("node calls = %v", node.calls)
	}
	background := primitives[0].(draw.RectPrimitive)
	fill := primitives[1].(draw.RectPrimitive)
	if background.Bounds != (rl.Rectangle{X: 10, Y: 2, Width: 80, Height: 20}) || fill.Bounds.Width != 20 {
		t.Fatalf("unexpected custom geometry: background=%+v fill=%+v", background.Bounds, fill.Bounds)
	}
	progress.Set(0.75)
	if width := child.Primitives(ctx)[1].(draw.RectPrimitive).Bounds.Width; width != 60 {
		t.Fatalf("custom paint should read current state, got width %v", width)
	}
	testLayout(root, 60, 50)
	if bounds := child.Primitives(ctx)[0].(draw.RectPrimitive).Bounds; bounds.Width != 56 {
		t.Fatalf("custom paint must use the new placement, got %+v", bounds)
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
