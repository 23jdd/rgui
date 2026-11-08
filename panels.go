package rgui

import rl "github.com/gen2brain/raylib-go/raylib"

// DockEdge is the side of the remaining space assigned to a docked child.
type DockEdge uint8

const (
	DockLeft DockEdge = iota
	DockTop
	DockRight
	DockBottom
)

// Docked places Child against Edge in a DockPanel.
type Docked struct {
	Edge  DockEdge
	Child Node
}

// DockPanel consumes space from the chosen edges in declaration order.
// If lastChildFill is true, the last child receives the remaining rectangle.
func DockPanel(lastChildFill bool, children ...Docked) Node {
	dock := dockLayout{lastChildFill: lastChildFill, edges: make([]DockEdge, len(children))}
	nodes := make([]Node, len(children))
	for i, child := range children {
		dock.edges[i], nodes[i] = child.Edge, child.Child
	}
	return panel(dock, nodes)
}

type dockLayout struct {
	lastChildFill bool
	edges         []DockEdge
}

func (d dockLayout) measure(children []*component, available rl.Vector2, text measureText) rl.Vector2 {
	for i, child := range children {
		size := child.measure(available, text)
		if d.edges[i] == DockTop || d.edges[i] == DockBottom {
			available.Y = max(0, available.Y-size.Y)
		} else {
			available.X = max(0, available.X-size.X)
		}
	}
	var desired rl.Vector2
	last := len(children) - 1
	if d.lastChildFill && last >= 0 {
		desired = children[last].desired
		last--
	}
	for i := last; i >= 0; i-- {
		size := children[i].desired
		if d.edges[i] == DockTop || d.edges[i] == DockBottom {
			desired.X = max(desired.X, size.X)
			desired.Y += size.Y
		} else {
			desired.X += size.X
			desired.Y = max(desired.Y, size.Y)
		}
	}
	return desired
}

func (d dockLayout) arrange(children []*component, area rl.Rectangle) {
	for i, child := range children {
		if d.lastChildFill && i == len(children)-1 {
			child.arrange(area)
			return
		}
		slot := area
		switch d.edges[i] {
		case DockTop, DockBottom:
			slot.Height = min(area.Height, child.desired.Y)
			area.Height -= slot.Height
			if d.edges[i] == DockTop {
				area.Y += slot.Height
			} else {
				slot.Y += area.Height
			}
		default:
			slot.Width = min(area.Width, child.desired.X)
			area.Width -= slot.Width
			if d.edges[i] == DockRight {
				slot.X += area.Width
			} else {
				area.X += slot.Width
			}
		}
		child.arrange(slot)
	}
}

// WrapPanel arranges children along direction, starting a new line when full.
// Gap separates both children and lines.
func WrapPanel(direction Direction, gap float32, children ...Node) Node {
	return panel(wrapLayout{direction: direction, gap: max(0, gap)}, children)
}

type wrapLayout struct {
	direction Direction
	gap       float32
}

type wrapLine struct {
	children  []*component
	length    float32
	thickness float32
}

func (w wrapLayout) eachLine(children []*component, limit float32, visit func(wrapLine)) {
	var line wrapLine
	start := 0
	for i, child := range children {
		main, cross := w.direction.axes(child.desired)
		if i > start && line.length+w.gap+main > limit {
			line.children = children[start:i]
			visit(line)
			line = wrapLine{}
			start = i
		}
		if i > start {
			line.length += w.gap
		}
		line.length += main
		line.thickness = max(line.thickness, cross)
	}
	if start < len(children) {
		line.children = children[start:]
		visit(line)
	}
}

func (w wrapLayout) measure(children []*component, available rl.Vector2, text measureText) rl.Vector2 {
	for _, child := range children {
		child.measure(available, text)
	}
	limit, _ := w.direction.axes(available)
	var main, cross float32
	lineCount := 0
	w.eachLine(children, limit, func(line wrapLine) {
		main = max(main, line.length)
		cross += line.thickness
		lineCount++
	})
	cross += w.gap * float32(max(0, lineCount-1))
	return w.direction.size(main, cross)
}

func (w wrapLayout) arrange(children []*component, area rl.Rectangle) {
	limit, _ := w.direction.axes(rl.Vector2{X: area.Width, Y: area.Height})
	var crossOffset float32
	w.eachLine(children, limit, func(line wrapLine) {
		lineArea := area
		if w.direction == Horizontal {
			lineArea.Y += crossOffset
			lineArea.Height = line.thickness
		} else {
			lineArea.X += crossOffset
			lineArea.Width = line.thickness
		}
		var mainOffset float32
		for _, child := range line.children {
			length, _ := w.direction.axes(child.desired)
			child.arrange(w.direction.slot(lineArea, mainOffset, length))
			mainOffset += length + w.gap
		}
		crossOffset += line.thickness + w.gap
	})
}
