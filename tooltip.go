package rgui

import rl "github.com/gen2brain/raylib-go/raylib"

// ToolTip displays a noninteractive label below content while the pointer
// remains over it.
func ToolTip(content Node, message string) Node {
	open := NewSignal(false)
	host := &toolTip{open: open}
	tip := Border(TextBlock(func() string { return message }, Start), BorderStyle{
		Background: rl.NewColor(42, 49, 63, 255), BorderColor: rl.Gray, Thickness: 1, Padding: UniformInsets(6),
	})
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
