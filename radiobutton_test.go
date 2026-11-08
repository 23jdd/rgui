package rgui

import "testing"

func TestRadioButtonSelectionIsExclusive(t *testing.T) {
	selected := NewSignal("a")
	a, b := RadioButton("A", "a", selected).(*choice), RadioButton("B", "b", selected).(*choice)
	b.Tapped(&InputEvent{Source: b})
	b.Tapped(&InputEvent{Source: b})
	if selected.Get() != "b" || a.checked() || !b.checked() {
		t.Fatal("only B should be selected")
	}
}
