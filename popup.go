package rgui

import (
	"math"

	"rgui/draw"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Popup shows content in the window's overlay layer at its layout origin.
// It takes no space along the stacking axis. Set open to false to dismiss it.
func Popup(content Node, open *Signal[bool]) Node {
	if open == nil {
		open = NewSignal(false)
	}
	node := componentOf(newContentSwitch(func() int {
		if open.Get() {
			return 0
		}
		return -1
	}, []Node{content}))
	node.popup = true
	node.panel = popupLayout{owner: node}
	return node
}

type popupLayout struct{ owner *component }

func (p popupLayout) measure(children []*component, available rl.Vector2, text measureText) rl.Vector2 {
	if len(children) != 0 {
		children[0].measure(rl.Vector2{X: available.X, Y: float32(math.Inf(1))}, text)
	}
	return rl.Vector2{}
}

func (p popupLayout) arrange(children []*component, area rl.Rectangle) {
	if len(children) == 0 {
		return
	}
	root := p.owner
	for root.parent != nil {
		root = root.parent
	}
	viewport := root.bounds
	child := children[0]
	width := min(viewport.Width, max(area.Width, child.desired.X))
	height := min(viewport.Height, child.desired.Y)
	x := max(viewport.X, min(area.X, viewport.X+viewport.Width-width))
	y := area.Y
	if y+height > viewport.Y+viewport.Height {
		y = max(viewport.Y, min(y-height, viewport.Y+viewport.Height-height))
	}
	child.arrange(rl.Rectangle{X: x, Y: y, Width: width, Height: height})
}

func popupLayers(node *component, layers []*component) []*component {
	if node.popup && len(node.children) != 0 {
		layers = append(layers, node)
	}
	for _, child := range node.children {
		layers = popupLayers(child, layers)
	}
	return layers
}

func (c *component) paintScene(canvas *draw.Canvas, font rl.Font, hovered, pressed, focused *component, damage *rl.Rectangle) {
	c.paint(canvas, font, hovered, pressed, focused, damage)
	for _, layer := range popupLayers(c, nil) {
		for _, child := range layer.children {
			child.paint(canvas, font, hovered, pressed, focused, damage)
		}
	}
}
