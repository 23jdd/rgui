package rgui

import "time"

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
	content := styleName(Border(calendar(selected, month, func() { open.Set(false) }),
		BorderStyle{Thickness: 1, Padding: UniformInsets(4)}), "calendarsurface")
	return styleName(StackPanel(Vertical, 0, button,
		transientPopup(WithLayout(content, Box{MinWidth: 252, Padding: UniformInsets(4)}), open, button, nil)), "datepicker")
}
