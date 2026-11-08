package rgui

// RadioButton selects option in a shared signal. Use one signal across the
// buttons in a group and assign a distinct option to each button.
func RadioButton(label, option string, selected *Signal[string]) Node {
	if selected == nil {
		selected = NewSignal("")
	}
	node := &choice{
		label: label, radio: true, checked: func() bool { return selected.Get() == option },
		activate: func() {
			if selected.Get() != option {
				selected.Set(option)
			}
		},
	}
	node.subscribe = func(pending subscriptions) *binding { return bindVisual(&node.BaseNode, selected, pending) }
	return node
}
