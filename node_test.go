package rgui

import (
	"slices"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type testPrimitiveNode struct {
	BaseNode
	progress *Signal[float32]
	calls    []string
}

func (n *testPrimitiveNode) Measure(ctx MeasureContext) rl.Vector2 {
	n.calls = append(n.calls, "measure")
	return rl.Vector2{X: min(80, ctx.Available.X), Y: 20}
}

func (n *testPrimitiveNode) Place(bounds rl.Rectangle) {
	n.calls = append(n.calls, "place")
	n.BaseNode.Place(bounds)
}

func (n *testPrimitiveNode) Primitives(_ PaintContext) []Node {
	n.calls = append(n.calls, "primitives")
	bounds := n.Bounds()
	fill := bounds
	fill.Width *= n.progress.Get()
	return []Node{&RectNode{Rect: bounds}, &RectNode{Rect: fill}}
}

func TestBaseNodeDefaults(t *testing.T) {
	var base BaseNode
	var node Node = &base
	if node.Measure(MeasureContext{}) != (rl.Vector2{}) || len(node.Primitives(PaintContext{})) != 0 {
		t.Fatal("base node should have no preferred size or primitives")
	}
	bounds := rl.Rectangle{X: 5, Y: 10, Width: 30, Height: 20}
	node.Place(bounds)
	if base.Bounds() != bounds {
		t.Fatalf("bounds = %+v, want %+v", base.Bounds(), bounds)
	}
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
	background := primitives[0].(*RectNode)
	fill := primitives[1].(*RectNode)
	if background.Rect != (rl.Rectangle{X: 10, Y: 2, Width: 80, Height: 20}) || fill.Rect.Width != 20 {
		t.Fatalf("unexpected custom geometry: background=%+v fill=%+v", background.Rect, fill.Rect)
	}
	progress.Set(0.75)
	if width := child.Primitives(ctx)[1].(*RectNode).Rect.Width; width != 60 {
		t.Fatalf("custom paint should read current state, got width %v", width)
	}
	testLayout(root, 60, 50)
	if bounds := child.Primitives(ctx)[0].(*RectNode).Rect; bounds.Width != 56 {
		t.Fatalf("custom paint must use the new placement, got %+v", bounds)
	}
}

type paintGroup struct {
	BaseNode
	children []Node
}

func (g *paintGroup) Primitives(PaintContext) []Node { return g.children }

func TestRenderPreservesNestedNodeOrder(t *testing.T) {
	var order []int
	renderNode(&paintGroup{children: []Node{
		DrawNode(func(PaintContext) { order = append(order, 1) }),
		nil,
		&paintGroup{children: []Node{
			DrawNode(func(PaintContext) { order = append(order, 2) }),
		}},
		DrawNode(func(PaintContext) { order = append(order, 3) }),
	}}, PaintContext{})
	if !slices.Equal(order, []int{1, 2, 3}) {
		t.Fatalf("paint order = %v", order)
	}
}

func TestRectNodeCanParticipateInLayout(t *testing.T) {
	rect := &RectNode{Rect: rl.Rectangle{Width: 30, Height: 20}}
	root := StackPanel(Vertical, 0, rect)
	testLayout(root, 100, 100)
	if rect.DesiredSize() != (rl.Vector2{X: 30, Y: 20}) ||
		rect.Bounds() != (rl.Rectangle{Width: 100, Height: 20}) {
		t.Fatalf("desired=%+v bounds=%+v", rect.DesiredSize(), rect.Bounds())
	}
}
