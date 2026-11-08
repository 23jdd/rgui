package rgui

import rl "github.com/gen2brain/raylib-go/raylib"

// GroupBox displays a header and decorates a group of content.
func GroupBox(header string, content Node) Node {
	return Border(StackPanel(Vertical, 6,
		TextBlock(func() string { return header }, Start),
		content,
	), BorderStyle{Background: rl.NewColor(34, 41, 54, 255), BorderColor: rl.Gray,
		Thickness: 1, Radius: 6, Padding: UniformInsets(10)})
}
