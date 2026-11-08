package rgui

import (
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type focusState struct {
	button *component
	key    int32
	keys   []int32
}

func (f *focusState) set(button *component) {
	if f.button != button {
		previous := f.button
		f.button = button
		f.key = 0
		routeEvent(previous, &InputEvent{Kind: FocusLost})
		if f.button == button && button.inputEnabled() {
			routeEvent(button, &InputEvent{Kind: FocusGained})
		}
	}
}

func (f *focusState) move(root *component, viewport rl.Vector2, backward bool) {
	buttons := tabOrder(root, viewport, nil)
	slices.SortStableFunc(buttons, func(a, b *component) int {
		ai, bi := a.layoutBase().Input.TabIndex, b.layoutBase().Input.TabIndex
		if ai < bi {
			return -1
		}
		if ai > bi {
			return 1
		}
		return 0
	})
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
	if node.focusable(viewport) && node.layoutBase().Input.TabIndex >= 0 {
		buttons = append(buttons, node)
	}
	for _, child := range node.children {
		buttons = tabOrder(child, viewport, buttons)
	}
	return buttons
}

func (c *component) focusable(viewport rl.Vector2) bool {
	if !c.inputEnabled() {
		return false
	}
	_, acceptsFocus := c.eventNode().(Focusable)
	if !acceptsFocus && !c.enabledButton() && !c.layoutBase().Input.Focusable {
		return false
	}
	b := visibleBounds(c)
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
	shift := rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift)
	dispatch := func(kind EventKind, key int32) bool {
		event := &InputEvent{Kind: kind, Key: key, Shift: shift,
			Ctrl: rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl),
			Alt:  rl.IsKeyDown(rl.KeyLeftAlt) || rl.IsKeyDown(rl.KeyRightAlt)}
		if f.button.inputEnabled() {
			routeEvent(f.button, event)
		}
		return event.Handled
	}
	release := func(key int32) {
		if dispatch(KeyUp, key) {
			if f.key == key {
				f.cancelPress()
			}
		} else {
			f.activate(key, false, true)
		}
	}
	f.keys = slices.DeleteFunc(f.keys, func(key int32) bool {
		if rl.IsKeyDown(key) {
			return false
		}
		release(key)
		return true
	})
	// The press queue also includes taps that begin and end between frames.
	for key := rl.GetKeyPressed(); key != 0; key = rl.GetKeyPressed() {
		if key == rl.KeyLeftShift || key == rl.KeyRightShift {
			shift = true
		}
		handled := dispatch(KeyDown, key)
		if !handled {
			switch key {
			case rl.KeyTab:
				f.cancelPress()
				f.move(root, viewport, shift)
			case rl.KeySpace, rl.KeyEnter, rl.KeyKpEnter:
				f.activate(key, true, false)
			}
		}
		if rl.IsKeyDown(key) {
			if !slices.Contains(f.keys, key) {
				f.keys = append(f.keys, key)
			}
		} else {
			release(key)
		}
	}
}
