package rgui

import rl "github.com/gen2brain/raylib-go/raylib"

// StatusBar displays status content in a horizontal strip. Dock it at the
// bottom of the window to keep it visible while the main area resizes.
func StatusBar(items ...Node) Node {
	return Border(StackPanel(Horizontal, 8, items...), BorderStyle{
		Background: rl.NewColor(34, 41, 54, 255), Padding: UniformInsets(5), ClipToBounds: true,
	})
}
