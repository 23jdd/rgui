package rgui

import (
	"testing"
	"time"
)

func TestDatePickerClosesAfterSelection(t *testing.T) {
	selected := NewSignal(time.Date(2024, time.February, 29, 0, 0, 0, 0, time.UTC))
	root := componentOf(DatePicker(selected))
	pending := make(subscriptions)
	root.mount(pending)
	defer root.unmount()
	root.children[0].onClick()
	flushBindings(pending)
	layers := popupLayers(root, nil)
	if len(layers) != 1 {
		t.Fatal("calendar did not open")
	}
	layers[0].children[0].children[0].children[1].children[7+4].onClick()
	flushBindings(pending)
	if !sameDate(selected.Get(), time.Date(2024, time.February, 1, 0, 0, 0, 0, time.UTC)) || len(popupLayers(root, nil)) != 0 {
		t.Fatal("selection did not update the date and close the calendar")
	}
}
