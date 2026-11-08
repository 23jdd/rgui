package rgui

import rl "github.com/gen2brain/raylib-go/raylib"

// StickyNoteControl shows collapsible text and ink notes. Both notes store
// their content in application-owned signals.
func StickyNoteControl(title string, text *Signal[string], ink *Signal[[]InkStroke], expanded *Signal[bool]) Node {
	note := TabControl([]TabItem{
		{Header: "Text", Content: TextBox(text)},
		{Header: "Ink", Content: InkCanvas(ink, rl.SkyBlue, 2)},
	}, nil)
	return styleName(Expander(title, note, expanded), "stickynotecontrol")
}
