package rgui

import (
	"unicode"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Label labels a focusable target. Click it or use Alt with its mnemonic to
// move focus to the target. Use an underscore to mark the mnemonic.
func Label(text string, target Node) Node {
	render := AccessText(text).(*accessText)
	var mnemonic rune
	if render.index >= 0 {
		mnemonic = unicode.ToUpper(render.text[render.index])
	}
	return WithChildren(&labelNode{target: target, mnemonic: mnemonic}, render)
}

type labelNode struct {
	BaseNode
	target   Node
	mnemonic rune
}

func (l *labelNode) Tapped(e *InputEvent) {
	if e.Focus(l.target) {
		e.Handled = true
	}
}

func focusAccessKey(root *component, focus *focusState, viewport rl.Vector2, key rune) bool {
	if root.inputEnabled() {
		if label, ok := root.custom.(*labelNode); ok && label.mnemonic == unicode.ToUpper(key) &&
			nonempty(visibleBounds(root)) && focus.focusNode(label.target, viewport) {
			return true
		}
		for _, child := range root.children {
			if focusAccessKey(child, focus, viewport, key) {
				return true
			}
		}
	}
	return false
}
