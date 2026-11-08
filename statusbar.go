package rgui

// StatusBar displays status content in a horizontal strip. Dock it at the
// bottom of the window to keep it visible while the main area resizes.
func StatusBar(items ...Node) Node {
	return styleName(Border(StackPanel(Horizontal, 8, items...), BorderStyle{
		Padding: UniformInsets(5), ClipToBounds: true,
	}), "statusbar")
}
