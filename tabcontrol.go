package rgui

// TabItem describes a tab header and the node shown when it is selected.
type TabItem struct {
	Header  string
	Content Node
}

// TabControl shows one tab's content at a time. The index signal may be
// changed by the application or by activating a header.
func TabControl(items []TabItem, selected *Signal[int]) Node {
	if selected == nil {
		selected = NewSignal(0)
	}
	headers := make([]Node, len(items))
	contents := make([]Node, len(items))
	for i, item := range items {
		header := componentOf(Button(item.Header, func() { selected.Set(i) }))
		header.readText = func() string {
			if selected.Get() == i {
				return "> " + item.Header
			}
			return item.Header
		}
		headers[i], contents[i] = header, item.Content
	}
	return DockPanel(true,
		Docked{Edge: DockTop, Child: WrapPanel(Horizontal, 4, headers...)},
		Docked{Child: newContentSwitch(selected.Get, contents)},
	)
}
