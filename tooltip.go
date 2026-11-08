package rgui

// ToolTip displays a noninteractive label below content while the pointer
// remains over it.
func ToolTip(content Node, message string) Node {
	open := NewSignal(false)
	host := &toolTip{open: open}
	tip := styleName(Border(TextBlock(func() string { return message }, Start),
		BorderStyle{Thickness: 1, Padding: UniformInsets(6)}), "tooltipsurface")
	WithInput(tip, InputOptions{HitTestInvisible: true})
	root := panel(stackLayout{direction: Vertical}, []Node{content, Popup(tip, open)})
	root.custom = host
	host.host = root
	return root
}

type toolTip struct {
	BaseNode
	open *Signal[bool]
}

func (t *toolTip) PointerEnter(*InputEvent) { t.open.Set(true) }
func (t *toolTip) PointerLeave(*InputEvent) { t.open.Set(false) }
