package rgui

// TreeItem describes a node in a TreeView. Keys are unique within the tree.
type TreeItem struct {
	Key, Label string
	Children   []TreeItem
}

// TreeView displays expandable branches and updates the selected key.
func TreeView(items []TreeItem, selected *Signal[string]) Node {
	if selected == nil {
		selected = NewSignal("")
	}
	seen := make(map[string]bool)
	var build func([]TreeItem) []Node
	build = func(items []TreeItem) []Node {
		nodes := make([]Node, 0, len(items))
		for _, item := range items {
			if item.Key == "" || seen[item.Key] {
				panic("rgui: TreeView keys must be nonempty and unique")
			}
			seen[item.Key] = true
			label := componentOf(Button(item.Label, func() { selected.Set(item.Key) }))
			label.readText = func() string {
				if selected.Get() == item.Key {
					return "* " + item.Label
				}
				return item.Label
			}
			if len(item.Children) == 0 {
				nodes = append(nodes, label)
				continue
			}
			open := NewSignal(false)
			toggle := componentOf(Button("+", func() { open.Set(!open.Get()) }))
			toggle.readText = func() string {
				if open.Get() {
					return "-"
				}
				return "+"
			}
			header := Grid([]Track{Fixed(40)}, []Track{Fixed(36), Star(1)}, Cell{Child: toggle}, Cell{Column: 1, Child: label})
			children := WithLayout(StackPanel(Vertical, 2, build(item.Children)...), Box{Margin: Insets{Left: 20}})
			body := newContentSwitch(func() int {
				if open.Get() {
					return 0
				}
				return -1
			}, []Node{children})
			nodes = append(nodes, StackPanel(Vertical, 2, header, body))
		}
		return nodes
	}
	return ScrollViewer(StackPanel(Vertical, 2, build(items)...), nil, false, true)
}
