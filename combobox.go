package rgui

import "slices"

// ComboBox displays a selected label and opens a list of choices. The signal
// holds a zero-based index, or -1 for no selection.
func ComboBox(items []string, selected *Signal[int]) Node {
	items = slices.Clone(items)
	if selected == nil {
		selected = NewSignal(-1)
	}
	open := NewSignal(false)
	trigger := componentOf(Button("Select", func() {
		if len(items) > 0 {
			open.Set(!open.Get())
		}
	}))
	trigger.readText = func() string {
		index := selected.Get()
		label := "Select"
		if index >= 0 && index < len(items) {
			label = items[index]
		}
		if open.Get() {
			return label + " ^"
		}
		return label + " v"
	}
	rows := make([]Node, len(items))
	for i, item := range items {
		rows[i] = Button(item, func() { selected.Set(i); open.Set(false) })
	}
	content := styleName(Border(ScrollViewer(StackPanel(Vertical, 2, rows...), nil, false, true),
		BorderStyle{Thickness: 1, Padding: UniformInsets(4)}), "combosurface")
	return styleName(StackPanel(Vertical, 0, trigger, transientPopup(content, open, trigger, nil)), "combobox")
}
