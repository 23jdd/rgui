package rgui

import rl "github.com/gen2brain/raylib-go/raylib"

// ContextMenu attaches a right-click menu to a node.
func ContextMenu(content Node, items ...MenuItem) Node {
	host := &contextMenu{open: NewSignal(false)}
	menu := menuContent(items, func() { host.open.Set(false) })
	popup := transientPopup(menu, host.open, nil, func() rl.Vector2 { return host.at })
	return WithChildren(host, content, popup)
}

type contextMenu struct {
	BaseNode
	open *Signal[bool]
	at   rl.Vector2
}

func (m *contextMenu) SecondaryTapped(e *InputEvent) {
	m.at = e.Position
	m.open.Set(true)
	m.InvalidateMeasure()
	e.Handled = true
}
