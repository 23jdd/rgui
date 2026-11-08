package rgui

import (
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// DatePicker opens a calendar and displays the selected date. A zero date
// represents no selection.
func DatePicker(selected *Signal[time.Time]) Node {
	if selected == nil {
		selected = NewSignal(time.Time{})
	}
	open, month := NewSignal(false), NewSignal(monthStart(time.Now()))
	button := componentOf(Button("Choose date", func() {
		if !open.Get() {
			if date := selected.Get(); !date.IsZero() {
				month.Set(monthStart(date))
			}
		}
		open.Set(!open.Get())
	}))
	button.readText = func() string {
		if date := selected.Get(); !date.IsZero() {
			return date.Format("2006-01-02") + " v"
		}
		return "Choose date v"
	}
	content := Border(calendar(selected, month, func() { open.Set(false) }), BorderStyle{
		Background: rl.NewColor(34, 41, 54, 255), BorderColor: rl.Gray, Thickness: 1, Padding: UniformInsets(4),
	})
	return StackPanel(Vertical, 0, button, transientPopup(WithLayout(content, Box{MinWidth: 252}), open, button, nil))
}
