package rgui

import rl "github.com/gen2brain/raylib-go/raylib"

// ToolBar groups commands, wrapping to another line when space runs out.
func ToolBar(commands ...Node) Node {
	return Border(WrapPanel(Horizontal, 4, commands...), BorderStyle{
		Background: rl.NewColor(42, 49, 63, 255), Padding: UniformInsets(4),
	})
}
