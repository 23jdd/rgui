package rgui

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// VirtualizingStackPanel scrolls a fixed-height vertical list. It creates
// nodes for visible rows and one overscan row on each side. Put persistent
// row state outside the factory so it survives scrolling out of view.
func VirtualizingStackPanel(count int, rowHeight float32, offset *Signal[float32], item func(int) Node) Node {
	if count < 0 || !(rowHeight > 0) || math.IsInf(float64(rowHeight)*float64(count), 0) || item == nil {
		panic("rgui: invalid virtualizing stack panel")
	}
	if offset == nil {
		offset = NewSignal(float32(0))
	}
	list := &virtualizingStack{count: count, rowHeight: rowHeight, offset: offset, viewport: NewSignal(float32(180)), item: item}
	list.ClipToBounds = true
	root := componentOf(WithChildren(list))
	root.panel = list
	return root
}

type virtualizingStack struct {
	BaseNode
	count            int
	rowHeight        float32
	offset, viewport *Signal[float32]
	item             func(int) Node
	first            int
	mounted          []Node
}

func (s *virtualizingStack) extent() float32 { return float32(s.count) * s.rowHeight }

func (s *virtualizingStack) clamp(offset, height float32) float32 {
	return float32(rangeValue(float64(offset), 0, float64(max(0, s.extent()-height))))
}

func (s *virtualizingStack) bind(pending subscriptions) *binding {
	update := &binding{pending: pending, dependencies: make(map[*subscriptions]struct{}), evaluate: func() {
		height := s.viewport.Get()
		offset := s.clamp(s.offset.Get(), height)
		first := max(0, int(offset/s.rowHeight)-1)
		end := min(s.count, int(math.Ceil(float64((offset+height)/s.rowHeight)))+1)
		if first != s.first || end-first != len(s.mounted) {
			rows := make([]Node, 0, max(0, end-first))
			for i := first; i < end; i++ {
				if prior := i - s.first; prior >= 0 && prior < len(s.mounted) {
					rows = append(rows, s.mounted[prior])
				} else {
					rows = append(rows, s.item(i))
				}
			}
			if err := SetChildren(s.host, rows...); err != nil {
				panic(err)
			}
			s.first, s.mounted = first, rows
		}
		s.host.invalidatePlace()
	}}
	update.evaluateTracked()
	return update
}

func (s *virtualizingStack) measure(children []*component, available rl.Vector2, text measureText) rl.Vector2 {
	var width float32
	for _, child := range children {
		size := child.measure(rl.Vector2{X: available.X, Y: s.rowHeight}, text)
		width = max(width, size.X)
	}
	return rl.Vector2{X: width, Y: min(available.Y, s.extent(), 180)}
}

func (s *virtualizingStack) arrange(children []*component, area rl.Rectangle) {
	if s.viewport.Get() != area.Height {
		s.viewport.Set(area.Height)
	}
	offset := s.clamp(s.offset.Get(), area.Height)
	for i, child := range children {
		child.arrange(Vertical.slot(area, float32(s.first+i)*s.rowHeight-offset, s.rowHeight))
	}
}

func (s *virtualizingStack) PointerWheel(e *InputEvent) {
	height := s.viewport.Get()
	old := s.clamp(s.offset.Get(), height)
	next := s.clamp(old-e.Wheel.Y*s.rowHeight*3, height)
	if old != next {
		s.offset.Set(next)
		e.Handled = true
	}
}
