package rgui

// GroupBox displays a header and decorates a group of content.
func GroupBox(header string, content Node) Node {
	return styleName(Border(StackPanel(Vertical, 6,
		WithStyle(TextBlock(func() string { return header }, Start), StyleOptions{Classes: []string{"group-title"}}),
		content,
	), BorderStyle{Thickness: 1, Padding: UniformInsets(10)}), "groupbox")
}
