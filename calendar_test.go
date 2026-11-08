package rgui

import (
	"testing"
	"time"
)

func TestCalendarLeapDayAndMonthNavigation(t *testing.T) {
	initial := time.Date(2024, time.February, 29, 12, 0, 0, 0, time.UTC)
	month, selected := NewSignal(initial), NewSignal(time.Time{})
	root := componentOf(Calendar(selected, month))
	grid := root.children[1]
	grid.children[7+32].onClick()
	if !sameDate(initial, selected.Get()) {
		t.Fatalf("selection = %v", selected.Get())
	}
	root.children[0].children[2].onClick()
	if month.Get().Month() != time.March || month.Get().Day() != 1 {
		t.Fatal("expected March 1")
	}
}
