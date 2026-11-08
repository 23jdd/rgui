package rgui

import rl "github.com/gen2brain/raylib-go/raylib"

// MenuItem is an action in a menu.
type MenuItem struct {
	Label    string
	Action   func()
	Disabled bool
}

// MenuGroup is a menu heading and its drop-down actions.
type MenuGroup struct {
	Label string
	Items []MenuItem
}

// Menu creates a horizontal menu bar with one open group at a time.
func Menu(groups ...MenuGroup) Node {
	menus := make([]Node, len(groups))
	open := make([]*Signal[bool], len(groups))
	for i := range open {
		open[i] = NewSignal(false)
	}
	for i, group := range groups {
		trigger := Button(group.Label, func() {
			show := !open[i].Get()
			for j, other := range open {
				other.Set(i == j && show)
			}
		})
		content := menuContent(group.Items, func() { open[i].Set(false) })
		popup := transientPopup(content, open[i], trigger, nil)
		menus[i] = StackPanel(Vertical, 0, trigger, popup)
	}
	return StackPanel(Horizontal, 2, menus...)
}

func menuContent(items []MenuItem, dismiss func()) Node {
	children := make([]Node, 0, len(items))
	for _, item := range items {
		button := Button(item.Label, func() {
			dismiss()
			if item.Action != nil {
				item.Action()
			}
		})
		children = append(children, WithDisabled(button, item.Disabled))
	}
	return Border(ScrollViewer(StackPanel(Vertical, 2, children...), nil, false, true), BorderStyle{
		Background: rl.NewColor(34, 41, 54, 255), BorderColor: rl.Gray, Thickness: 1, Padding: UniformInsets(4),
	})
}
