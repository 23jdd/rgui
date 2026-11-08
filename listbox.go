package rgui

import (
	"math"
	"slices"

	"rgui/draw"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ListBox shows selectable labels. The signal stores the selected index;
// -1 represents no selection.
func ListBox(items []string, selected *Signal[int]) Node {
	if selected == nil {
		selected = NewSignal(-1)
	}
	offset := NewSignal(rl.Vector2{})
	list := &listBox{items: slices.Clone(items), selected: selected, offset: offset}
	return ScrollViewer(list, offset, false, true)
}

const listRowHeight = 34

type listBox struct {
	BaseNode
	items    []string
	selected *Signal[int]
	offset   *Signal[rl.Vector2]
	focused  bool
}

func (n *listBox) Measure(ctx MeasureContext) rl.Vector2 {
	var width float32
	for _, item := range n.items {
		width = max(width, ctx.MeasureText(item).X)
	}
	return rl.Vector2{X: width + 20, Y: float32(max(1, len(n.items))) * listRowHeight}
}

func (n *listBox) bind(pending subscriptions) *binding {
	return bindVisual(&n.BaseNode, n.selected, pending)
}

func (n *listBox) FocusGained(e *InputEvent) {
	if e.Source == n {
		n.focused = true
		n.InvalidateVisual()
	}
}

func (n *listBox) FocusLost(e *InputEvent) {
	if e.Source == n {
		n.focused = false
		n.InvalidateVisual()
	}
}

func (n *listBox) selectIndex(index int) {
	if index >= 0 && index < len(n.items) && n.selected.Get() != index {
		n.selected.Set(index)
	}
}

func (n *listBox) Tapped(e *InputEvent) {
	if e.Source != n {
		return
	}
	index := int(math.Floor(float64((e.Position.Y - n.Bounds().Y) / listRowHeight)))
	if index >= 0 && index < len(n.items) {
		n.selectIndex(index)
		e.Handled = true
	}
}

func (n *listBox) KeyRepeat(e *InputEvent) { n.KeyDown(e) }
func (n *listBox) KeyDown(e *InputEvent) {
	if e.Source != n || len(n.items) == 0 {
		return
	}
	index := n.selected.Get()
	switch e.Key {
	case rl.KeyUp:
		index = max(0, index-1)
	case rl.KeyDown:
		index = max(0, min(len(n.items)-1, index+1))
	case rl.KeyHome:
		index = 0
	case rl.KeyEnd:
		index = len(n.items) - 1
	default:
		return
	}
	index = min(index, len(n.items)-1)
	n.selectIndex(index)
	if n.host != nil && n.host.parent != nil {
		viewport := n.host.parent.contentBounds()
		y := n.Bounds().Y + float32(index)*listRowHeight
		if delta := min(0, y-viewport.Y) + max(0, y+listRowHeight-viewport.Y-viewport.Height); delta != 0 {
			offset := n.offset.Get()
			offset.Y = viewport.Y - n.Bounds().Y + delta
			n.offset.Set(offset)
		}
	}
	e.Handled = true
}

func (n *listBox) Primitives(ctx PaintContext) []Node {
	bounds := n.Bounds()
	visible := bounds
	if n.host != nil {
		visible = visibleBounds(n.host)
	}
	first := max(0, int(math.Floor(float64((visible.Y-bounds.Y)/listRowHeight))))
	end := min(len(n.items), int(math.Ceil(float64((visible.Y+visible.Height-bounds.Y)/listRowHeight))))
	var nodes []Node
	for i := first; i < end; i++ {
		y := bounds.Y + float32(i)*listRowHeight
		if i == n.selected.Get() {
			style := draw.Style{Fill: rl.DarkBlue}
			if n.focused {
				style.Stroke = draw.Stroke{Color: rl.Gold, Width: 1}
			}
			nodes = append(nodes, &RectNode{Rect: rl.Rectangle{X: bounds.X, Y: y, Width: bounds.Width, Height: listRowHeight}, Style: style})
		}
		nodes = append(nodes, &TextNode{Position: rl.Vector2{X: bounds.X + 10, Y: y + (listRowHeight-fontSize)/2}, Text: n.items[i], Style: textStyle(ctx.Font)})
	}
	return nodes
}
