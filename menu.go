package rgui

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
	return styleName(StackPanel(Horizontal, 2, menus...), "menu")
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
	return styleName(Border(ScrollViewer(StackPanel(Vertical, 2, children...), nil, false, true),
		BorderStyle{Thickness: 1, Padding: UniformInsets(4)}), "menusurface")
}
