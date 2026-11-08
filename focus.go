package rgui

import (
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type focusState struct {
	button *component
	key    int32
}

func (f *focusState) set(button *component) {
	if f.button != button {
		f.button = button
		f.key = 0
	}
}

func (f *focusState) move(root *component, viewport rl.Vector2, backward bool) {
	buttons := tabOrder(root, viewport, nil)
	if len(buttons) == 0 {
		f.set(nil)
		return
	}
	index := slices.Index(buttons, f.button)
	if backward {
		if index <= 0 {
			index = len(buttons)
		}
		index--
	} else {
		index = (index + 1) % len(buttons)
	}
	f.set(buttons[index])
}

func tabOrder(node *component, viewport rl.Vector2, buttons []*component) []*component {
	if node.focusable(viewport) {
		buttons = append(buttons, node)
	}
	for _, child := range node.children {
		buttons = tabOrder(child, viewport, buttons)
	}
	return buttons
}

func (c *component) focusable(viewport rl.Vector2) bool {
	if !c.enabledButton() {
		return false
	}
	b := c.bounds
	return b.Width > 0 && b.Height > 0 &&
		b.X < viewport.X && b.Y < viewport.Y && b.X+b.Width > 0 && b.Y+b.Height > 0
}

func (f *focusState) activate(key int32, down, up bool) {
	if down && f.key == 0 && f.button.enabledButton() {
		f.key = key
	}
	if up && f.key == key {
		f.key = 0
		if f.button.enabledButton() {
			f.button.onClick()
		}
	}
}

func (f *focusState) cancelPress() {
	f.key = 0
}

func (f *focusState) readKeyboard(root *component, viewport rl.Vector2) {
	if f.key != 0 && !rl.IsKeyDown(f.key) {
		f.activate(f.key, false, true)
	}
	shift := rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift)
	// The press queue also includes taps that begin and end between frames.
	for key := rl.GetKeyPressed(); key != 0; key = rl.GetKeyPressed() {
		switch key {
		case rl.KeyLeftShift, rl.KeyRightShift:
			shift = true
		case rl.KeyTab:
			f.cancelPress()
			f.move(root, viewport, shift)
		case rl.KeySpace, rl.KeyEnter, rl.KeyKpEnter:
			f.activate(key, true, !rl.IsKeyDown(key))
		}
	}
}
