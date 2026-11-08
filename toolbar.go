package rgui

// ToolBar groups commands, wrapping to another line when space runs out.
func ToolBar(commands ...Node) Node {
	return styleName(Border(WrapPanel(Horizontal, 4, commands...), BorderStyle{Padding: UniformInsets(4)}), "toolbar")
}
