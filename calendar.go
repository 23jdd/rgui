package rgui

import (
	"strconv"
	"time"
)

// Calendar displays one month and writes the chosen day to selected. The
// month signal controls the displayed month. Nil signals use private state.
func Calendar(selected, month *Signal[time.Time]) Node {
	return calendar(selected, month, nil)
}

func calendar(selected, month *Signal[time.Time], onSelect func()) Node {
	if selected == nil {
		selected = NewSignal(time.Time{})
	}
	if month == nil {
		initial := selected.Get()
		if initial.IsZero() {
			initial = time.Now()
		}
		month = NewSignal(monthStart(initial))
	}
	previous := Button("<", func() { month.Set(monthStart(month.Get()).AddDate(0, -1, 0)) })
	next := Button(">", func() { month.Set(monthStart(month.Get()).AddDate(0, 1, 0)) })
	header := Grid([]Track{Fixed(42)}, []Track{Fixed(38), Star(1), Fixed(38)},
		Cell{Child: previous},
		Cell{Column: 1, Child: Text(func() string { return month.Get().Format("January 2006") })},
		Cell{Column: 2, Child: next},
	)
	rows, columns := make([]Track, 7), make([]Track, 7)
	cells := make([]Cell, 0, 49)
	for i, day := range []string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"} {
		rows[i], columns[i] = Fixed(32), Star(1)
		cells = append(cells, Cell{Column: i, Child: Text(func() string { return day })})
	}
	for i := 0; i < 42; i++ {
		day := componentOf(Button("", func() {
			date := calendarDay(month.Get(), i)
			selected.Set(date)
			if onSelect != nil {
				onSelect()
			}
		}))
		day.readText = func() string {
			date := calendarDay(month.Get(), i)
			label := strconv.Itoa(date.Day())
			if sameDate(date, selected.Get()) {
				return "[" + label + "]"
			}
			return label
		}
		cells = append(cells, Cell{Row: 1 + i/7, Column: i % 7, Child: day})
	}
	return StackPanel(Vertical, 2, header, Grid(rows, columns, cells...))
}

func monthStart(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
}

func calendarDay(date time.Time, cell int) time.Time {
	start := monthStart(date)
	return start.AddDate(0, 0, cell-int(start.Weekday()))
}

func sameDate(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}
